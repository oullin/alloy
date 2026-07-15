package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestClosureResolution(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	v, err := c.Make("name")

	if err != nil {
		t.Fatal(err)
	}

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestBindReturnsDifferentInstances(t *testing.T) {
	t.Parallel()

	c := newContainer()

	type stub struct{ id int }

	n := 0

	c.Bind("stub", func(_ *container.App) (any, error) {
		n++

		return &stub{id: n}, nil
	}, false)

	a, _ := c.Make("stub")
	b, _ := c.Make("stub")

	if a == b {
		t.Fatal("expected different instances for non-shared binding")
	}
}

func TestBindIfDoesNotRegisterIfAlreadyBound(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.BindIf("name", func(_ *container.App) (any, error) {
		return "Dayle", nil
	}, false)

	v, _ := c.Make("name")

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestBindIfDoesRegisterIfNotBound(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.BindIf("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	v, _ := c.Make("name")

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestSingletonReturnsSameInstance(t *testing.T) {
	t.Parallel()

	c := newContainer()

	type stub struct{ value string }

	c.Singleton("stub", func(_ *container.App) (any, error) {
		return &stub{value: "hello"}, nil
	})

	a, _ := c.Make("stub")
	b, _ := c.Make("stub")

	if a != b {
		t.Fatal("expected same instance for singleton")
	}
}

func TestSingletonIfDoesNotRegisterIfAlreadyBound(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Singleton("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	})

	c.SingletonIf("name", func(_ *container.App) (any, error) {
		return "Dayle", nil
	})

	v, _ := c.Make("name")

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestSingletonIfDoesRegisterIfNotBound(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.SingletonIf("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	})

	v, _ := c.Make("name")

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestScopedReturnsSameInstanceThenResets(t *testing.T) {
	t.Parallel()

	c := newContainer()

	type stub struct{ id int }

	n := 0

	c.Scoped("stub", func(_ *container.App) (any, error) {
		n++

		return &stub{id: n}, nil
	})

	a, _ := c.Make("stub")
	b, _ := c.Make("stub")

	if a != b {
		t.Fatal("expected same instance for scoped binding")
	}

	c.ForgetScopedInstances()

	d, _ := c.Make("stub")

	if a == d {
		t.Fatal("expected new instance after ForgetScopedInstances")
	}
}

func TestScopedIfDoesNotRegisterIfAlreadyBound(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Scoped("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	})

	c.ScopedIf("name", func(_ *container.App) (any, error) {
		return "Dayle", nil
	})

	v, _ := c.Make("name")

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestInstanceReturnsRegisteredValue(t *testing.T) {
	t.Parallel()

	c := newContainer()

	type stub struct{ Value string }

	obj := &stub{Value: "hello"}
	c.Instance("stub", obj)

	v, _ := c.Make("stub")

	if v != obj {
		t.Fatal("expected exact same instance")
	}
}

func TestInstanceReturnsTheInstance(t *testing.T) {
	t.Parallel()

	c := newContainer()

	type stub struct{}

	obj := &stub{}
	ret := c.Instance("stub", obj)

	if ret != obj {
		t.Fatal("Instance should return the instance")
	}
}

func TestBindingAnInstanceAsShared(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Instance("stub", "hello")

	if !c.IsShared("stub") {
		t.Fatal("instance should be shared")
	}
}

func TestBindingsCanBeOverridden(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Bind("name", func(_ *container.App) (any, error) {
		return "Dayle", nil
	}, false)

	v, _ := c.Make("name")

	if v != "Dayle" {
		t.Fatalf("expected Dayle, got %v", v)
	}
}

func TestContainerCanBindAnyWord(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("Taylor", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	if v, _ := c.Make("Taylor"); v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}
