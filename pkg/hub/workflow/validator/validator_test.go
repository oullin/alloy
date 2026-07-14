package validator_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/workflow"
	"github.com/oullin/alloy/pkg/hub/workflow/validator"
)

func stateMachineDef(t *testing.T) *workflow.Definition {
	t.Helper()

	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("b").
		SetInitialPlaces("a").
		AddTransition("go", []string{"a"}, []string{"b"}).
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	return def
}

func TestValidateDefinitionRequiresDefinition(t *testing.T) {
	if err := validator.ValidateDefinition(nil); err == nil {
		t.Fatal("expected error for nil definition")
	}
}

func TestValidateDefinitionAcceptsValidDefinition(t *testing.T) {
	if err := validator.ValidateDefinition(stateMachineDef(t)); err != nil {
		t.Fatalf("ValidateDefinition: %v", err)
	}
}

func TestValidateDefinitionRejectsInvalidDefinition(t *testing.T) {
	def := &workflow.Definition{
		Places:         []string{"a", "orphan"},
		InitialMarking: workflow.Marking{Places: map[string]int{"a": 1}},
	}

	if err := validator.ValidateDefinition(def); err == nil {
		t.Fatal("expected error for unreachable place")
	}
}

func TestValidateStateMachineAcceptsValidDefinition(t *testing.T) {
	if err := validator.ValidateStateMachine(stateMachineDef(t)); err != nil {
		t.Fatalf("ValidateStateMachine: %v", err)
	}
}

func TestValidateStateMachineRejectsInvalidDefinition(t *testing.T) {
	if err := validator.ValidateStateMachine(nil); err == nil {
		t.Fatal("expected error for nil definition")
	}
}

func TestValidateStateMachineRequiresSingleInitialPlace(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("b").
		SetInitialPlaces("a", "b").
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	if err := validator.ValidateStateMachine(def); err == nil {
		t.Fatal("expected error for multiple initial places")
	}
}

func TestValidateStateMachineRequiresSingleTargetTransitions(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("b").
		AddPlace("c").
		SetInitialPlaces("a").
		AddTransition("split", []string{"a"}, []string{"b", "c"}).
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	if err := validator.ValidateStateMachine(def); err == nil {
		t.Fatal("expected error for multi-target transition")
	}
}
