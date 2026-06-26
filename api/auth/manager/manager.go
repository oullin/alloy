package manager

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/oullin/alloy/api/auth/errorsx"
	"github.com/oullin/alloy/api/auth/httpx"
	cauth "github.com/oullin/alloy/api/contracts/auth"
)

// RequestCallback resolves the user from an HTTP request.
type RequestCallback = httpx.Callback

// GuardCreator creates a guard from a config map.
type GuardCreator func(name string, config map[string]any, provider cauth.UserProvider) (cauth.Guard, error)

// ProviderCreator creates a user provider from a config map.
type ProviderCreator func(config map[string]any) (cauth.UserProvider, error)

// Registry creates and manages named guards and user providers.
type Registry struct {
	mu               sync.RWMutex
	guards           map[string]cauth.Guard
	guardCreators    map[string]GuardCreator
	providers        map[string]cauth.UserProvider
	providerCreators map[string]ProviderCreator
	configs          map[string]map[string]any
	defaultGuard     string
	userResolver     func(context.Context) cauth.User
}

// New creates a Registry with no pre-registered drivers.
func New(defaultGuard string) *Registry {
	return &Registry{
		guards:           make(map[string]cauth.Guard),
		guardCreators:    make(map[string]GuardCreator),
		providers:        make(map[string]cauth.UserProvider),
		providerCreators: make(map[string]ProviderCreator),
		configs:          make(map[string]map[string]any),
		defaultGuard:     defaultGuard,
	}
}

// Extend registers a custom guard driver factory.
func (m *Registry) Extend(driver string, creator GuardCreator) *Registry {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.guardCreators[driver] = creator

	return m
}

// Provider registers a custom user provider factory.
func (m *Registry) Provider(name string, creator ProviderCreator) *Registry {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.providerCreators[name] = creator

	return m
}

// ViaRequest registers a guard that resolves users via a custom callback.
func (m *Registry) ViaRequest(name string, callback RequestCallback) *Registry {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.guards[name] = httpx.NewRequestGuard(callback)

	return m
}

// SetConfig stores configuration for a named guard.
func (m *Registry) SetConfig(name string, config map[string]any) *Registry {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.configs[name] = config

	return m
}

// Guard returns the guard identified by name. Uses the default if name is "".
func (m *Registry) Guard(ctx context.Context, name string) (cauth.Guard, error) {
	if name == "" {
		name = m.defaultGuard
	}

	m.mu.Lock()

	defer m.mu.Unlock()

	if g, ok := m.guards[name]; ok {
		return g, nil
	}

	cfg := m.configs[name]
	driver, _ := cfg["driver"].(string)

	creator, ok := m.guardCreators[driver]

	if !ok {
		return nil, fmt.Errorf("%w: %q (driver: %q)", errorsx.ErrInvalidGuard, name, driver)
	}

	var provider cauth.UserProvider

	if providerName, ok := cfg["provider"].(string); ok {
		p, err := m.resolveProvider(ctx, providerName)

		if err != nil {
			return nil, err
		}

		provider = p
	}

	guard, err := creator(name, cfg, provider)

	if err != nil {
		return nil, fmt.Errorf("auth: create guard %q: %w", name, err)
	}

	m.guards[name] = guard

	return guard, nil
}

// SetRequest sets the HTTP request on all request-aware guards.
func (m *Registry) SetRequest(r *http.Request) {
	m.mu.RLock()

	defer m.mu.RUnlock()

	type requestSetter interface{ SetRequest(*http.Request) }

	for _, g := range m.guards {
		if rs, ok := g.(requestSetter); ok {
			rs.SetRequest(r)
		}
	}
}

// ShouldUse sets the guard that should be used by default.
func (m *Registry) ShouldUse(name string) {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.defaultGuard = name
}

// GetDefaultDriver returns the name of the default guard.
func (m *Registry) GetDefaultDriver() string {
	m.mu.RLock()

	defer m.mu.RUnlock()

	return m.defaultGuard
}

// SetDefaultDriver sets the name of the default guard.
func (m *Registry) SetDefaultDriver(name string) {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.defaultGuard = name
}

// HasResolvedGuards reports whether any guards have been resolved.
func (m *Registry) HasResolvedGuards() bool {
	m.mu.RLock()

	defer m.mu.RUnlock()

	return len(m.guards) > 0
}

// ForgetGuards clears all resolved guard instances.
func (m *Registry) ForgetGuards() {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.guards = make(map[string]cauth.Guard)
}

// UserResolver returns the user resolver function.
func (m *Registry) UserResolver() func(context.Context) cauth.User {
	m.mu.RLock()

	defer m.mu.RUnlock()

	return m.userResolver
}

// ResolveUsersUsing sets the user resolver function.
func (m *Registry) ResolveUsersUsing(fn func(context.Context) cauth.User) {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.userResolver = fn
}

func (m *Registry) resolveProvider(ctx context.Context, name string) (cauth.UserProvider, error) {
	if p, ok := m.providers[name]; ok {
		return p, nil
	}

	creator, ok := m.providerCreators[name]

	if !ok {
		return nil, fmt.Errorf("%w: %q", errorsx.ErrInvalidProvider, name)
	}

	cfg := m.configs["provider."+name]

	if cfg == nil {
		cfg = make(map[string]any)
	}

	p, err := creator(cfg)

	if err != nil {
		return nil, fmt.Errorf("auth: create provider %q: %w", name, err)
	}

	m.providers[name] = p

	return p, nil
}
