package workflow_test

import (
	"testing"

	"hara.sh/alloy/workflow"
)

func TestDefinitionBuilderZeroValueBuildFails(t *testing.T) {
	b := &workflow.DefinitionBuilder{}

	// Every builder method must be nil-safe on the zero value.
	b.AddPlace("a").
		SetInitialPlaces("a").
		AddTransition("go", []string{"a"}, []string{"a"}).
		SetMetadata("k", "v").
		SetPlaceMetadata("a", "k", "v").
		SetTransitionMetadata("go", "k", "v")

	if _, err := b.Build(); err == nil {
		t.Fatal("expected error building from a zero-value builder")
	}
}

func TestDefinitionBuilderIgnoresEmptyAndDuplicatePlaces(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("").
		AddPlace("a").
		SetInitialPlaces("a").
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	if len(def.Places) != 1 || def.Places[0] != "a" {
		t.Fatalf("expected places [a], got %v", def.Places)
	}
}

func TestDefinitionBuilderBuildRejectsInvalidDefinition(t *testing.T) {
	_, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		Build()

	if err == nil {
		t.Fatal("expected error for definition without an initial marking")
	}
}

func TestDefinitionBuilderMergesTransitionMetadataIntoTransitions(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("a").
		AddPlace("b").
		SetInitialPlaces("a").
		AddTransition("go", []string{"a"}, []string{"b"}).
		SetTransitionMetadata("go", "verb", "activate").
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	if len(def.Transitions) != 1 {
		t.Fatalf("expected one transition, got %d", len(def.Transitions))
	}

	if def.Transitions[0].Metadata["verb"] != "activate" {
		t.Fatalf("expected merged transition metadata, got %#v", def.Transitions[0].Metadata)
	}
}
