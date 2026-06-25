package queue

import cqueue "github.com/oullin/alloy/api/contracts/queue"

// Job represents a queued job instance.
type Job = cqueue.Job

// Handler handles a job type.
type Handler = cqueue.Handler

// HandlerFunc is a function that implements Handler.
type HandlerFunc = cqueue.HandlerFunc

// FailureHandler is the optional contract a Handler can implement to
// receive a callback when a job it was processing has been marked as
// failed. The driver or worker invokes Failed after Job.Fail has been
// called and before the JobFailed event is emitted.
//
// callback invoked by CallQueuedHandler::failed.
type FailureHandler = cqueue.FailureHandler

// Handle implements Handler.

// JobOptions configures job dispatch options.
type JobOptions = cqueue.JobOptions
