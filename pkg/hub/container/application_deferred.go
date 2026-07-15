package container

import "github.com/oullin/alloy/pkg/hub/container/contracts/provider"

// recordDeferred remembers the provider's keys without calling Register.
// The provider IS appended to a.providers so introspection sees it.
func (a *Application) recordDeferred(p provider.ServiceProvider) {
	provides, ok := p.(provider.Provides)

	if !ok {
		a.doRegister(p)

		return
	}

	keys := provides.Provides()

	if len(keys) == 0 {
		a.doRegister(p)

		return
	}

	a.providers = append(a.providers, p)

	for _, key := range keys {
		a.deferredByKey[key] = p
	}
}

// flushDeferredFor runs the deferred provider that owns abstract, if any,
// and removes its tracking entries.
func (a *Application) flushDeferredFor(abstract string) {
	p, ok := a.deferredByKey[abstract]

	if !ok {
		return
	}

	if a.registered[p] {
		return
	}

	if provides, ok := p.(provider.Provides); ok {
		for _, k := range provides.Provides() {
			delete(a.deferredByKey, k)
		}
	}

	p.Register()
	a.registered[p] = true

	// If the application has already booted, give the freshly-registered
	// provider a chance to Boot too.
	if a.booted {
		if b, ok := p.(provider.Bootable); ok {
			b.Boot()
		}
	}
}

// Make resolves an abstract through the container, flushing any deferred
// provider that claims the key first. This shadows App.Make so
// callers that hold an *Application get deferred resolution automatically.
func (a *Application) Make(abstract string) (any, error) {
	a.flushDeferredFor(abstract)

	return a.App.Make(abstract)
}

// MakeWith is the parameterised counterpart of Make. It also flushes any
// deferred provider for the abstract.
func (a *Application) MakeWith(abstract string, parameters map[string]any) (any, error) {
	a.flushDeferredFor(abstract)

	return a.App.MakeWith(abstract, parameters)
}

// Get is the PSR-11 alias of Make. Same deferred-flush behaviour.
func (a *Application) Get(abstract string) (any, error) {
	a.flushDeferredFor(abstract)

	return a.App.Get(abstract)
}

// isDeferred returns true when p opts in via provider.Deferred AND declares
// its keys via provider.Provides.
func isDeferred(p provider.ServiceProvider) bool {
	d, ok := p.(provider.Deferred)

	if !ok {
		return false
	}

	if !d.Deferred() {
		return false
	}

	_, hasProvides := p.(provider.Provides)

	return hasProvides
}
