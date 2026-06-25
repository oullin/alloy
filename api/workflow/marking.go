package workflow

import cworkflow "github.com/oullin/alloy/api/contracts/workflow"

// Marking tracks active places and token counts.
type Marking = cworkflow.Marking

// NewMarking constructs a marking from active places.
var NewMarking = cworkflow.NewMarking
