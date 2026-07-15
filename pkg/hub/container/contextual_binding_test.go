package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestContextualBindingResolvesCorrectImplementation(t *testing.T) {
	t.Parallel()

	c := newContainer()

	c.Bind("logger", func(_ *container.App) (any, error) {
		return "default-logger", nil
	}, false)

	c.When("service-a").Needs("logger").Give(container.Factory(func(_ *container.App) (any, error) {
		return "special-logger", nil
	}))

	c.Bind("service-a", func(cc *container.App) (any, error) {
		return cc.Make("logger")
	}, false)

	c.Bind("service-b", func(cc *container.App) (any, error) {
		return cc.Make("logger")
	}, false)

	a, _ := c.Make("service-a")
	b, _ := c.Make("service-b")

	if a != "special-logger" {
		t.Fatalf("expected special-logger for service-a, got %v", a)
	}

	if b != "default-logger" {
		t.Fatalf("expected default-logger for service-b, got %v", b)
	}
}

func TestContextualBindingWithRawValue(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.When("service").Needs("$apiKey").Give("secret-key-123")

	c.Bind("service", func(cc *container.App) (any, error) {
		return cc.Make("$apiKey")
	}, false)

	v, _ := c.Make("service")

	if v != "secret-key-123" {
		t.Fatalf("expected secret-key-123, got %v", v)
	}
}

func TestContextualBindingForMultipleConcretes(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("logger", func(_ *container.App) (any, error) {
		return "default-logger", nil
	}, false)

	c.When("service-a", "service-b").Needs("logger").Give(container.Factory(func(_ *container.App) (any, error) {
		return "shared-logger", nil
	}))

	c.Bind("service-a", func(cc *container.App) (any, error) {
		return cc.Make("logger")
	}, false)

	c.Bind("service-b", func(cc *container.App) (any, error) {
		return cc.Make("logger")
	}, false)

	c.Bind("service-c", func(cc *container.App) (any, error) {
		return cc.Make("logger")
	}, false)

	a, _ := c.Make("service-a")
	b, _ := c.Make("service-b")
	d, _ := c.Make("service-c")

	if a != "shared-logger" {
		t.Fatalf("expected shared-logger for service-a, got %v", a)
	}

	if b != "shared-logger" {
		t.Fatalf("expected shared-logger for service-b, got %v", b)
	}

	if d != "default-logger" {
		t.Fatalf("expected default-logger for service-c, got %v", d)
	}
}

func TestContextualBindingDoesNotOverrideNonContextualResolution(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("logger", func(_ *container.App) (any, error) {
		return "default-logger", nil
	}, false)

	c.When("service-a").Needs("logger").Give(container.Factory(func(_ *container.App) (any, error) {
		return "special-logger", nil
	}))

	v, _ := c.Make("logger")

	if v != "default-logger" {
		t.Fatalf("expected default-logger, got %v", v)
	}
}

func TestContextualBindingWorksForExistingInstancedBindings(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("logger", "instanced-logger")

	c.When("service").Needs("logger").Give(container.Factory(func(_ *container.App) (any, error) {
		return "contextual-logger", nil
	}))

	c.Bind("service", func(cc *container.App) (any, error) {
		return cc.Make("logger")
	}, false)

	v, _ := c.Make("service")

	if v != "contextual-logger" {
		t.Fatalf("expected contextual-logger, got %v", v)
	}
}

func TestContextualBindingWorksWithAliasedTargets(t *testing.T) {
	t.Parallel()

	c := newContainer()

	c.Bind("logger", func(_ *container.App) (any, error) {
		return "default", nil
	}, false)

	c.Alias("logger", "log")

	c.When("service").Needs("log").Give(container.Factory(func(_ *container.App) (any, error) {
		return "contextual", nil
	}))

	c.Bind("service", func(cc *container.App) (any, error) {
		return cc.Make("log")
	}, false)

	v, _ := c.Make("service")

	if v != "contextual" {
		t.Fatalf("expected contextual, got %v", v)
	}
}

func TestContextualBindingNotRecreatedUnnecessarily(t *testing.T) {
	t.Parallel()

	c := newContainer()

	count := 0

	c.When("service").Needs("dep").Give(container.Factory(func(_ *container.App) (any, error) {
		count++

		return count, nil
	}))

	c.Bind("service", func(cc *container.App) (any, error) {
		return cc.Make("dep")
	}, false)

	a, _ := c.Make("service")
	b, _ := c.Make("service")

	// Each resolution should call the factory.
	if a == b {
		t.Fatal("expected different instances for non-shared contextual binding")
	}
}

func TestAddContextualBindingDirectly(t *testing.T) {
	t.Parallel()

	c := newContainer()

	c.AddContextualBinding("service", "dep", container.Factory(func(_ *container.App) (any, error) {
		return "direct-contextual", nil
	}))

	c.Bind("service", func(cc *container.App) (any, error) {
		return cc.Make("dep")
	}, false)

	v, _ := c.Make("service")

	if v != "direct-contextual" {
		t.Fatalf("expected direct-contextual, got %v", v)
	}
}
