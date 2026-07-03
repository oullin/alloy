package workflow_test

import (
	"testing"

	"alloy.dev/foundation/workflow"
)

func TestTransitionBlockerList(t *testing.T) {
	list := &workflow.TransitionBlockerList{}

	if !list.Empty() {
		t.Fatal("new list should be empty")
	}

	if len(list.All()) != 0 {
		t.Fatal("new list should have no blockers")
	}

	list.Add(workflow.TransitionBlocker{Message: "first", Code: "f1"})
	list.Add(workflow.TransitionBlocker{Message: "second"})

	if list.Empty() {
		t.Fatal("list with blockers should not be empty")
	}

	all := list.All()

	if len(all) != 2 || all[0].Message != "first" || all[0].Code != "f1" || all[1].Message != "second" {
		t.Fatalf("All() = %#v", all)
	}

	all[0].Message = "mutated"

	if list.All()[0].Message != "first" {
		t.Fatal("All() should return a copy")
	}
}
