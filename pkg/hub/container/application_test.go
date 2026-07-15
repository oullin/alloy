package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestNewApplication_HasUsableContainer(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()

	if app.App == nil {
		t.Fatal("expected embedded *App to be non-nil")
	}

	app.Instance("answer", 42)

	v, err := app.Make("answer")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}
