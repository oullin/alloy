package workflow_test

import (
	"slices"
	"testing"

	"alloy.dev/foundation/workflow"
	"alloy.dev/foundation/workflow/events"
)

func TestApply_EventOrder(t *testing.T) {
	def := subscriptionDef(t)
	dispatcher := events.NewDispatcher[*Subscription]()

	var order []string

	record := func(name string) events.Listener[*Subscription] {
		return func(events.Event[*Subscription]) { order = append(order, name) }
	}

	dispatcher.On(workflow.EventNameGuard("subscription"), record("guard"))
	dispatcher.On(workflow.EventNameLeave("subscription"), record("leave"))
	dispatcher.On(workflow.EventNameTransition("subscription"), record("transition"))
	dispatcher.On(workflow.EventNameEnter("subscription"), record("enter"))
	dispatcher.On(workflow.EventNameEntered("subscription"), record("entered"))
	dispatcher.On(workflow.EventNameCompleted("subscription"), record("completed"))
	dispatcher.On(workflow.EventNameAnnounce("subscription"), record("announce"))

	sm, _ := workflow.NewStateMachine("subscription", def, subscriptionStore(), dispatcher)

	if _, err := sm.Apply(&Subscription{State: "trial"}, "activate", nil); err != nil {
		t.Fatalf("apply activate: %v", err)
	}

	// The trailing guard fires while computing the enabled transitions that
	// are included in the announce event.
	want := []string{"guard", "leave", "transition", "enter", "entered", "completed", "guard", "announce"}

	if !slices.Equal(order, want) {
		t.Fatalf("event order = %v, want %v", order, want)
	}
}

func TestApply_NamedEventVariantsFire(t *testing.T) {
	def := subscriptionDef(t)
	dispatcher := events.NewDispatcher[*Subscription]()

	fired := map[string]bool{}

	record := func(name string) events.Listener[*Subscription] {
		return func(events.Event[*Subscription]) { fired[name] = true }
	}

	names := []string{
		workflow.EventNameGuardNamed("subscription", "activate"),
		workflow.EventNameLeavePlace("subscription", "trial"),
		workflow.EventNameTransitionNamed("subscription", "activate"),
		workflow.EventNameEnterPlace("subscription", "active"),
		workflow.EventNameEnteredPlace("subscription", "active"),
		workflow.EventNameCompletedNamed("subscription", "activate"),
		workflow.EventNameAnnounceNamed("subscription", "activate"),
	}

	for _, name := range names {
		dispatcher.On(name, record(name))
	}

	sm, _ := workflow.NewStateMachine("subscription", def, subscriptionStore(), dispatcher)

	if _, err := sm.Apply(&Subscription{State: "trial"}, "activate", nil); err != nil {
		t.Fatalf("apply activate: %v", err)
	}

	for _, name := range names {
		if !fired[name] {
			t.Errorf("expected event %q to fire", name)
		}
	}
}

func TestApply_EventPayloads(t *testing.T) {
	def := subscriptionDef(t)
	dispatcher := events.NewDispatcher[*Subscription]()

	var leave *events.LeaveEvent[*Subscription]
	var enter *events.EnterEvent[*Subscription]
	var entered *events.EnteredEvent[*Subscription]
	var completed *events.CompletedEvent[*Subscription]
	var announce *events.AnnounceEvent[*Subscription]

	dispatcher.On(workflow.EventNameLeave("subscription"), func(e events.Event[*Subscription]) {
		leave = e.(*events.LeaveEvent[*Subscription])
	})
	dispatcher.On(workflow.EventNameEnter("subscription"), func(e events.Event[*Subscription]) {
		enter = e.(*events.EnterEvent[*Subscription])
	})
	dispatcher.On(workflow.EventNameEntered("subscription"), func(e events.Event[*Subscription]) {
		entered = e.(*events.EnteredEvent[*Subscription])
	})
	dispatcher.On(workflow.EventNameCompleted("subscription"), func(e events.Event[*Subscription]) {
		completed = e.(*events.CompletedEvent[*Subscription])
	})
	dispatcher.On(workflow.EventNameAnnounce("subscription"), func(e events.Event[*Subscription]) {
		announce = e.(*events.AnnounceEvent[*Subscription])
	})

	sm, _ := workflow.NewStateMachine("subscription", def, subscriptionStore(), dispatcher)

	sub := &Subscription{ID: "s1", State: "trial"}

	if _, err := sm.Apply(sub, "activate", map[string]any{"actor": "billing"}); err != nil {
		t.Fatalf("apply activate: %v", err)
	}

	if leave == nil || leave.Place != "trial" {
		t.Fatalf("leave event = %#v", leave)
	}

	if leave.Marking()["trial"] != 1 {
		t.Fatalf("leave should see the pre-transition marking, got %v", leave.Marking())
	}

	if enter == nil || enter.Place != "active" {
		t.Fatalf("enter event = %#v", enter)
	}

	if marking := enter.Marking(); marking["trial"] != 0 || marking["active"] != 0 {
		t.Fatalf("enter should see the in-between marking, got %v", marking)
	}

	if entered == nil || entered.Place != "active" || entered.Marking()["active"] != 1 {
		t.Fatalf("entered event = %#v", entered)
	}

	if completed == nil || completed.Transition().Name != "activate" {
		t.Fatalf("completed event = %#v", completed)
	}

	if completed.Subject() != sub {
		t.Fatal("completed event should carry the subject")
	}

	if completed.Context()["actor"] != "billing" {
		t.Fatalf("completed context = %v", completed.Context())
	}

	if announce == nil || len(announce.Enabled) != 1 || announce.Enabled[0].Name != "cancel" {
		t.Fatalf("announce event should list [cancel], got %#v", announce)
	}
}

func TestApply_ANDSplitEmitsEnterEventPerPlace(t *testing.T) {
	def := orderDef(t)
	dispatcher := events.NewDispatcher[*Order]()

	var enteredPlaces []string

	dispatcher.On(workflow.EventNameEnter("order"), func(e events.Event[*Order]) {
		enteredPlaces = append(enteredPlaces, e.(*events.EnterEvent[*Order]).Place)
	})

	machine, _ := workflow.New("order", def, orderStore(), dispatcher)

	if _, err := machine.Apply(&Order{ID: "o1"}, "start", nil); err != nil {
		t.Fatalf("apply start: %v", err)
	}

	want := []string{"production", "quality"}

	if !slices.Equal(enteredPlaces, want) {
		t.Fatalf("enter places = %v, want %v", enteredPlaces, want)
	}
}
