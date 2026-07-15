package container_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/container"
)

func TestApplication_DeferredProvider_NotRegisteredUntilResolved(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()

	dp := &deferredProvider{
		keys: []string{"deferred.svc"},
		register: func(_ *container.App) {
			app.Instance("deferred.svc", "value")
		},
	}

	app.Register(dp)

	if dp.registered {
		t.Fatal("expected deferred provider NOT to be registered immediately")
	}

	v, err := app.Make("deferred.svc")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v != "value" {
		t.Fatalf("expected %q, got %v", "value", v)
	}

	if !dp.registered {
		t.Fatal("expected deferred provider to be registered after first Make")
	}
}

func TestApplication_DeferredProvider_BootRunsAfterFlush(t *testing.T) {
	t.Parallel()

	app := container.NewApplication()

	dp := &deferredProvider{
		keys: []string{"deferred.boot"},
		register: func(_ *container.App) {
			app.Instance("deferred.boot", 1)
		},
	}

	app.Register(dp)
	app.Boot()

	if dp.bootCalls != 0 {
		t.Fatalf("expected Boot to be deferred, got %d calls", dp.bootCalls)
	}

	_, _ = app.Make("deferred.boot")

	if dp.bootCalls != 1 {
		t.Fatalf("expected one Boot call after flush, got %d", dp.bootCalls)
	}
}
