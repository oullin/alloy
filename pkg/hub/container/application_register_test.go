package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
	"github.com/oullin/alloy/pkg/hub/container/contracts/provider"
)

func TestApplication_Register_CallsRegisterOnce(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()
	p := &fakeProvider{name: "p1"}

	app.Register(p)

	if p.registerCalls != 1 {
		t.Fatalf("expected Register to be called once, got %d", p.registerCalls)
	}

	if got := len(app.Providers()); got != 1 {
		t.Fatalf("expected 1 provider stored, got %d", got)
	}
}

func TestApplication_Boot_OnlyCallsBootable(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()
	bootable := &fakeProvider{name: "bootable"}
	plain := &nonBootable{}

	app.Register(bootable)
	app.Register(plain)
	app.Boot()

	if bootable.bootCalls != 1 {
		t.Fatalf("expected Bootable provider to receive one Boot call, got %d", bootable.bootCalls)
	}

	if plain.registerCalls != 1 {
		t.Fatalf("expected non-bootable to still be registered, got %d", plain.registerCalls)
	}

	if !app.Booted() {
		t.Fatal("expected Booted() to be true after Boot()")
	}
}

func TestApplication_Boot_IsIdempotent(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()
	p := &fakeProvider{name: "p"}
	app.Register(p)

	app.Boot()
	app.Boot()
	app.Boot()

	if p.bootCalls != 1 {
		t.Fatalf("expected Boot to be called once across multiple app.Boot() calls, got %d", p.bootCalls)
	}
}

func TestApplication_Boot_PreservesRegistrationOrder(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()
	log := []string{}

	app.Register(&orderRecorder{tag: "a", log: &log})
	app.Register(&orderRecorder{tag: "b", log: &log})
	app.Register(&orderRecorder{tag: "c", log: &log})
	app.Boot()

	want := []string{
		"register:a", "register:b", "register:c",
		"boot:a", "boot:b", "boot:c",
	}

	if len(log) != len(want) {
		t.Fatalf("expected %d events, got %d: %v", len(want), len(log), log)
	}

	for i, ev := range want {
		if log[i] != ev {
			t.Fatalf("at %d: expected %q, got %q (full: %v)", i, ev, log[i], log)
		}
	}
}

func TestApplication_LazyResolution_HandlesOutOfOrderRegistration(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()

	app.Register(&closureProvider{
		register: func() {
			app.Singleton("B", func(c *container.App) (any, error) {
				rawA, err := c.Make("A")

				if err != nil {
					return nil, err
				}

				return "B(" + rawA.(string) + ")", nil
			})
		},
	})

	app.Register(&closureProvider{
		register: func() {
			app.Singleton("A", func(_ *container.App) (any, error) {
				return "A", nil
			})
		},
	})

	v, err := app.Make("B")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v != "B(A)" {
		t.Fatalf("expected B(A), got %v", v)
	}
}

func TestApplication_RegisterMany_RegistersAllInOrder(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()
	log := []string{}

	app.RegisterMany([]provider.ServiceProvider{
		&orderRecorder{tag: "a", log: &log},
		&orderRecorder{tag: "b", log: &log},
		&orderRecorder{tag: "c", log: &log},
	})

	if len(app.Providers()) != 3 {
		t.Fatalf("expected 3 providers, got %d", len(app.Providers()))
	}

	want := []string{"register:a", "register:b", "register:c"}

	for i, ev := range want {
		if log[i] != ev {
			t.Fatalf("at %d: expected %q, got %q", i, ev, log[i])
		}
	}
}

func TestApplication_RegisterMany_TopoSortsByDependsOn(t *testing.T) {
	t.Parallel()

	log := []string{}

	// C depends on B which depends on A. Register them in REVERSE order
	// to prove the sort actually runs.
	pA := &depProvider{name: "A", provides: []string{"a"}, log: &log}
	pB := &depProvider{name: "B", provides: []string{"b"}, depends: []string{"a"}, log: &log}
	pC := &depProvider{name: "C", provides: []string{"c"}, depends: []string{"b"}, log: &log}

	app := container.NewApplication()
	app.RegisterMany([]provider.ServiceProvider{pC, pB, pA})

	want := []string{"A", "B", "C"}

	if len(log) != len(want) {
		t.Fatalf("expected %v, got %v", want, log)
	}

	for i, ev := range want {
		if log[i] != ev {
			t.Fatalf("at %d: expected %q, got %q (full: %v)", i, ev, log[i], log)
		}
	}
}

func TestApplication_RegisterMany_PanicsOnCycle(t *testing.T) {
	t.Parallel()

	log := []string{}

	pA := &depProvider{name: "A", provides: []string{"a"}, depends: []string{"b"}, log: &log}
	pB := &depProvider{name: "B", provides: []string{"b"}, depends: []string{"a"}, log: &log}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on cycle")
		}
	}()

	container.NewApplication().RegisterMany([]provider.ServiceProvider{pA, pB})
}
