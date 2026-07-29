package session_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"hara.sh/alloy/session"
	"hara.sh/alloy/session/handlers"
)

// countingHandler wraps ArrayHandler and counts Write calls so tests can assert
// that a clean, read-only request never touches the backend.
type countingHandler struct {
	*handlers.ArrayHandler
	writes atomic.Int64
}

func newCountingHandler() *countingHandler {
	return &countingHandler{ArrayHandler: handlers.NewArrayHandler()}
}

func (h *countingHandler) Write(ctx context.Context, id, data string) error {
	h.writes.Add(1)

	return h.ArrayHandler.Write(ctx, id, data)
}

// cleanSeeded returns a started store that already exists in the backend and
// holds seed, with its dirty flag reset — the shape of a read-only request that
// loaded an existing session.
func cleanSeeded(t *testing.T, seed map[string]any) *session.Store {
	t.Helper()

	ctx := context.Background()
	h := handlers.NewArrayHandler()

	s1 := session.New("t", h)

	if err := s1.Start(ctx); err != nil {
		t.Fatalf("seed start: %v", err)
	}

	for k, v := range seed {
		s1.Put(k, v)
	}

	if err := s1.Save(ctx); err != nil {
		t.Fatalf("seed save: %v", err)
	}

	s2 := session.NewWithID("t", h, s1.GetID())

	if err := s2.Start(ctx); err != nil {
		t.Fatalf("reload start: %v", err)
	}

	if s2.IsDirty() {
		t.Fatal("a freshly loaded, unmodified session must not be dirty")
	}

	return s2
}

func TestStoreDirtyTrackingCoversAllMutators(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name   string
		seed   map[string]any
		mutate func(s *session.Store)
	}{
		{"Put", nil, func(s *session.Store) { s.Put("a", 1) }},
		{"Pull", map[string]any{"a": 1}, func(s *session.Store) { s.Pull("a", nil) }},
		{"Remove", map[string]any{"a": 1}, func(s *session.Store) { s.Remove("a") }},
		{"PushNew", nil, func(s *session.Store) { s.Push("a", 1) }},
		{"PushExisting", map[string]any{"a": []any{1}}, func(s *session.Store) { s.Push("a", 2) }},
		{"Forget", map[string]any{"a": 1}, func(s *session.Store) { s.Forget("a") }},
		{"Flush", map[string]any{"a": 1}, func(s *session.Store) { s.Flush() }},
		{"Replace", nil, func(s *session.Store) { s.Replace(map[string]any{"a": 1}) }},
		{"Increment", nil, func(s *session.Store) { s.Increment("a", 1) }},
		{"Decrement", nil, func(s *session.Store) { s.Decrement("a", 1) }},
		{"Remember", nil, func(s *session.Store) { s.Remember("a", func() any { return 1 }) }},
		{"Flash", nil, func(s *session.Store) { s.Flash("a", 1) }},
		{"Now", nil, func(s *session.Store) { s.Now("a", 1) }},
		{"FlashInput", nil, func(s *session.Store) { s.FlashInput(map[string]any{"a": 1}) }},
		{"Reflash", nil, func(s *session.Store) { s.Reflash() }},
		{"Keep", nil, func(s *session.Store) { s.Keep("a") }},
		{"Token", nil, func(s *session.Store) { s.Token() }},
		{"RegenerateToken", nil, func(s *session.Store) { s.RegenerateToken() }},
		{"Regenerate", nil, func(s *session.Store) { _ = s.Regenerate(ctx, false) }},
		{"Migrate", nil, func(s *session.Store) { _ = s.Migrate(ctx, false) }},
		{"Invalidate", nil, func(s *session.Store) { _ = s.Invalidate(ctx) }},
		{"SetPreviousURL", nil, func(s *session.Store) { s.SetPreviousURL("/x") }},
		{"SetPreviousRoute", nil, func(s *session.Store) { s.SetPreviousRoute("home") }},
		{"PasswordConfirmed", nil, func(s *session.Store) { s.PasswordConfirmed() }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := cleanSeeded(t, tc.seed)
			tc.mutate(s)

			if !s.IsDirty() {
				t.Errorf("%s must mark the session dirty", tc.name)
			}
		})
	}
}

