package container

import (
	"fmt"

	"hara.sh/alloy/container/contracts/provider"
	"hara.sh/alloy/container/internal/provsort"
)

// Register calls p.Register() and records the provider so Boot can
// call p.Boot() if the provider also implements provider.Bootable.
//
// Deferred providers (implementing provider.Deferred AND provider.Provides
// with Deferred() returning true) are NOT registered immediately. Their
// abstract keys are tracked, and Register() runs the first time any
// tracked key is resolved through Application.Make.
func (a *Application) Register(p provider.ServiceProvider) {
	a.mu.Lock()

	// Guards both paths: an eagerly-registered provider and a deferred
	// provider that has already been flushed must not be re-tracked, or
	// a.providers would accumulate duplicates and Boot would run twice.
	if a.registered[p] {
		a.mu.Unlock()

		return
	}

	if isDeferred(p) {
		provides, ok := p.(provider.Provides)

		if ok {
			keys := provides.Provides()

			if len(keys) > 0 {
				// Already tracked as deferred (registered twice before its
				// first flush): keep the single existing entry.
				for _, tracked := range a.deferredByKey {
					if tracked == p {
						a.mu.Unlock()

						return
					}
				}

				a.providers = append(a.providers, p)

				for _, key := range keys {
					a.deferredByKey[key] = p
				}

				a.mu.Unlock()

				return
			}
		}
	}

	a.providers = append(a.providers, p)
	a.registered[p] = true
	isBooted := a.booted
	a.mu.Unlock()

	p.Register()

	if isBooted {
		if b, ok := p.(provider.Bootable); ok {
			b.Boot()
		}
	}
}

// RegisterMany calls Register on each provider, topologically sorted by
// any provider.DependsOn declarations so dependencies always come first.
// Cycles cause a panic. Providers without DependsOn keep their original
// relative order (the sort is stable).
func (a *Application) RegisterMany(providers []provider.ServiceProvider) {
	sorted, err := provsort.Sort(providers)

	if err != nil {
		panic(fmt.Errorf("%w: %v", ErrProviderCycle, err))
	}

	for _, p := range sorted {
		a.Register(p)
	}
}

// Boot calls Boot() on every eagerly-registered provider that implements
// provider.Bootable. Deferred providers that have not yet been flushed are
// NOT booted until their first Make() call. Idempotent.
func (a *Application) Boot() {
	a.mu.Lock()

	if a.booted {
		a.mu.Unlock()

		return
	}

	a.booted = true

	// Snapshot currently registered providers to boot outside the lock
	providers := make([]provider.ServiceProvider, len(a.providers))
	copy(providers, a.providers)

	var bootedProviders []provider.ServiceProvider

	for _, p := range providers {
		if a.registered[p] {
			bootedProviders = append(bootedProviders, p)
		}
	}

	a.mu.Unlock()

	for _, p := range bootedProviders {
		if b, ok := p.(provider.Bootable); ok {
			b.Boot()
		}
	}
}

// Booted reports whether Boot has been called.
func (a *Application) Booted() bool {
	a.mu.Lock()

	defer a.mu.Unlock()

	return a.booted
}
