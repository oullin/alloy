package workflow

import cworkflow "alloy.dev/go/contracts/workflow"

// Transition is an arc in the Petri-Net moving tokens from From places to To places.
type Transition = cworkflow.Transition

// NewTransition constructs a Transition.
var NewTransition = cworkflow.NewTransition
