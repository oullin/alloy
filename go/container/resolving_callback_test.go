package container_test

import (
	"testing"

	"alloy.dev/api/container"
)

func TestResolvingCallbackFired(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var called bool

	c.Resolving("name", func(_ any, _ *container.App) {
		called = true
	})

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	if !called {
		t.Fatal("resolving callback should have been called")
	}
}

func TestResolvingCallbackReceivesInstance(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var received any

	c.Resolving("name", func(instance any, _ *container.App) {
		received = instance
	})

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	if received != "Taylor" {
		t.Fatalf("expected Taylor, got %v", received)
	}
}

func TestAfterResolvingCallbackFired(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var called bool

	c.AfterResolving("name", func(_ any, _ *container.App) {
		called = true
	})

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	if !called {
		t.Fatal("after resolving callback should have been called")
	}
}

func TestBeforeResolvingCallbackFired(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var called bool

	c.BeforeResolving("name", func(_ string, _ map[string]any, _ *container.App) {
		called = true
	})

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	if !called {
		t.Fatal("before resolving callback should have been called")
	}
}

func TestBeforeResolvingCallbackReceivesAbstract(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var receivedAbstract string

	c.BeforeResolving("name", func(abstract string, _ map[string]any, _ *container.App) {
		receivedAbstract = abstract
	})

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	if receivedAbstract != "name" {
		t.Fatalf("expected name, got %s", receivedAbstract)
	}
}

func TestGlobalResolvingCallbackFired(t *testing.T) {
	t.Parallel()

	c := newContainer()

	count := 0

	c.ResolvingAny(func(_ any, _ *container.App) {
		count++
	})

	c.Bind("a", func(_ *container.App) (any, error) { return "a", nil }, false)
	c.Bind("b", func(_ *container.App) (any, error) { return "b", nil }, false)

	c.Make("a") //nolint:errcheck
	c.Make("b") //nolint:errcheck

	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestGlobalAfterResolvingCallbackFired(t *testing.T) {
	t.Parallel()

	c := newContainer()

	count := 0

	c.AfterResolvingAny(func(_ any, _ *container.App) {
		count++
	})

	c.Bind("a", func(_ *container.App) (any, error) { return "a", nil }, false)
	c.Bind("b", func(_ *container.App) (any, error) { return "b", nil }, false)

	c.Make("a") //nolint:errcheck
	c.Make("b") //nolint:errcheck

	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestGlobalBeforeResolvingCallbackFired(t *testing.T) {
	t.Parallel()

	c := newContainer()

	count := 0

	c.BeforeResolvingAny(func(_ string, _ map[string]any, _ *container.App) {
		count++
	})

	c.Bind("a", func(_ *container.App) (any, error) { return "a", nil }, false)
	c.Bind("b", func(_ *container.App) (any, error) { return "b", nil }, false)

	c.Make("a") //nolint:errcheck
	c.Make("b") //nolint:errcheck

	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestResolvingCallbackOnlyForSpecificAbstract(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var calledFor string

	c.Resolving("name", func(_ any, _ *container.App) {
		calledFor = "name"
	})

	c.Bind("name", func(_ *container.App) (any, error) { return "Taylor", nil }, false)
	c.Bind("other", func(_ *container.App) (any, error) { return "Other", nil }, false)

	c.Make("other") //nolint:errcheck

	if calledFor != "" {
		t.Fatal("resolving callback should not fire for different abstract")
	}

	c.Make("name") //nolint:errcheck

	if calledFor != "name" {
		t.Fatalf("expected name, got %s", calledFor)
	}
}

func TestResolvingCallbackCalledOnceForSingleton(t *testing.T) {
	t.Parallel()

	c := newContainer()

	count := 0

	c.Resolving("name", func(_ any, _ *container.App) {
		count++
	})

	c.Singleton("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	})

	c.Make("name") //nolint:errcheck
	c.Make("name") //nolint:errcheck

	if count != 1 {
		t.Fatalf("expected 1 callback for singleton, got %d", count)
	}
}

func TestResolvingCallbacksCanBeAddedAfterFirstResolution(t *testing.T) {
	t.Parallel()

	c := newContainer()

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	var called bool

	c.Resolving("name", func(_ any, _ *container.App) {
		called = true
	})

	c.Make("name") //nolint:errcheck

	if !called {
		t.Fatal("callback added after first resolution should fire on next resolution")
	}
}

func TestCallbackFiringOrder(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var order []string

	c.BeforeResolvingAny(func(_ string, _ map[string]any, _ *container.App) {
		order = append(order, "global-before")
	})

	c.BeforeResolving("name", func(_ string, _ map[string]any, _ *container.App) {
		order = append(order, "specific-before")
	})

	c.ResolvingAny(func(_ any, _ *container.App) {
		order = append(order, "global-resolving")
	})

	c.Resolving("name", func(_ any, _ *container.App) {
		order = append(order, "specific-resolving")
	})

	c.AfterResolvingAny(func(_ any, _ *container.App) {
		order = append(order, "global-after")
	})

	c.AfterResolving("name", func(_ any, _ *container.App) {
		order = append(order, "specific-after")
	})

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Make("name") //nolint:errcheck

	expected := []string{
		"global-before",
		"specific-before",
		"global-resolving",
		"specific-resolving",
		"global-after",
		"specific-after",
	}

	if len(order) != len(expected) {
		t.Fatalf("expected %d callbacks, got %d: %v", len(expected), len(order), order)
	}

	for i, v := range expected {
		if order[i] != v {
			t.Fatalf("expected %s at position %d, got %s", v, i, order[i])
		}
	}
}

func TestResolvingCallbacksWithAlias(t *testing.T) {
	t.Parallel()

	c := newContainer()

	var called bool

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Alias("name", "shortName")

	c.Resolving("name", func(_ any, _ *container.App) {
		called = true
	})

	c.Make("shortName") //nolint:errcheck

	if !called {
		t.Fatal("resolving callback should fire when resolved via alias")
	}
}
