package multisteps

import "context"

// Task is a unit of concurrent work.
type Task func() (any, error)

// Driver defines the concurrency backend contract.
type Driver interface {
	Run(ctx context.Context, tasks []Task) ([]any, error)
}

// Result captures the outcome of a workflow run.
type Result struct {
	Responses map[string]any
	Skipped   []string
}

// Job is the minimal description of a step in a multi-step workflow.
type Job interface {
	Name() string
	IsAsync() bool
}
