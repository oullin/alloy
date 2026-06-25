package auth

// ServiceFactory creates a service from the configured container.
type ServiceFactory func(ServiceContainer) (any, error)

// ServiceContainer is the minimal container surface required by AuthServiceProvider.
type ServiceContainer interface {
	Singleton(abstract string, factory ServiceFactory)
	Make(abstract string) (any, error)
}

// AuthServiceProvider registers the authentication manager into the container.
type AuthServiceProvider struct {
	app          ServiceContainer
	defaultGuard string
	onBoot       func(*Manager)
}

// NewAuthServiceProvider constructs the provider.
// defaultGuard is the name of the guard to use by default (e.g. "session").
func NewAuthServiceProvider(app ServiceContainer, defaultGuard string) *AuthServiceProvider {
	return &AuthServiceProvider{app: app, defaultGuard: defaultGuard}
}

// WithBoot registers a callback invoked at Boot time with the resolved
// auth manager. Use this to register custom guards/providers and the user
// resolver after every other provider has finished its Register phase.
func (p *AuthServiceProvider) WithBoot(fn func(*Manager)) *AuthServiceProvider {
	p.onBoot = fn

	return p
}

// Register binds the auth manager as a singleton under "auth".
func (p *AuthServiceProvider) Register() {
	p.app.Singleton("auth", func(_ ServiceContainer) (any, error) {
		return NewManager(p.defaultGuard), nil
	})
}

// Boot resolves the manager and runs the user-supplied boot callback, if any.
func (p *AuthServiceProvider) Boot() {
	if p.onBoot == nil {
		return
	}

	raw, err := p.app.Make("auth")

	if err != nil {
		return
	}

	if m, ok := raw.(*Manager); ok {
		p.onBoot(m)
	}
}

// Provides returns the abstract keys registered by this provider.
func (p *AuthServiceProvider) Provides() []string {
	return []string{"auth"}
}
