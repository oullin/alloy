package workflow_test

import (
	"strings"
	"testing"

	"github.com/oullin/alloy/pkg/hub/workflow"
)

func TestTransitionErrorMessageWithoutBlockers(t *testing.T) {
	err := &workflow.TransitionError{Machine: "subscription", Transition: "activate"}

	want := `cannot apply transition "activate" on workflow "subscription"`

	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestTransitionErrorMessageJoinsBlockers(t *testing.T) {
	err := &workflow.TransitionError{
		Machine:    "subscription",
		Transition: "activate",
		Blockers: []workflow.TransitionBlocker{
			{Message: "billing not configured"},
			{Message: "email unverified"},
		},
	}

	msg := err.Error()

	if !strings.Contains(msg, "billing not configured; email unverified") {
		t.Fatalf("Error() = %q, want joined blocker messages", msg)
	}
}
