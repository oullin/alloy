package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestApplication_Providers_ReturnsCopy(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()
	p := &fakeProvider{}
	app.Register(p)

	got := app.Providers()

	if len(got) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(got))
	}

	got[0] = nil

	if app.Providers()[0] == nil {
		t.Fatal("Providers() returned slice aliases internal state — must return a copy")
	}
}

func TestApplication_HasProviderAndProviderFor(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()
	p := &fakeProvider{provides: []string{"x", "y"}}
	app.Register(p)

	if !app.HasProvider("x") {
		t.Fatal("expected HasProvider(x) to be true")
	}

	if !app.HasProvider("y") {
		t.Fatal("expected HasProvider(y) to be true")
	}

	if app.HasProvider("z") {
		t.Fatal("expected HasProvider(z) to be false")
	}

	if got := app.ProviderFor("x"); got != p {
		t.Fatalf("expected ProviderFor(x) to return original provider")
	}

	if got := app.ProviderFor("z"); got != nil {
		t.Fatalf("expected ProviderFor(z) to be nil, got %v", got)
	}
}
