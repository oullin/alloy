package container

import (
	"fmt"

	"github.com/oullin/alloy/pkg/hub/container/contracts/provider"
	"github.com/oullin/alloy/pkg/hub/container/internal/provsort"
)

// Register calls p.Register() and records the provider so Boot can
// call p.Boot() if the provider also implements provider.Bootable.
//
// Deferred providers (implementing provider.Deferred AND provider.Provides
// with Deferred() returning true) are NOT registered immediately. Their
// abstract keys are tracked, and Register() runs the first time any
// tracked key is resolved through Application.Make.
func (a *Application) Register(p provider.ServiceProvider) {
	if isDeferred(p) {
		a.recordDeferred(p)

		return
	}

	a.doRegister(p)
}

// doRegister performs the actual provider registration.
func (a *Application) doRegister(p provider.ServiceProvider) {
	if a.registered[p] {
		return
	}

	p.Register()
	a.providers = append(a.providers, p)
	a.registered[p] = true
}

// RegisterMany calls Register on each provider, topologically sorted by
// any provider.DependsOn declarations so dependencies always come first.
// Cycles cause a panic. Providers without DependsOn keep their original
// relative order (the sort is stable).
func (a *Application) RegisterMany(providers []provider.ServiceProvider) {
	sorted, err := provsort.Sort(providers)

	if err != nil {
		panic(fmt.Sprintf("container: RegisterMany: %v", err))
	}

	for _, p := range sorted {
		a.Register(p)
	}
}

// Boot calls Boot() on every eagerly-registered provider that implements
// provider.Bootable. Deferred providers that have not yet been flushed are
// NOT booted until their first Make() call. Idempotent.
func (a *Application) Boot() {
	if a.booted {
		return
	}

	for _, p := range a.providers {
		if !a.registered[p] {
			continue // deferred-and-not-yet-flushed
		}

		if b, ok := p.(provider.Bootable); ok {
			b.Boot()
		}
	}

	a.booted = true
}

// Booted reports whether Boot has been called.
func (a *Application) Booted() bool {
	return a.booted
}
