package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestForgetInstanceForgetsInstance(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("name", "Taylor")

	c.ForgetInstance("name")

	if c.Bound("name") {
		t.Fatal("instance should be forgotten (not in bindings)")
	}
}

func TestForgetInstancesForgetsAll(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("a", "1")
	c.Instance("b", "2")

	c.ForgetInstances()

	if c.Bound("a") || c.Bound("b") {
		t.Fatal("all instances should be forgotten")
	}
}

func TestForgetScopedInstancesOnlyClearsScoped(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Singleton("shared", func(_ *container.App) (any, error) {
		return "shared", nil
	})

	c.Scoped("scoped", func(_ *container.App) (any, error) {
		return "scoped", nil
	})

	c.Make("shared") //nolint:errcheck
	c.Make("scoped") //nolint:errcheck

	c.ForgetScopedInstances()

	// Shared should still return cached instance.
	a, _ := c.Make("shared")

	if a != "shared" {
		t.Fatal("shared instance should survive ForgetScopedInstances")
	}
}

func TestGetBindings(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("a", func(_ *container.App) (any, error) { return "a", nil }, false)
	c.Singleton("b", func(_ *container.App) (any, error) { return "b", nil })

	bindings := c.GetBindings()

	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
}

func TestUnsetRemovesBoundInstances(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("stub", "value")

	if !c.Bound("stub") {
		t.Fatal("expected bound")
	}

	c.ForgetInstance("stub")

	if c.Bound("stub") {
		t.Fatal("expected unbound after forget")
	}
}
