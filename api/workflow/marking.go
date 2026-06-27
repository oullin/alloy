package workflow

import cworkflow "alloy.dev/api/contracts/workflow"

// Marking tracks active places and token counts.
type Marking = cworkflow.Marking

// NewMarking constructs a marking from active places.
var NewMarking = cworkflow.NewMarking
