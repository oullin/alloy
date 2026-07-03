package workflow_test

import (
	"testing"

	"github.com/oullin/alloy/packages/foundation/workflow"
)

func TestDefinitionMetadataValue(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		SetInitialPlaces("a").
		SetMetadata("title", "Subscription").
		Build()

	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if v, ok := def.MetadataValue("title"); !ok || v != "Subscription" {
		t.Errorf("MetadataValue = %v,%v, want Subscription,true", v, ok)
	}

	if _, ok := def.MetadataValue("missing"); ok {
		t.Errorf("MetadataValue(missing) reported true")
	}
}

func TestDefinitionPlaceMetadataValue(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("trial").
		SetInitialPlaces("trial").
		SetPlaceMetadata("trial", "label", "Free trial").
		Build()

	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if v, ok := def.PlaceMetadataValue("trial", "label"); !ok || v != "Free trial" {
		t.Errorf("PlaceMetadataValue = %v,%v", v, ok)
	}

	if _, ok := def.PlaceMetadataValue("trial", "missing"); ok {
		t.Errorf("missing key reported true")
	}

	if _, ok := def.PlaceMetadataValue("nonexistent", "label"); ok {
		t.Errorf("missing place reported true")
	}
}

func TestDefinitionTransitionMetadataValue(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("b").
		SetInitialPlaces("a").
		AddTransition("go", []string{"a"}, []string{"b"}).
		SetTransitionMetadata("go", "verb", "activate").
		Build()

	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if v, ok := def.TransitionMetadataValue("go", "verb"); !ok || v != "activate" {
		t.Errorf("TransitionMetadataValue = %v,%v", v, ok)
	}

	if _, ok := def.TransitionMetadataValue("missing-transition", "verb"); ok {
		t.Errorf("missing transition reported true")
	}

	if _, ok := def.TransitionMetadataValue("go", "missing-key"); ok {
		t.Errorf("missing key reported true")
	}
}

func TestDefinitionValidateRejectsEmptyTransitionName(t *testing.T) {
	def := &workflow.Definition{
		Places:         []string{"a", "b"},
		InitialMarking: workflow.Marking{Places: map[string]int{"a": 1}},
		Transitions:    []workflow.Transition{{Name: "", From: []string{"a"}, To: []string{"b"}}},
	}

	if err := def.Validate(); err == nil {
		t.Fatal("expected error for empty transition name")
	}
}

func TestDefinitionValidateRejectsMissingFromOrTo(t *testing.T) {
	cases := []workflow.Transition{
		{Name: "t1", From: nil, To: []string{"b"}},
		{Name: "t2", From: []string{"a"}, To: nil},
	}

	for _, tr := range cases {
		def := &workflow.Definition{
			Places:         []string{"a", "b"},
			InitialMarking: workflow.Marking{Places: map[string]int{"a": 1}},
			Transitions:    []workflow.Transition{tr},
		}

		if err := def.Validate(); err == nil {
			t.Errorf("expected error for transition with empty endpoints: %+v", tr)
		}
	}
}

func TestDefinitionValidateRejectsUnknownPlace(t *testing.T) {
	def := &workflow.Definition{
		Places:         []string{"a"},
		InitialMarking: workflow.Marking{Places: map[string]int{"a": 1}},
		Transitions:    []workflow.Transition{{Name: "t", From: []string{"a"}, To: []string{"ghost"}}},
	}

	if err := def.Validate(); err == nil {
		t.Fatal("expected error for unknown place in transition")
	}
}

func TestDefinitionValidateRejectsMissingInitialMarking(t *testing.T) {
	def := &workflow.Definition{
		Places: []string{"a"},
	}

	if err := def.Validate(); err == nil {
		t.Fatal("expected error for missing initial marking")
	}
}

func TestDefinitionValidateRejectsUnreachablePlace(t *testing.T) {
	def := &workflow.Definition{
		Places:         []string{"a", "orphan"},
		InitialMarking: workflow.Marking{Places: map[string]int{"a": 1}},
	}

	if err := def.Validate(); err == nil {
		t.Fatal("expected error for unreachable place")
	}
}

func TestDefinitionCloneNilReturnsNil(t *testing.T) {
	var def *workflow.Definition

	if def.Clone() != nil {
		t.Fatal("cloning a nil definition should return nil")
	}
}

func TestDefinitionTransitionAttachesMetadataFromMap(t *testing.T) {
	def := &workflow.Definition{
		Places:             []string{"a", "b"},
		InitialMarking:     workflow.Marking{Places: map[string]int{"a": 1}},
		Transitions:        []workflow.Transition{{Name: "go", From: []string{"a"}, To: []string{"b"}}},
		TransitionMetadata: map[string]map[string]any{"go": {"verb": "activate"}},
	}

	transition, ok := def.Transition("go")

	if !ok {
		t.Fatal("expected transition to be found")
	}

	if transition.Metadata["verb"] != "activate" {
		t.Fatalf("expected metadata attached from the definition map, got %#v", transition.Metadata)
	}

	if _, ok := def.Transition("missing"); ok {
		t.Fatal("unknown transition reported true")
	}
}

func TestDefinitionValidateRejectsNoPlaces(t *testing.T) {
	def := &workflow.Definition{}

	if err := def.Validate(); err == nil {
		t.Fatal("expected error for definition without places")
	}
}

func TestDefinitionValidateRejectsUnknownInitialPlace(t *testing.T) {
	def := &workflow.Definition{
		Places:         []string{"a"},
		InitialMarking: workflow.Marking{Places: map[string]int{"ghost": 1}},
	}

	if err := def.Validate(); err == nil {
		t.Fatal("expected error for unknown place in initial marking")
	}
}

func TestDefinitionValidateAllowsHappyPath(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("b").
		SetInitialPlaces("a").
		AddTransition("go", []string{"a"}, []string{"b"}).
		Build()

	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if err := def.Validate(); err != nil {
		t.Errorf("happy-path Validate: %v", err)
	}
}
