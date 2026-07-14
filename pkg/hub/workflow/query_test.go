package workflow_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/workflow"
	"github.com/oullin/alloy/pkg/hub/workflow/store"
)

func TestQuery_CanAndCanNot(t *testing.T) {
	def := subscriptionDef(t)
	sm, _ := workflow.NewStateMachine("subscription", def, subscriptionStore(), nil)

	sub := &Subscription{State: "trial"}

	if !sm.Can(sub, "activate") {
		t.Fatal("activate should be enabled from trial")
	}

	if sm.Can(sub, "cancel") {
		t.Fatal("cancel should not be enabled from trial")
	}

	if !sm.CanNot(sub, "cancel") {
		t.Fatal("CanNot should be the inverse of Can")
	}

	if sm.Can(sub, "bogus") {
		t.Fatal("unknown transition should never be enabled")
	}
}

func TestQuery_EnabledAndDisabledTransitions(t *testing.T) {
	def := subscriptionDef(t)
	sm, _ := workflow.NewStateMachine("subscription", def, subscriptionStore(), nil)

	sub := &Subscription{State: "trial"}

	enabled, err := sm.EnabledTransitions(sub)

	if err != nil {
		t.Fatalf("enabled transitions: %v", err)
	}

	if len(enabled) != 1 || enabled[0].Name != "activate" {
		t.Fatalf("expected [activate], got %v", enabled)
	}

	disabled, err := sm.DisabledTransitions(sub)

	if err != nil {
		t.Fatalf("disabled transitions: %v", err)
	}

	if len(disabled) != 1 || disabled[0].Name != "cancel" {
		t.Fatalf("expected [cancel], got %v", disabled)
	}

	if _, err := sm.Apply(sub, "activate", nil); err != nil {
		t.Fatalf("apply activate: %v", err)
	}

	enabled, err = sm.EnabledTransitions(sub)

	if err != nil {
		t.Fatalf("enabled transitions after apply: %v", err)
	}

	if len(enabled) != 1 || enabled[0].Name != "cancel" {
		t.Fatalf("expected [cancel] after activation, got %v", enabled)
	}
}

func TestQuery_TransitionListsPropagateStoreError(t *testing.T) {
	def := subscriptionDef(t)
	broken := &store.SingleState[*Subscription]{}

	sm, _ := workflow.NewStateMachine("subscription", def, broken, nil)

	if _, err := sm.EnabledTransitions(&Subscription{}); err == nil {
		t.Fatal("expected error from broken store")
	}

	if _, err := sm.DisabledTransitions(&Subscription{}); err == nil {
		t.Fatal("expected error from broken store")
	}
}
