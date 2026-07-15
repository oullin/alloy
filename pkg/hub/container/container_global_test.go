package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestGetSetInstance(t *testing.T) {
	t.Parallel()

	c := container.New()
	container.SetInstance(c)

	if container.GetInstance() != c {
		t.Fatal("expected same container")
	}

	container.SetInstance(nil)

	got := container.GetInstance()

	if got == nil {
		t.Fatal("GetInstance should create a new container if nil")
	}

	if got == c {
		t.Fatal("should be a new container")
	}

	// Cleanup.
	container.SetInstance(nil)
}
