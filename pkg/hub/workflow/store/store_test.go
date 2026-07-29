package store_test

import (
	"slices"
	"testing"

	"hara.sh/alloy/workflow"
	"hara.sh/alloy/workflow/store"
)

type entity struct {
	state  string
	places []string
}

func singleStore() *store.SingleState[*entity] {
	return &store.SingleState[*entity]{
		Getter: func(e *entity) string { return e.state },
		Setter: func(e *entity, place string) { e.state = place },
	}
}

func multiStore() *store.MultiState[*entity] {
	return &store.MultiState[*entity]{
		Getter: func(e *entity) []string { return e.places },
		Setter: func(e *entity, places []string) { e.places = places },
	}
}

func TestSingleState_RoundTrip(t *testing.T) {
	s := singleStore()
	e := &entity{}

	if err := s.SetMarking(e, workflow.NewMarking("active"), nil, nil); err != nil {
		t.Fatalf("set marking: %v", err)
	}

	if e.state != "active" {
		t.Fatalf("expected subject state 'active', got %q", e.state)
	}

	marking, err := s.GetMarking(e, nil)

	if err != nil {
		t.Fatalf("get marking: %v", err)
	}

	if !marking.Has("active") || len(marking.ActivePlaces()) != 1 {
		t.Fatalf("expected [active], got %v", marking.ActivePlaces())
	}
}

func TestSingleState_EmptyStateYieldsEmptyMarking(t *testing.T) {
	marking, err := singleStore().GetMarking(&entity{}, nil)

	if err != nil {
		t.Fatalf("get marking: %v", err)
	}

	if len(marking.ActivePlaces()) != 0 {
		t.Fatalf("expected empty marking, got %v", marking.ActivePlaces())
	}
}

func TestSingleState_SetEmptyMarkingClearsState(t *testing.T) {
	s := singleStore()
	e := &entity{state: "active"}

	if err := s.SetMarking(e, workflow.Marking{}, nil, nil); err != nil {
		t.Fatalf("set marking: %v", err)
	}

	if e.state != "" {
		t.Fatalf("expected cleared state, got %q", e.state)
	}
}

func TestSingleState_RejectsMultiplePlaces(t *testing.T) {
	err := singleStore().SetMarking(&entity{}, workflow.NewMarking("a", "b"), nil, nil)

	if err == nil {
		t.Fatal("expected error persisting two active places")
	}
}

func TestSingleState_RequiresGetterAndSetter(t *testing.T) {
	var nilStore *store.SingleState[*entity]

	if _, err := nilStore.GetMarking(&entity{}, nil); err == nil {
		t.Fatal("expected error from nil receiver")
	}

	if err := nilStore.SetMarking(&entity{}, workflow.Marking{}, nil, nil); err == nil {
		t.Fatal("expected error from nil receiver")
	}

	empty := &store.SingleState[*entity]{}

	if _, err := empty.GetMarking(&entity{}, nil); err == nil {
		t.Fatal("expected error for missing getter")
	}

	if err := empty.SetMarking(&entity{}, workflow.Marking{}, nil, nil); err == nil {
		t.Fatal("expected error for missing setter")
	}
}

func TestMultiState_RoundTrip(t *testing.T) {
	s := multiStore()
	e := &entity{}

	if err := s.SetMarking(e, workflow.NewMarking("production", "quality"), nil, nil); err != nil {
		t.Fatalf("set marking: %v", err)
	}

	if !slices.Equal(e.places, []string{"production", "quality"}) {
		t.Fatalf("expected persisted places, got %v", e.places)
	}

	marking, err := s.GetMarking(e, nil)

	if err != nil {
		t.Fatalf("get marking: %v", err)
	}

	if !marking.Has("production") || !marking.Has("quality") {
		t.Fatalf("expected both places, got %v", marking.ActivePlaces())
	}
}

func TestMultiState_RequiresGetterAndSetter(t *testing.T) {
	var nilStore *store.MultiState[*entity]

	if _, err := nilStore.GetMarking(&entity{}, nil); err == nil {
		t.Fatal("expected error from nil receiver")
	}

	if err := nilStore.SetMarking(&entity{}, workflow.Marking{}, nil, nil); err == nil {
		t.Fatal("expected error from nil receiver")
	}

	empty := &store.MultiState[*entity]{}

	if _, err := empty.GetMarking(&entity{}, nil); err == nil {
		t.Fatal("expected error for missing getter")
	}

	if err := empty.SetMarking(&entity{}, workflow.Marking{}, nil, nil); err == nil {
		t.Fatal("expected error for missing setter")
	}
}
