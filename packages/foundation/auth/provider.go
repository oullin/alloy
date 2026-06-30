package auth

import (
	"alloy.dev/foundation/auth/manager"
	"alloy.dev/foundation/container"
)

// ServiceProvider registers the authentication registry into the container.
type ServiceProvider struct {
	app          *container.App
	defaultGuard string
	onBoot       func(*manager.Registry)
}

// NewServiceProvider constructs the provider.
// defaultGuard is the name of the guard to use by default (e.g. "session").
func NewServiceProvider(app *container.App, defaultGuard string) *ServiceProvider {
	return &ServiceProvider{app: app, defaultGuard: defaultGuard}
}

// WithBoot registers a callback invoked at Boot time with the resolved
// auth registry. Use this to register custom guards/providers and the user
// resolver after every other provider has finished its Register phase.
func (p *ServiceProvider) WithBoot(fn func(*manager.Registry)) *ServiceProvider {
	p.onBoot = fn

	return p
}

// Register binds the auth registry as a singleton under "auth".
func (p *ServiceProvider) Register() {
	p.app.Singleton("auth", func(_ *container.App) (any, error) {
		return manager.New(p.defaultGuard), nil
	})
}

// Boot resolves the registry and runs the user-supplied boot callback, if any.
func (p *ServiceProvider) Boot() {
	if p.onBoot == nil {
		return
	}

	raw, err := p.app.Make("auth")

	if err != nil {
		return
	}

	if m, ok := raw.(*manager.Registry); ok {
		p.onBoot(m)
	}
}

// Provides returns the abstract keys registered by this provider.
func (p *ServiceProvider) Provides() []string {
	return []string{"auth"}
}
