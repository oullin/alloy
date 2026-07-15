package container_test

import (
	"errors"
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestContainerIsPassedToResolvers(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("something", func(cc *container.App) (any, error) {
		return cc, nil
	}, false)

	v, _ := c.Make("something")

	if v != c {
		t.Fatal("expected container to be passed to factory")
	}
}

func TestNestedDependencyResolution(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("inner", func(_ *container.App) (any, error) {
		return "inner-value", nil
	}, false)

	c.Bind("outer", func(cc *container.App) (any, error) {
		inner, err := cc.Make("inner")

		if err != nil {
			return nil, err
		}

		return "outer:" + inner.(string), nil
	}, false)

	v, err := c.Make("outer")

	if err != nil {
		t.Fatal(err)
	}

	if v != "outer:inner-value" {
		t.Fatalf("expected outer:inner-value, got %v", v)
	}
}

func TestMakeWithParameters(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("greeting", func(cc *container.App) (any, error) {
		params := cc.Parameters()

		name, _ := params["name"]

		return "Hello, " + name.(string), nil
	}, false)

	v, err := c.MakeWith("greeting", map[string]any{"name": "Taylor"})

	if err != nil {
		t.Fatal(err)
	}

	if v != "Hello, Taylor" {
		t.Fatalf("expected Hello, Taylor, got %v", v)
	}
}

func TestSingletonNotCachedWithMakeParameters(t *testing.T) {
	t.Parallel()

	c := newContainer()

	count := 0

	c.Singleton("counter", func(_ *container.App) (any, error) {
		count++

		return count, nil
	})

	a, _ := c.MakeWith("counter", map[string]any{"force": true})
	b, _ := c.MakeWith("counter", map[string]any{"force": true})

	if a == b {
		t.Fatal("singleton with parameters should not be cached")
	}
}

func TestNestedParametersAreReset(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("inner", func(cc *container.App) (any, error) {
		params := cc.Parameters()

		if params != nil {
			return "has-params", nil
		}

		return "no-params", nil
	}, false)

	c.Bind("outer", func(cc *container.App) (any, error) {
		return cc.Make("inner")
	}, false)

	v, _ := c.MakeWith("outer", map[string]any{"key": "value"})

	if v != "no-params" {
		t.Fatalf("expected no-params for nested make, got %v", v)
	}
}

func TestBuildExecutesFactory(t *testing.T) {
	t.Parallel()

	c := newContainer()

	v, err := c.Build(func(_ *container.App) (any, error) {
		return "built", nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if v != "built" {
		t.Fatalf("expected built, got %v", v)
	}
}

func TestGetReturnsErrNotBoundForUnknown(t *testing.T) {
	t.Parallel()

	c := newContainer()

	_, err := c.Get("unknown")

	if !errors.Is(err, container.ErrNotBound) {
		t.Fatalf("expected ErrNotBound, got %v", err)
	}
}

func TestGetResolvesBinding(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	v, err := c.Get("name")

	if err != nil {
		t.Fatal(err)
	}

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestFactoryFuncReturnsNewInstanceEachTime(t *testing.T) {
	t.Parallel()

	c := newContainer()

	type stub struct{ id int }

	n := 0

	c.Bind("stub", func(_ *container.App) (any, error) {
		n++

		return &stub{id: n}, nil
	}, false)

	factory := c.FactoryFunc("stub")

	a, _ := factory()
	b, _ := factory()

	if a == b {
		t.Fatal("factory should return new instance each time")
	}
}

func TestMakeWithReturnsErrNotBound(t *testing.T) {
	t.Parallel()

	c := newContainer()

	_, err := c.Make("nonexistent")

	if !errors.Is(err, container.ErrNotBound) {
		t.Fatalf("expected ErrNotBound, got %v", err)
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	t.Parallel()

	c := newContainer()

	c.Bind("a", func(cc *container.App) (any, error) {
		return cc.Make("b")
	}, false)

	c.Bind("b", func(cc *container.App) (any, error) {
		return cc.Make("c")
	}, false)

	c.Bind("c", func(cc *container.App) (any, error) {
		return cc.Make("a")
	}, false)

	_, err := c.Make("a")

	if !errors.Is(err, container.ErrCircularDependency) {
		t.Fatalf("expected ErrCircularDependency, got %v", err)
	}
}

func TestMakeWithIsAliasForMake(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	v, err := c.MakeWith("name", nil)

	if err != nil {
		t.Fatal(err)
	}

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}
