package container_test

import (
	"errors"
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestAliasResolvesToAbstract(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Alias("name", "shortName")

	v, _ := c.Make("shortName")

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestAliasChainResolution(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Alias("name", "alias1")
	c.Alias("alias1", "alias2")

	v, _ := c.Make("alias2")

	if v != "Taylor" {
		t.Fatalf("expected Taylor, got %v", v)
	}
}

func TestGetAliasResolvesChain(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Alias("name", "alias1")
	c.Alias("alias1", "alias2")

	if c.GetAlias("alias2") != "name" {
		t.Fatalf("expected name, got %s", c.GetAlias("alias2"))
	}
}

func TestGetAliasReturnsAbstractIfNotAlias(t *testing.T) {
	t.Parallel()

	c := newContainer()

	if c.GetAlias("notanalias") != "notanalias" {
		t.Fatal("expected same name back")
	}
}

func TestIsAlias(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Alias("name", "shortName")

	if !c.IsAlias("shortName") {
		t.Fatal("expected shortName to be an alias")
	}

	if c.IsAlias("name") {
		t.Fatal("name is not an alias")
	}
}

func TestSelfAliasErrors(t *testing.T) {
	t.Parallel()

	c := newContainer()

	if err := c.Alias("name", "name"); !errors.Is(err, container.ErrSelfAlias) {
		t.Fatalf("Alias error = %v, want ErrSelfAlias", err)
	}
}

func TestBoundWithAlias(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Alias("name", "shortName")

	if !c.Bound("shortName") {
		t.Fatal("alias should report as bound")
	}
}

func TestResolvedResolvesAlias(t *testing.T) {
	t.Parallel()

	c := newContainer()
	c.Bind("name", func(_ *container.App) (any, error) {
		return "Taylor", nil
	}, false)

	c.Alias("name", "shortName")
	c.Make("name") //nolint:errcheck

	if !c.Resolved("shortName") {
		t.Fatal("alias should show as resolved")
	}
}
