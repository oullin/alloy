package provider

// ServiceProvider is implemented by packages that register their services.
type ServiceProvider interface {
	Register()
}

// Bootable is implemented by providers that require post-registration setup.
type Bootable interface {
	Boot()
}

// Provides is an optional hint from a provider declaring which container
// abstract keys it binds.
type Provides interface {
	Provides() []string
}

// Deferred is a marker interface for lazily registered providers.
type Deferred interface {
	Deferred() bool
}

// DependsOn is an optional hint declaring abstract keys that must be
// registered before this provider's Register runs.
type DependsOn interface {
	DependsOn() []string
}
