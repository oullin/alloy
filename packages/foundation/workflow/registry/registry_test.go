package registry_test

import (
	"testing"

	"alloy.dev/foundation/workflow"
	"alloy.dev/foundation/workflow/registry"
	"alloy.dev/foundation/workflow/store"
)

type subject struct {
	kind  string
	state string
}

func newMachine(t *testing.T, name string) *workflow.Machine[*subject] {
	t.Helper()

	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("b").
		SetInitialPlaces("a").
		AddTransition("go", []string{"a"}, []string{"b"}).
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	machine, err := workflow.New(name, def, &store.SingleState[*subject]{
		Getter: func(s *subject) string { return s.state },
		Setter: func(s *subject, place string) { s.state = place },
	}, nil)

	if err != nil {
		t.Fatalf("new machine: %v", err)
	}

	return machine
}

func TestStore_GetByName(t *testing.T) {
	reg := registry.New[*subject]()

	reg.Add(registry.Entry[*subject]{Name: "orders", Machine: newMachine(t, "orders")})
	reg.Add(registry.Entry[*subject]{Name: "billing", Machine: newMachine(t, "billing")})

	got, err := reg.Get(&subject{}, "billing")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.Name() != "billing" {
		t.Fatalf("expected billing workflow, got %q", got.Name())
	}
}

func TestStore_GetUnknownName(t *testing.T) {
	reg := registry.New[*subject]()

	reg.Add(registry.Entry[*subject]{Name: "orders", Machine: newMachine(t, "orders")})

	if _, err := reg.Get(&subject{}, "missing"); err == nil {
		t.Fatal("expected error for unknown workflow name")
	}
}

func TestStore_GetWithoutNameUsesSupportStrategy(t *testing.T) {
	reg := registry.New[*subject]()

	reg.Add(registry.Entry[*subject]{
		Name:     "orders",
		Machine:  newMachine(t, "orders"),
		Supports: func(s *subject) bool { return s.kind == "order" },
	})
	reg.Add(registry.Entry[*subject]{
		Name:     "billing",
		Machine:  newMachine(t, "billing"),
		Supports: func(s *subject) bool { return s.kind == "invoice" },
	})

	got, err := reg.Get(&subject{kind: "invoice"}, "")

	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.Name() != "billing" {
		t.Fatalf("expected billing workflow, got %q", got.Name())
	}
}

func TestStore_GetNoMatchingSubject(t *testing.T) {
	reg := registry.New[*subject]()

	reg.Add(registry.Entry[*subject]{
		Name:     "orders",
		Machine:  newMachine(t, "orders"),
		Supports: func(*subject) bool { return false },
	})

	if _, err := reg.Get(&subject{}, ""); err == nil {
		t.Fatal("expected error when no workflow supports the subject")
	}
}

func TestStore_SupportsFiltersNamedLookups(t *testing.T) {
	reg := registry.New[*subject]()

	reg.Add(registry.Entry[*subject]{
		Name:     "orders",
		Machine:  newMachine(t, "orders"),
		Supports: func(*subject) bool { return false },
	})

	if _, err := reg.Get(&subject{}, "orders"); err == nil {
		t.Fatal("expected error when the named workflow rejects the subject")
	}
}

func TestStore_AddIgnoresNilMachine(t *testing.T) {
	reg := registry.New[*subject]()

	reg.Add(registry.Entry[*subject]{Name: "orders"})

	if _, err := reg.Get(&subject{}, "orders"); err == nil {
		t.Fatal("nil-machine entries should be ignored")
	}
}
