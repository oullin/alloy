package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestReboundListeners(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var rebound bool

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	c.Rebinding("name", func(_ any, _ *container.App) { //nolint:errcheck
		rebound = true
	})

	// Rebinding triggers the callback because "name" is already bound.
	if !rebound {
		t.Fatal("expected rebound callback to fire")
	}
}

func TestReboundListenersOnInstances(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("name", "Taylor")

	var called bool

	c.Rebinding("name", func(_ any, _ *container.App) { //nolint:errcheck
		called = true
	})

	if !called {
		t.Fatal("expected callback on already-bound instance")
	}
}

func TestReboundListenersNotCalledWhenNotBound(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var called bool

	c.Rebinding("name", func(_ any, _ *container.App) { //nolint:errcheck
		called = true
	})

	if called {
		t.Fatal("should not call callback when not bound")
	}
}

func TestReboundListenersFiredOnRebind(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	var newVal any

	c.Rebinding("name", func(instance any, _ *container.App) { //nolint:errcheck
		newVal = instance
	})

	// Rebind.
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Dayle", nil
	}, false)

	if newVal != "Dayle" {
		t.Fatalf("expected Dayle, got %v", newVal)
	}
}

func TestRefreshUpdatesTarget(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("name", "Taylor")

	var updated any

	c.Refresh("name", func(v any) {
		updated = v
	})

	if updated != "Taylor" {
		t.Fatalf("expected Taylor, got %v", updated)
	}

	c.Instance("name", "Dayle")

	if updated != "Dayle" {
		t.Fatalf("expected Dayle after refresh, got %v", updated)
	}
}
