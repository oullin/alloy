package events_test

import (
	"slices"
	"testing"

	"hara.sh/alloy/workflow/events"
)

func baseEvent() events.Base[string] {
	return events.Base[string]{
		Machine:    "orders",
		SubjectVal: "o1",
		Step:       events.Transition{Name: "go", From: []string{"a"}, To: []string{"b"}},
		Tokens:     map[string]int{"a": 1},
		Ctx:        map[string]any{"actor": "ops"},
	}
}

func TestDispatcherRunsListenersInOrder(t *testing.T) {
	dispatcher := events.NewDispatcher[string]()

	var order []string

	dispatcher.On("orders.go", func(events.Event[string]) { order = append(order, "first") })
	dispatcher.On("orders.go", func(events.Event[string]) { order = append(order, "second") })

	dispatcher.Dispatch("orders.go", &events.TransitionEvent[string]{Base: baseEvent()})

	if !slices.Equal(order, []string{"first", "second"}) {
		t.Fatalf("listener order = %v", order)
	}
}

func TestDispatcherIgnoresUnknownNames(t *testing.T) {
	dispatcher := events.NewDispatcher[string]()

	dispatcher.Dispatch("nobody-listens", &events.TransitionEvent[string]{Base: baseEvent()})
}

func TestDispatchOnNilDispatcherIsSafe(t *testing.T) {
	var dispatcher *events.Dispatcher[string]

	dispatcher.Dispatch("orders.go", &events.TransitionEvent[string]{Base: baseEvent()})
}

func TestOnGuardOnlyReceivesGuardEvents(t *testing.T) {
	dispatcher := events.NewDispatcher[string]()

	var calls int

	dispatcher.OnGuard("orders.guard", func(*events.GuardEvent[string]) { calls++ })

	dispatcher.Dispatch("orders.guard", &events.TransitionEvent[string]{Base: baseEvent()})

	if calls != 0 {
		t.Fatal("non-guard events should be ignored by OnGuard listeners")
	}

	dispatcher.Dispatch("orders.guard", &events.GuardEvent[string]{Base: baseEvent()})

	if calls != 1 {
		t.Fatalf("expected one guard call, got %d", calls)
	}
}

func TestGuardEventSetBlocked(t *testing.T) {
	guard := &events.GuardEvent[string]{Base: baseEvent()}

	if guard.Blocked() {
		t.Fatal("new guard event should not be blocked")
	}

	guard.SetBlocked(true, "billing not configured")

	if !guard.Blocked() {
		t.Fatal("guard should be blocked")
	}

	blockers := guard.Blockers()

	if len(blockers) != 1 || blockers[0].Message != "billing not configured" {
		t.Fatalf("blockers = %#v", blockers)
	}

	guard.SetBlocked(true, "")

	if len(guard.Blockers()) != 1 {
		t.Fatal("empty message should not add a blocker")
	}

	guard.SetBlocked(false, "ignored")

	if guard.Blocked() {
		t.Fatal("guard should be unblocked")
	}
}

func TestGuardEventAddTransitionBlocker(t *testing.T) {
	guard := &events.GuardEvent[string]{Base: baseEvent()}

	guard.AddTransitionBlocker(events.TransitionBlocker{Message: "no stock", Code: "stock"})

	if !guard.Blocked() {
		t.Fatal("adding a blocker should mark the guard blocked")
	}

	blockers := guard.Blockers()

	if len(blockers) != 1 || blockers[0].Code != "stock" {
		t.Fatalf("blockers = %#v", blockers)
	}

	blockers[0].Message = "mutated"

	if guard.Blockers()[0].Message != "no stock" {
		t.Fatal("Blockers() should return a copy")
	}
}

func TestBaseAccessors(t *testing.T) {
	base := baseEvent()

	if base.WorkflowName() != "orders" {
		t.Fatalf("WorkflowName() = %q", base.WorkflowName())
	}

	if base.Subject() != "o1" {
		t.Fatalf("Subject() = %q", base.Subject())
	}

	step := base.Transition()

	if step.Name != "go" || !slices.Equal(step.From, []string{"a"}) || !slices.Equal(step.To, []string{"b"}) {
		t.Fatalf("Transition() = %#v", step)
	}

	marking := base.Marking()

	if marking["a"] != 1 {
		t.Fatalf("Marking() = %v", marking)
	}

	marking["a"] = 99

	if base.Marking()["a"] != 1 {
		t.Fatal("Marking() should return a copy")
	}

	ctx := base.Context()

	if ctx["actor"] != "ops" {
		t.Fatalf("Context() = %v", ctx)
	}

	ctx["actor"] = "mutated"

	if base.Context()["actor"] != "ops" {
		t.Fatal("Context() should return a copy")
	}
}
