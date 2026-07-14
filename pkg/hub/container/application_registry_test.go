package container_test

import (
	"errors"
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

type sample struct{ Value string }

func TestSetAppAndApp(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	container.SetApp(nil)

	application := container.NewApplication()
	container.SetApp(application)

	got, err := container.Global()

	if err != nil {
		t.Fatalf("Global() returned error: %v", err)
	}

	if got != application {
		t.Fatalf("Global() = %p, want %p", got, application)
	}
}

func TestGlobalErrorsWhenUnset(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	container.SetApp(nil)

	_, err := container.Global()

	if !errors.Is(err, container.ErrNoApplication) {
		t.Fatalf("Global() error = %v, want ErrNoApplication", err)
	}
}

func TestMustGlobalPanicsWhenUnset(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	container.SetApp(nil)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when app is unset")
		}
	}()

	container.MustGlobal()
}

func TestHasApp(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	container.SetApp(nil)

	if container.HasApp() {
		t.Fatal("HasApp should be false after SetApp(nil)")
	}

	container.SetApp(container.NewApplication())

	if !container.HasApp() {
		t.Fatal("HasApp should be true after installing an app")
	}
}

func TestMakeErrorsWhenUnset(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	container.SetApp(nil)

	_, err := container.Make("anything")

	if !errors.Is(err, container.ErrNoApplication) {
		t.Fatalf("Make() error = %v, want ErrNoApplication", err)
	}
}

func TestMustMake(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	application := container.NewApplication()
	application.Instance("answer", 42)
	container.SetApp(application)

	if got := container.MustMake("answer"); got != 42 {
		t.Fatalf("MustMake = %v, want 42", got)
	}
}

func TestMustMakePanicsWhenMissing(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	container.SetApp(container.NewApplication())

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for missing binding")
		}
	}()

	container.MustMake("nope")
}

func TestResolve(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	application := container.NewApplication()
	application.Instance("sample", &sample{Value: "ok"})
	container.SetApp(application)

	got, err := container.Resolve[*sample]("sample")

	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	if got.Value != "ok" {
		t.Fatalf("Resolve returned %+v", got)
	}
}

func TestResolveErrorsOnWrongType(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	application := container.NewApplication()
	application.Instance("sample", 123)
	container.SetApp(application)

	_, err := container.Resolve[*sample]("sample")

	if err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestResolveErrorsWhenMissing(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	container.SetApp(container.NewApplication())

	_, err := container.Resolve[*sample]("nope")

	if err == nil {
		t.Fatal("expected error for missing binding")
	}
}

func TestResolveErrorsWhenUnset(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	container.SetApp(nil)

	_, err := container.Resolve[*sample]("anything")

	if !errors.Is(err, container.ErrNoApplication) {
		t.Fatalf("Resolve error = %v, want ErrNoApplication", err)
	}
}

func TestMustResolve(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	application := container.NewApplication()
	application.Instance("sample", &sample{Value: "ok"})
	container.SetApp(application)

	got := container.MustResolve[*sample]("sample")

	if got.Value != "ok" {
		t.Fatalf("MustResolve returned %+v", got)
	}
}

func TestMustResolvePanicsOnWrongType(t *testing.T) {
	t.Cleanup(func() {
		container.SetApp(nil)
	})

	application := container.NewApplication()
	application.Instance("sample", 123)
	container.SetApp(application)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for wrong type")
		}
	}()

	container.MustResolve[*sample]("sample")
}
