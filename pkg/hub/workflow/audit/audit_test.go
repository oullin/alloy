package audit_test

import (
	"testing"

	"hara.sh/alloy/workflow"
	"hara.sh/alloy/workflow/audit"
	"hara.sh/alloy/workflow/events"
)

func completedEvent(subject string, transition string) *events.CompletedEvent[string] {
	return &events.CompletedEvent[string]{
		Base: events.Base[string]{
			Machine:    "orders",
			SubjectVal: subject,
			Step:       events.Transition{Name: transition, From: []string{"a"}, To: []string{"b"}},
			Tokens:     map[string]int{"b": 1},
			Ctx:        map[string]any{"actor": "ops"},
		},
	}
}

func TestTrailRecordsCompletedEvents(t *testing.T) {
	dispatcher := events.NewDispatcher[string]()

	trail := &audit.Trail[string]{}
	trail.Attach("orders", dispatcher)

	dispatcher.Dispatch(workflow.EventNameCompleted("orders"), completedEvent("o1", "go"))
	dispatcher.Dispatch(workflow.EventNameCompleted("orders"), completedEvent("o2", "go"))

	if len(trail.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(trail.Entries))
	}

	entry := trail.Entries[0]

	if entry.Machine != "orders" || entry.Transition != "go" || entry.Subject != "o1" {
		t.Fatalf("entry = %#v", entry)
	}

	if !entry.Marking.Has("b") {
		t.Fatalf("expected marking with b, got %v", entry.Marking.ActivePlaces())
	}

	if entry.Context["actor"] != "ops" {
		t.Fatalf("expected recorded context, got %v", entry.Context)
	}
}

func TestTrailIgnoresOtherEventTypes(t *testing.T) {
	dispatcher := events.NewDispatcher[string]()

	trail := &audit.Trail[string]{}
	trail.Attach("orders", dispatcher)

	dispatcher.Dispatch(workflow.EventNameCompleted("orders"), &events.TransitionEvent[string]{})

	if len(trail.Entries) != 0 {
		t.Fatalf("expected no entries, got %d", len(trail.Entries))
	}
}

func TestTrailAttachIsNilSafe(t *testing.T) {
	var trail *audit.Trail[string]

	trail.Attach("orders", events.NewDispatcher[string]())

	(&audit.Trail[string]{}).Attach("orders", nil)
}
