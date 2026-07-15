package container

import "github.com/oullin/alloy/pkg/hub/container/contracts/provider"

// Providers returns every service provider that has been recorded with the
// application, including deferred providers that have not yet been flushed.
// The returned slice is a copy.
func (a *Application) Providers() []provider.ServiceProvider {
	a.mu.Lock()

	defer a.mu.Unlock()

	out := make([]provider.ServiceProvider, len(a.providers))
	copy(out, a.providers)

	return out
}

// HasProvider reports whether any registered or deferred provider declares
// the given abstract key via provider.Provides.
func (a *Application) HasProvider(abstract string) bool {
	a.mu.Lock()

	defer a.mu.Unlock()

	if _, ok := a.deferredByKey[abstract]; ok {
		return true
	}

	for _, p := range a.providers {
		if provides, ok := p.(provider.Provides); ok {
			for _, k := range provides.Provides() {
				if k == abstract {
					return true
				}
			}
		}
	}

	return false
}

// ProviderFor returns the (first) provider that declares the given abstract,
// or nil if none does.
func (a *Application) ProviderFor(abstract string) provider.ServiceProvider {
	a.mu.Lock()

	defer a.mu.Unlock()

	if p, ok := a.deferredByKey[abstract]; ok {
		return p
	}

	for _, p := range a.providers {
		if provides, ok := p.(provider.Provides); ok {
			for _, k := range provides.Provides() {
				if k == abstract {
					return p
				}
			}
		}
	}

	return nil
}
