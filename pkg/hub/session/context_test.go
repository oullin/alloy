package session_test

import (
	"context"
	"testing"

	"github.com/oullin/alloy/pkg/hub/session"
	"github.com/oullin/alloy/pkg/hub/session/handlers"
)

func TestFromContextRoundTrip(t *testing.T) {
	t.Parallel()

	store := session.New("session", handlers.NewArrayHandler())
	ctx := session.NewContext(context.Background(), store)

	got, ok := session.FromContext(ctx)

	if !ok {
		t.Fatal("expected store to be present in context")
	}

	if got != store {
		t.Errorf("expected same store instance, got %p, want %p", got, store)
	}
}

func TestFromContextBareContext(t *testing.T) {
	t.Parallel()

	got, ok := session.FromContext(context.Background())

	if ok {
		t.Error("expected ok to be false on a bare context")
	}

	if got != nil {
		t.Errorf("expected nil store, got %p", got)
	}
}
