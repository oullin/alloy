package workflow_test

import (
	"errors"
	"slices"
	"sync"
	"testing"

	"hara.sh/alloy/workflow"
	"hara.sh/alloy/workflow/events"
)

// failingSetStore reads the marking from the subject's state but always fails
// the write, letting a test assert which events fire when SetMarking errors.
type failingSetStore struct {
	getter func(*Subscription) string
	setErr error
}

// raceStore serializes reads/writes to the shared subject's state through the
// same SingleState semantics; the machine lock is what makes the composite
// read-guard-write atomic, which this test exercises under -race.
type raceStore struct {
	subject *Subscription
}

func (s *failingSetStore) GetMarking(subject *Subscription, _ *workflow.Definition) (workflow.Marking, error) {
	place := s.getter(subject)

	if place == "" {
		return workflow.Marking{}, nil
	}

	return workflow.NewMarking(place), nil
}

func (s *failingSetStore) SetMarking(_ *Subscription, _ workflow.Marking, _ *workflow.Definition, _ map[string]any) error {
	return s.setErr
}

func TestApply_FailedWriteEventsDoNotFire(t *testing.T) {
	def := subscriptionDef(t)
	dispatcher := events.NewDispatcher[*Subscription]()

	fired := map[string]bool{}

	record := func(name string) events.Listener[*Subscription] {
		return func(events.Event[*Subscription]) { fired[name] = true }
	}

	for _, name := range []string{
		workflow.EventNameLeave("subscription"),
		workflow.EventNameTransition("subscription"),
		workflow.EventNameEnter("subscription"),
		workflow.EventNameEntered("subscription"),
		workflow.EventNameCompleted("subscription"),
		workflow.EventNameAnnounce("subscription"),
	} {
		dispatcher.On(name, record(name))
	}

	// The guard must still fire — it is part of the pre-write decision.
	var guarded bool

	dispatcher.On(workflow.EventNameGuard("subscription"), func(events.Event[*Subscription]) { guarded = true })

	writeErr := errors.New("store unavailable")
	store := &failingSetStore{
		getter: func(s *Subscription) string { return s.State },
		setErr: writeErr,
	}

	sm, err := workflow.NewStateMachine("subscription", def, store, dispatcher)

	if err != nil {
		t.Fatalf("new state machine: %v", err)
	}

	sub := &Subscription{ID: "s1", State: "trial"}

	if _, err := sm.Apply(sub, "activate", nil); !errors.Is(err, writeErr) {
		t.Fatalf("expected write error, got %v", err)
	}

	if !guarded {
		t.Fatal("guard event should still fire before the write")
	}

	for name, ok := range fired {
		if ok {
			t.Errorf("event %q must not fire when SetMarking fails", name)
		}
	}

	if fired[workflow.EventNameEnter("subscription")] || fired[workflow.EventNameEntered("subscription")] {
		t.Fatal("enter/entered events must not fire on a failed write")
	}
}

func TestApply_ConcurrentConflictingTransitions(t *testing.T) {
	// A single token in "start" enables two mutually exclusive transitions.
	// Two goroutines racing to consume it must produce exactly one winner and
	// one guard/conflict error — never a double-consume.
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("start").
		AddPlace("left").
		AddPlace("right").
		SetInitialPlaces("start").
		AddTransition("goLeft", []string{"start"}, []string{"left"}).
		AddTransition("goRight", []string{"start"}, []string{"right"}).
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	sub := &Subscription{ID: "race", State: "start"}

	sm, err := workflow.NewStateMachine("fork", def, &raceStore{subject: sub}, nil)

	if err != nil {
		t.Fatalf("new state machine: %v", err)
	}

	var wg sync.WaitGroup

	results := make([]error, 2)
	transitions := []string{"goLeft", "goRight"}

	start := make(chan struct{})

	for i := range transitions {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			<-start

			_, results[idx] = sm.Apply(sub, transitions[idx], nil)
		}(i)
	}

	close(start)
	wg.Wait()

	var successes, failures int

	for _, err := range results {
		if err == nil {
			successes++

			continue
		}

		failures++

		var te *workflow.TransitionError

		if !errors.As(err, &te) {
			t.Fatalf("expected *TransitionError for the loser, got %T: %v", err, err)
		}
	}

	if successes != 1 || failures != 1 {
		t.Fatalf("expected exactly one winner and one failure, got %d successes / %d failures", successes, failures)
	}

	if sub.State != "left" && sub.State != "right" {
		t.Fatalf("subject must land in exactly one target place, got %q", sub.State)
	}
}

func (s *raceStore) GetMarking(subject *Subscription, _ *workflow.Definition) (workflow.Marking, error) {
	if subject.State == "" {
		return workflow.Marking{}, nil
	}

	return workflow.NewMarking(subject.State), nil
}

func (s *raceStore) SetMarking(subject *Subscription, marking workflow.Marking, _ *workflow.Definition, _ map[string]any) error {
	places := marking.ActivePlaces()

	if len(places) == 1 {
		subject.State = places[0]
	} else {
		subject.State = ""
	}

	return nil
}

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
