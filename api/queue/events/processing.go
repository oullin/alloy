package events

// JobPopping is dispatched before the worker asks the driver for the next
type JobPopping struct {
	ConnectionName string
}

// JobPopped is dispatched after the worker has retrieved a job from the
// driver but before it begins processing.
type JobPopped struct {
	ConnectionName string
	Job            any
}

// JobProcessing is dispatched immediately before a job is fired.
type JobProcessing struct {
	ConnectionName string
	Job            any
}

// JobAttempted is dispatched after every attempt to run a job, regardless
// of whether it succeeded or threw.
type JobAttempted struct {
	ConnectionName string
	Job            any
}

// JobProcessed is dispatched after a job has been processed successfully.
type JobProcessed struct {
	ConnectionName string
	Job            any
}

// JobFailed is dispatched when a job has exhausted its retry budget and
// is being marked as permanently failed.
type JobFailed struct {
	ConnectionName string
	Job            any
	Err            error
}

// JobExceptionOccurred is dispatched every time a job attempt throws an
// exception, regardless of whether the job will be retried or failed.
type JobExceptionOccurred struct {
	ConnectionName string
	Job            any
	Err            error
}

// JobReleasedAfterException is dispatched when the worker releases a job
// back onto the queue following a caught exception.
type JobReleasedAfterException struct {
	ConnectionName string
	Job            any
}

// JobTimedOut is dispatched when a job exceeds its configured timeout.
type JobTimedOut struct {
	ConnectionName string
	Job            any
}

// JobRetryRequested is dispatched when an operator re-queues a failed job
// via the queue:retry command.
type JobRetryRequested struct {
	// Payload is the failed-job record being retried.
	Payload any
}
