package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestBound(t *testing.T) {
	t.Parallel()

	c := newContainer()

	if c.Bound("something") {
		t.Fatal("should not be bound")
	}

	c.Bind("something", func(_ *container.App) (any, error) {
		return "value", nil
	}, false)

	if !c.Bound("something") {
		t.Fatal("should be bound")
	}
}

func TestBoundWithInstance(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("name", "Taylor")

	if !c.Bound("name") {
		t.Fatal("instance should be bound")
	}
}

func TestHasDelegatesToBound(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	if !c.Has("name") {
		t.Fatal("Has should delegate to Bound")
	}

	if c.Has("missing") {
		t.Fatal("Has should return false for unbound")
	}
}

func TestResolved(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	if c.Resolved("name") {
		t.Fatal("should not be resolved before Make")
	}

	c.Make("name") //nolint:errcheck

	if !c.Resolved("name") {
		t.Fatal("should be resolved after Make")
	}
}

func TestResolvedWithInstance(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("name", "Taylor")

	if !c.Resolved("name") {
		t.Fatal("instance should report as resolved")
	}
}

func TestIsShared(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("transient", func(_ *container.App) (any, error) {
		return "value", nil
	}, false)

	c.Singleton("shared", func(_ *container.App) (any, error) {
		return "value", nil
	})

	if c.IsShared("transient") {
		t.Fatal("transient should not be shared")
	}

	if !c.IsShared("shared") {
		t.Fatal("singleton should be shared")
	}
}

func TestIsSharedWithScoped(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Scoped("scoped", func(_ *container.App) (any, error) {
		return "value", nil
	})

	if !c.IsShared("scoped") {
		t.Fatal("scoped should be shared")
	}
}

func TestIsSharedWithInstance(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("name", "Taylor")

	if !c.IsShared("name") {
		t.Fatal("instance should be shared")
	}
}

func TestCurrentlyResolving(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var resolving string

	c.Bind("outer", func(cc *container.App) (any, error) {
		resolving = cc.CurrentlyResolving()

		return "outer", nil
	}, false)

	c.Make("outer") //nolint:errcheck

	if resolving != "outer" {
		t.Fatalf("expected outer, got %s", resolving)
	}
}

func TestCurrentlyResolvingEmpty(t *testing.T) {
	t.Parallel()

	c := newContainer()

	if c.CurrentlyResolving() != "" {
		t.Fatal("expected empty string when not resolving")
	}
}
