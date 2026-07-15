package workflow_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/oullin/alloy/pkg/hub/workflow"
	"github.com/oullin/alloy/pkg/hub/workflow/events"
	"github.com/oullin/alloy/pkg/hub/workflow/store"
)

type captureSink struct {
	debugs []string
	infos  []string
	warns  []string
	errs   []string
}

// failingStore wraps a MarkingStore and forces errors on selected operations.
type failingStore struct {
	inner        workflow.MarkingStore[*Subscription]
	failGetAfter int // fail GetMarking once this many calls happened (0 = never)
	failSet      bool
	gets         int
}

// Order is the multi-token Petri-net test subject.
type Order struct {
	ID     string
	Places []string
}

func (s *captureSink) Debug(msg string, args ...any) { s.debugs = append(s.debugs, msg) }
func (s *captureSink) Info(msg string, args ...any)  { s.infos = append(s.infos, msg) }
func (s *captureSink) Warn(msg string, args ...any)  { s.warns = append(s.warns, msg) }
func (s *captureSink) Error(msg string, args ...any) { s.errs = append(s.errs, msg) }

func (s *failingStore) GetMarking(subject *Subscription, definition *workflow.Definition) (workflow.Marking, error) {
	s.gets++

	if s.failGetAfter > 0 && s.gets >= s.failGetAfter {
		return workflow.Marking{}, fmt.Errorf("get marking failed")
	}

	return s.inner.GetMarking(subject, definition)
}

func (s *failingStore) SetMarking(subject *Subscription, marking workflow.Marking, definition *workflow.Definition, context map[string]any) error {
	if s.failSet {
		return fmt.Errorf("set marking failed")
	}

	return s.inner.SetMarking(subject, marking, definition, context)
}

func orderDef(t *testing.T) *workflow.Definition {
	t.Helper()

	def, err := workflow.NewDefinitionBuilder().
		AddPlace("new").
		AddPlace("production").
		AddPlace("quality").
		AddPlace("shipped").
		SetInitialPlaces("new").
		AddTransition("start", []string{"new"}, []string{"production", "quality"}).
		AddTransition("finish", []string{"production", "quality"}, []string{"shipped"}).
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	return def
}

func orderStore() *store.MultiState[*Order] {
	return &store.MultiState[*Order]{
		Getter: func(o *Order) []string { return o.Places },
		Setter: func(o *Order, places []string) { o.Places = places },
	}
}

