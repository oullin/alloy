package authtest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/oullin/alloy/api/auth/sessionx"
	"github.com/oullin/alloy/api/auth/user"
	cauth "github.com/oullin/alloy/api/contracts/auth"
	"github.com/oullin/alloy/api/contracts/auth/events"
)

// ErrProvider is a reusable provider error for auth tests.

// Provider is a test user provider backed by a map.
type Provider struct {
	Users              map[string]cauth.User
	RetrieveByIDErr    error
	RetrieveByTokenErr error
}

// Session is a minimal in-memory session store.
type Session struct {
	Data           map[string]any
	MigrateCalls   int
	MigrateDestroy []bool
}

// NewSession creates a fresh in-memory session.

// CookieManager records queued and expired cookies.
type CookieManager struct {
	Queued    []*http.Cookie
	Forgotten []string
}

// Dispatcher collects dispatched events for assertions.
type Dispatcher struct {
	Mu     sync.Mutex
	Events []any
}

var ErrProvider = errors.New("stub provider error")

func (p *Provider) RetrieveByID(_ context.Context, id string) (cauth.User, error) {
	if p.RetrieveByIDErr != nil {
		return nil, p.RetrieveByIDErr
	}

	return p.Users[id], nil
}

func (p *Provider) RetrieveByToken(_ context.Context, id string, token string) (cauth.User, error) {
	if p.RetrieveByTokenErr != nil {
		return nil, p.RetrieveByTokenErr
	}

	u := p.Users[id]

	if u == nil || u.GetRememberToken() != token {
		return nil, nil
	}

	return u, nil
}

func (p *Provider) UpdateRememberToken(_ context.Context, user cauth.User, token string) error {
	user.SetRememberToken(token)

	return nil
}

func (p *Provider) RetrieveByCredentials(_ context.Context, creds map[string]string) (cauth.User, error) {
	for _, u := range p.Users {
		gen, ok := u.(*user.GenericUser)

		if !ok {
			continue
		}

		match := true

		for k, v := range creds {
			if k == "password" {
				continue
			}

			attr, _ := gen.Attributes[k].(string)

			if attr != v {
				match = false

				break
			}
		}

		if match {
			return u, nil
		}
	}

	return nil, nil
}

func (p *Provider) ValidateCredentials(_ context.Context, user cauth.User, creds map[string]string) (bool, error) {
	pw := creds["password"]

	return user.GetAuthPassword() == pw, nil
}

func (p *Provider) RehashPasswordIfRequired(context.Context, cauth.User, map[string]string, bool) error {
	return nil
}

func NewSession() *Session {
	return &Session{Data: make(map[string]any)}
}

func (s *Session) Get(key string, fallback any) any {
	if v, ok := s.Data[key]; ok {
		return v
	}

	return fallback
}

func (s *Session) Put(key string, value any) { s.Data[key] = value }

func (s *Session) Remove(key string) any {
	v := s.Data[key]
	delete(s.Data, key)

	return v
}

func (s *Session) Forget(keys ...string) {
	for _, k := range keys {
		delete(s.Data, k)
	}
}

func (s *Session) Migrate(_ context.Context, destroy bool) error {
	s.MigrateCalls++
	s.MigrateDestroy = append(s.MigrateDestroy, destroy)

	return nil
}

func (m *CookieManager) Queue(cookie *http.Cookie) error {
	m.Queued = append(m.Queued, cookie)

	return nil
}

func (m *CookieManager) Expire(name string, options sessionx.CookieOptions) error {
	m.Forgotten = append(m.Forgotten, name)
	m.Queued = append(m.Queued, &http.Cookie{Name: name, Path: options.Path, Domain: options.Domain, MaxAge: -1})

	return nil
}

func (d *Dispatcher) Listen(any, ...events.Listener)          {}
func (d *Dispatcher) HasListeners(any) bool                   { return false }
func (d *Dispatcher) HasWildcardListeners(any) bool           { return false }
func (d *Dispatcher) Subscribe(events.Subscriber)             {}
func (d *Dispatcher) Until(context.Context, any) (any, error) { return nil, nil }
func (d *Dispatcher) Push(context.Context, any)               {}
func (d *Dispatcher) Flush(context.Context, string) error     { return nil }
func (d *Dispatcher) Forget(any)                              {}
func (d *Dispatcher) ForgetPushed()                           {}
func (d *Dispatcher) GetListeners(any) []events.Listener      { return nil }
func (d *Dispatcher) Dispatch(_ context.Context, event any) ([]any, error) {
	d.Mu.Lock()

	defer d.Mu.Unlock()

	d.Events = append(d.Events, event)

	return nil, nil
}

// Has asserts that an event with the given type suffix was dispatched.
func (d *Dispatcher) Has(t *testing.T, typeName string) {
	t.Helper()

	d.Mu.Lock()

	defer d.Mu.Unlock()

	for _, e := range d.Events {
		if typeNameOf(e) == typeName {
			return
		}
	}

	t.Errorf("expected event %q to be dispatched, got %v", typeName, d.TypeNames())
}

// HasNot asserts that an event with the given type suffix was not dispatched.
func (d *Dispatcher) HasNot(t *testing.T, typeName string) {
	t.Helper()

	d.Mu.Lock()

	defer d.Mu.Unlock()

	for _, e := range d.Events {
		if typeNameOf(e) == typeName {
			t.Errorf("expected event %q NOT to be dispatched", typeName)

			return
		}
	}
}

// TypeNames returns the dispatched event type names.
func (d *Dispatcher) TypeNames() []string {
	names := make([]string, len(d.Events))

	for i, e := range d.Events {
		names[i] = typeNameOf(e)
	}

	return names
}

func typeNameOf(v any) string {
	return fmt.Sprintf("%T", v)
}