func TestStoreDirtyReadsStayClean(t *testing.T) {
	cases := []struct {
		name string
		seed map[string]any
		read func(s *session.Store)
	}{
		{"Get", map[string]any{"a": 1}, func(s *session.Store) { s.Get("a", nil) }},
		{"Has", map[string]any{"a": 1}, func(s *session.Store) { s.Has("a") }},
		{"Exists", map[string]any{"a": 1}, func(s *session.Store) { s.Exists("a") }},
		{"Missing", nil, func(s *session.Store) { s.Missing("a") }},
		{"All", map[string]any{"a": 1}, func(s *session.Store) { s.All() }},
		{"Only", map[string]any{"a": 1}, func(s *session.Store) { s.Only("a") }},
		{"Except", map[string]any{"a": 1}, func(s *session.Store) { s.Except("a") }},
		{"HasAny", map[string]any{"a": 1}, func(s *session.Store) { s.HasAny("a") }},
		{"ForgetAbsent", nil, func(s *session.Store) { s.Forget("missing") }},
		{"PullAbsent", nil, func(s *session.Store) { s.Pull("missing", nil) }},
		{"ReplaceEmpty", nil, func(s *session.Store) { s.Replace(map[string]any{}) }},
		{"TokenExisting", map[string]any{"_token": "0123456789abcdef0123456789abcdef01234567"}, func(s *session.Store) { s.Token() }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := cleanSeeded(t, tc.seed)
			tc.read(s)

			if s.IsDirty() {
				t.Errorf("%s must not mark the session dirty", tc.name)
			}
		})
	}
}

func TestStoreSaveSkipsCleanReadOnlyRequest(t *testing.T) {
	ctx := context.Background()
	h := newCountingHandler()

	// First request: a brand-new session persists once.
	s1 := session.New("t", h)

	if err := s1.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}

	s1.Put("user", 42)

	if err := s1.Save(ctx); err != nil {
		t.Fatalf("save: %v", err)
	}

	if got := h.writes.Load(); got != 1 {
		t.Fatalf("expected 1 write for the initial save, got %d", got)
	}

	// Second request: load the existing session and only read it.
	h.writes.Store(0)

	s2 := session.NewWithID("t", h, s1.GetID())

	if err := s2.Start(ctx); err != nil {
		t.Fatalf("reload start: %v", err)
	}

	if got := s2.Get("user", nil); got == nil {
		t.Fatal("expected to read the seeded value")
	}

	if err := s2.Save(ctx); err != nil {
		t.Fatalf("read-only save: %v", err)
	}

	if got := h.writes.Load(); got != 0 {
		t.Errorf("a read-only request must issue zero writes, got %d", got)
	}
}

func TestStoreSaveWritesWhenMutated(t *testing.T) {
	ctx := context.Background()
	h := newCountingHandler()

	s1 := session.New("t", h)
	_ = s1.Start(ctx)
	s1.Put("user", 42)
	_ = s1.Save(ctx)

	h.writes.Store(0)

	s2 := session.NewWithID("t", h, s1.GetID())
	_ = s2.Start(ctx)
	s2.Put("user", 43)

	if err := s2.Save(ctx); err != nil {
		t.Fatalf("save: %v", err)
	}

	if got := h.writes.Load(); got != 1 {
		t.Errorf("a mutating request must issue exactly one write, got %d", got)
	}
}

func TestStoreSaveWritesBrandNewSession(t *testing.T) {
	ctx := context.Background()
	h := newCountingHandler()

	s := session.New("t", h)
	_ = s.Start(ctx)

	// No mutation at all, but a brand-new session must still be persisted.
	if err := s.Save(ctx); err != nil {
		t.Fatalf("save: %v", err)
	}

	if got := h.writes.Load(); got != 1 {
		t.Errorf("a brand-new session must be written even without mutations, got %d", got)
	}
}

