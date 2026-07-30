package workflow_test

import (
	"testing"

	"hara.sh/alloy/workflow"
)

func TestMetadataStoreExposesDefinitionMetadata(t *testing.T) {
	def, err := workflow.NewDefinitionBuilder().
		AddPlace("trial").
		AddPlace("active").
		SetInitialPlaces("trial").
		AddTransition("activate", []string{"trial"}, []string{"active"}).
		SetMetadata("title", "Subscription").
		SetPlaceMetadata("trial", "label", "Free trial").
		SetTransitionMetadata("activate", "verb", "Activate").
		Build()

	if err != nil {
		t.Fatalf("build definition: %v", err)
	}

	sm, err := workflow.NewStateMachine("subscription", def, subscriptionStore(), nil)

	if err != nil {
		t.Fatalf("new state machine: %v", err)
	}

	ms := sm.MetadataStore()

	if v, ok := ms.WorkflowMetadata("title"); !ok || v != "Subscription" {
		t.Fatalf("WorkflowMetadata = %v, %v", v, ok)
	}

	if _, ok := ms.WorkflowMetadata("missing"); ok {
		t.Fatal("missing workflow metadata key reported true")
	}

	if v, ok := ms.PlaceMetadata("trial", "label"); !ok || v != "Free trial" {
		t.Fatalf("PlaceMetadata = %v, %v", v, ok)
	}

	if _, ok := ms.PlaceMetadata("ghost", "label"); ok {
		t.Fatal("missing place reported true")
	}

	if v, ok := ms.TransitionMetadata("activate", "verb"); !ok || v != "Activate" {
		t.Fatalf("TransitionMetadata = %v, %v", v, ok)
	}

	if _, ok := ms.TransitionMetadata("activate", "missing"); ok {
		t.Fatal("missing transition metadata key reported true")
	}
}
