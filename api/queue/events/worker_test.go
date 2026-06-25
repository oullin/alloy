package events_test

import (
	"testing"

	"github.com/oullin/alloy/queue"
	"github.com/oullin/alloy/queue/events"
)

func TestWorkerLifecycleEventFields(t *testing.T) {
	t.Parallel()

	starting := events.WorkerStarting{ConnectionName: "redis", Backend: "default", WorkerName: "w1"}

	if starting.ConnectionName != "redis" || starting.Backend != "default" || starting.WorkerName != "w1" {
		t.Errorf("WorkerStarting fields not preserved: %+v", starting)
	}

	stopping := events.WorkerStopping{Status: 1, WorkerName: "w1"}

	if stopping.Status != 1 || stopping.WorkerName != "w1" {
		t.Errorf("WorkerStopping fields not preserved: %+v", stopping)
	}

	pausing := events.WorkerPausing{ConnectionName: "redis", Backend: "default", WorkerName: "w1"}

	if pausing.ConnectionName != "redis" || pausing.Backend != "default" || pausing.WorkerName != "w1" {
		t.Errorf("WorkerPausing fields not preserved: %+v", pausing)
	}

	resuming := events.WorkerResuming{ConnectionName: "redis", Backend: "default", WorkerName: "w1"}

	if resuming.ConnectionName != "redis" || resuming.Backend != "default" || resuming.WorkerName != "w1" {
		t.Errorf("WorkerResuming fields not preserved: %+v", resuming)
	}
}

func TestQueuePackageReExportsWorkerEvents(t *testing.T) {
	t.Parallel()

	// The api/queue/events.go file re-aliases events.* under the
	// queue.* namespace for ergonomic callers. Verify each alias still
	// resolves to the same underlying struct.
	var (
		_ queue.WorkerStarting = events.WorkerStarting{}
		_ queue.WorkerStopping = events.WorkerStopping{}
		_ queue.WorkerPausing  = events.WorkerPausing{}
		_ queue.WorkerResuming = events.WorkerResuming{}
	)
}
