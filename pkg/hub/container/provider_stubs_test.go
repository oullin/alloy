package container_test

import (
	"github.com/oullin/alloy/pkg/hub/container"
	"github.com/oullin/alloy/pkg/hub/container/contracts/provider"
)

// fakeProvider records lifecycle calls for assertions.
type fakeProvider struct {
	name          string
	registerCalls int
	bootCalls     int
	provides      []string
}

// nonBootable satisfies only ServiceProvider, not Bootable.
type nonBootable struct{ registerCalls int }

// orderRecorder records the sequence in which providers were called.
type orderRecorder struct {
	tag string
	log *[]string
}

// closureProvider is a tiny adapter that lets a test write inline Register logic.
type closureProvider struct{ register func() }

// deferredProvider claims abstract keys but only registers them lazily.
type deferredProvider struct {
	keys       []string
	register   func(*container.App)
	bootCalls  int
	registered bool
}

// depProvider declares both the keys it binds and the keys it needs, so
// RegisterMany can topologically sort it.
type depProvider struct {
	name     string
	provides []string
	depends  []string
	log      *[]string
}

var (
	_ provider.ServiceProvider = (*fakeProvider)(nil)
	_ provider.Bootable        = (*fakeProvider)(nil)
	_ provider.Deferred        = (*deferredProvider)(nil)
	_ provider.DependsOn       = (*depProvider)(nil)
)

func (p *fakeProvider) Register()          { p.registerCalls++ }
func (p *fakeProvider) Boot()              { p.bootCalls++ }
func (p *fakeProvider) Provides() []string { return p.provides }

func (p *nonBootable) Register() { p.registerCalls++ }

func (p *orderRecorder) Register() { *p.log = append(*p.log, "register:"+p.tag) }
func (p *orderRecorder) Boot()     { *p.log = append(*p.log, "boot:"+p.tag) }

func (p *closureProvider) Register() { p.register() }

// Register passes a nil container; tests that need the container use a closure.
func (p *deferredProvider) Register() {
	p.registered = true
	p.register(nil)
}

func (p *deferredProvider) Provides() []string { return p.keys }
func (p *deferredProvider) Deferred() bool     { return true }
func (p *deferredProvider) Boot()              { p.bootCalls++ }

func (p *depProvider) Register()           { *p.log = append(*p.log, p.name) }
func (p *depProvider) Provides() []string  { return p.provides }
func (p *depProvider) DependsOn() []string { return p.depends }