func TestNew_RequiresNameDefinitionAndStore(t *testing.T) {
	def := subscriptionDef(t)

	if _, err := workflow.New("", def, subscriptionStore(), nil); err == nil {
		t.Fatal("expected error for empty name")
	}

	if _, err := workflow.New[*Subscription]("subscription", nil, subscriptionStore(), nil); err == nil {
		t.Fatal("expected error for nil definition")
	}

	if _, err := workflow.New[*Subscription]("subscription", def, nil, nil); err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_RejectsInvalidDefinition(t *testing.T) {
	def := &workflow.Definition{
		Places:         []string{"a", "orphan"},
		InitialMarking: workflow.Marking{Places: map[string]int{"a": 1}},
	}

	if _, err := workflow.New("subscription", def, subscriptionStore(), nil); err == nil {
		t.Fatal("expected error for invalid definition")
	}
}

func TestNewStateMachine_RequiresDefinition(t *testing.T) {
	if _, err := workflow.NewStateMachine[*Subscription]("subscription", nil, subscriptionStore(), nil); err == nil {
		t.Fatal("expected error for nil definition")
	}
}

func TestNewStateMachine_RequiresSingleInitialPlace(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("b").
		SetInitialPlaces("a", "b").
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	if _, err := workflow.NewStateMachine("wf", def, subscriptionStore(), nil); err == nil {
		t.Fatal("expected error for multiple initial places")
	}
}

func TestNewStateMachine_RequiresSingleTargetTransitions(t *testing.T) {
	def := orderDef(t)

	if _, err := workflow.NewStateMachine("order", def, subscriptionStore(), nil); err == nil {
		t.Fatal("expected error for multi-target transition")
	}
}

func TestMachine_Accessors(t *testing.T) {
	def := subscriptionDef(t)
	dispatcher := events.NewDispatcher[*Subscription]()

	sm, err := workflow.New("subscription", def, subscriptionStore(), dispatcher)

	if err != nil {
		t.Fatalf("new machine: %v", err)
	}

	if sm.Name() != "subscription" {
		t.Fatalf("Name() = %q", sm.Name())
	}

	cloned := sm.Definition()

	if cloned == nil || len(cloned.Places) != 3 {
		t.Fatalf("Definition() returned unexpected value: %#v", cloned)
	}

	cloned.Places[0] = "mutated"

	if sm.Definition().Places[0] == "mutated" {
		t.Fatal("Definition() should return a clone")
	}

	if sm.MetadataStore() == nil {
		t.Fatal("MetadataStore() should not be nil")
	}

	if sm.EventDispatcher() != dispatcher {
		t.Fatal("EventDispatcher() should return the configured dispatcher")
	}
}

func TestMachine_LoggerReceivesLifecycleLogs(t *testing.T) {
	def := subscriptionDef(t)
	sm, _ := workflow.NewStateMachine("subscription", def, subscriptionStore(), nil)

	sink := &captureSink{}
	sm.SetLogger(sink)

	sub := &Subscription{ID: "s1", State: "trial"}

	if _, err := sm.Apply(sub, "activate", nil); err != nil {
		t.Fatalf("apply activate: %v", err)
	}

	if len(sink.debugs) == 0 {
		t.Fatal("expected a debug log for the applied transition")
	}

	if len(sink.infos) == 0 {
		t.Fatal("expected an info log for the applied transition")
	}

	if _, err := sm.Apply(sub, "bogus", nil); err == nil {
		t.Fatal("expected error for unknown transition")
	}

	if len(sink.errs) == 0 {
		t.Fatal("expected an error log for the failed transition")
	}
}

func TestMachine_GetMarkingFallsBackToInitialMarking(t *testing.T) {
	def := subscriptionDef(t)
	sm, _ := workflow.NewStateMachine("subscription", def, subscriptionStore(), nil)

	marking, err := sm.GetMarking(&Subscription{State: ""})

	if err != nil {
		t.Fatalf("get marking: %v", err)
	}

	if !marking.Has("trial") {
		t.Fatalf("expected initial marking, got %v", marking.ActivePlaces())
	}
}

func TestMachine_GetMarkingPropagatesStoreError(t *testing.T) {
	def := subscriptionDef(t)
	broken := &store.SingleState[*Subscription]{}

	sm, _ := workflow.NewStateMachine("subscription", def, broken, nil)

	if _, err := sm.GetMarking(&Subscription{}); err == nil {
		t.Fatal("expected error from broken store")
	}

	if sm.Can(&Subscription{}, "activate") {
		t.Fatal("Can should be false when the marking cannot be read")
	}

	if _, err := sm.Apply(&Subscription{}, "activate", nil); err == nil {
		t.Fatal("expected apply to fail when the marking cannot be read")
	}
}

func TestMachine_ApplyFailsWhenSetMarkingFails(t *testing.T) {
	def := subscriptionDef(t)
	failing := &failingStore{inner: subscriptionStore(), failSet: true}

	sm, _ := workflow.NewStateMachine("subscription", def, failing, nil)

	sub := &Subscription{State: "trial"}

	if _, err := sm.Apply(sub, "activate", nil); err == nil {
		t.Fatal("expected error when SetMarking fails")
	}

	if sub.State != "trial" {
		t.Fatalf("subject state should be unchanged, got %q", sub.State)
	}
}

func TestMachine_ApplyFailsWhenCompletionMarkingLookupFails(t *testing.T) {
	def := subscriptionDef(t)
	failing := &failingStore{inner: subscriptionStore(), failGetAfter: 2}

	sm, _ := workflow.NewStateMachine("subscription", def, failing, nil)

	marking, err := sm.Apply(&Subscription{State: "trial"}, "activate", nil)

	if err == nil {
		t.Fatal("expected error when the post-transition marking lookup fails")
	}

	// The write itself committed before the completion dispatch failed, so the
	// returned marking must reflect the new state rather than a zero Marking.
	if !marking.Has("active") {
		t.Fatalf("expected committed marking with 'active' place alongside the error, got %v", marking.ActivePlaces())
	}
}

func TestPetriNet_ANDSplitAndJoin(t *testing.T) {
	def := orderDef(t)

	machine, err := workflow.New("order", def, orderStore(), nil)

	if err != nil {
		t.Fatalf("new machine: %v", err)
	}

	order := &Order{ID: "o1"}

	marking, err := machine.Apply(order, "start", nil)

	if err != nil {
		t.Fatalf("apply start: %v", err)
	}

	if !marking.Has("production") || !marking.Has("quality") {
		t.Fatalf("expected AND-split marking, got %v", marking.ActivePlaces())
	}

	if marking.Has("new") {
		t.Fatal("token should have left the new place")
	}

	if !machine.Can(order, "finish") {
		t.Fatal("finish should be enabled with both tokens present")
	}

	partial := &Order{ID: "o2", Places: []string{"production"}}

	if machine.Can(partial, "finish") {
		t.Fatal("finish must not be enabled with only one of two from-places marked")
	}

	marking, err = machine.Apply(order, "finish", nil)

	if err != nil {
		t.Fatalf("apply finish: %v", err)
	}

	if got := marking.ActivePlaces(); len(got) != 1 || got[0] != "shipped" {
		t.Fatalf("expected shipped only, got %v", got)
	}
}

func TestPetriNet_BlockedTransitionReturnsTransitionError(t *testing.T) {
	def := orderDef(t)

	machine, _ := workflow.New("order", def, orderStore(), nil)

	order := &Order{ID: "o3", Places: []string{"new"}}

	_, err := machine.Apply(order, "finish", nil)

	var te *workflow.TransitionError

	if !errors.As(err, &te) {
		t.Fatalf("expected *TransitionError, got %T: %v", err, err)
	}

	if te.Transition != "finish" || te.Machine != "order" {
		t.Fatalf("unexpected error fields: %#v", te)
	}

	if len(te.Blockers) != 0 {
		t.Fatalf("marking-based rejection should carry no blockers, got %#v", te.Blockers)
	}
}