func TestStoreSaveWritesAfterRegenerate(t *testing.T) {
	ctx := context.Background()
	h := newCountingHandler()

	s1 := session.New("t", h)
	_ = s1.Start(ctx)
	s1.Put("user", 42)
	_ = s1.Save(ctx)

	h.writes.Store(0)

	s2 := session.NewWithID("t", h, s1.GetID())
	_ = s2.Start(ctx)

	if err := s2.Regenerate(ctx, false); err != nil {
		t.Fatalf("regenerate: %v", err)
	}

	if err := s2.Save(ctx); err != nil {
		t.Fatalf("save: %v", err)
	}

	if got := h.writes.Load(); got != 1 {
		t.Errorf("a regenerated ID must be written even without mutations, got %d", got)
	}
}

func TestStoreSaveTouchActivityRefreshesOnInterval(t *testing.T) {
	now := time.Now()

	// Disabled: a non-positive interval never touches or dirties.
	s := cleanSeeded(t, nil)

	if s.TouchActivity(now, 0) {
		t.Error("a zero interval must not touch")
	}

	if s.IsDirty() {
		t.Error("a disabled touch must not dirty the session")
	}

	// Within the interval: the marker is fresh, so no write is needed.
	within := cleanSeeded(t, map[string]any{"_last_activity": now.Add(-time.Minute).Unix()})

	if within.TouchActivity(now, time.Hour) {
		t.Error("a marker within the interval must not touch")
	}

	if within.IsDirty() {
		t.Error("a within-interval touch must not dirty the session")
	}

	// Elapsed: the marker is stale, so the touch refreshes and dirties.
	elapsed := cleanSeeded(t, map[string]any{"_last_activity": now.Add(-2 * time.Hour).Unix()})

	if !elapsed.TouchActivity(now, time.Hour) {
		t.Error("a stale marker past the interval must touch")
	}

	if !elapsed.IsDirty() {
		t.Error("an interval touch must dirty the session so Save persists it")
	}

	// A fresh session with no marker touches on first activity.
	fresh := cleanSeeded(t, nil)

	if !fresh.TouchActivity(now, time.Hour) {
		t.Error("a session with no marker must touch on first activity")
	}
}

// hookedWriteHandler wraps ArrayHandler and runs a callback at the start of
// every backend Write, letting tests interleave store mutations with the
// out-of-lock write that Save performs.
type hookedWriteHandler struct {
	*handlers.ArrayHandler
	onWrite func()
}

func (h *hookedWriteHandler) Write(ctx context.Context, id, data string) error {
	if h.onWrite != nil {
		h.onWrite()
	}

	return h.ArrayHandler.Write(ctx, id, data)
}

func TestSaveKeepsDirtyWhenMutatedDuringBackendWrite(t *testing.T) {
	ctx := context.Background()
	h := &hookedWriteHandler{ArrayHandler: handlers.NewArrayHandler()}
	s := session.New("t", h)

	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}

	s.Put("k", "v1")

	// Mutate the session while Save's backend write is in flight (Save holds
	// no lock during handler.Write). The interleaved change must keep the
	// session dirty; clearing it would silently drop v2.
	h.onWrite = func() {
		h.onWrite = nil

		s.Put("k", "v2")
	}

	if err := s.Save(ctx); err != nil {
		t.Fatal(err)
	}

	if !s.IsDirty() {
		t.Fatal("mutation during the backend write must leave the session dirty")
	}

	if err := s.Save(ctx); err != nil {
		t.Fatal(err)
	}

	reloaded := session.NewWithID("t", h, s.GetID())

	if err := reloaded.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if got := reloaded.Get("k", nil); got != "v2" {
		t.Fatalf("k = %v, want v2 after the follow-up save", got)
	}
}
