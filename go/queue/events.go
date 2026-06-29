package queue

import "alloy.dev/go/queue/events"

// This file re-exports the event types defined in the events subpackage at
// the root of the queue package. The aliases exist for two reasons:
//
//  1. Callers (bus, worker, tests) can keep writing queue.JobProcessing
//     rather than events.JobProcessing — a shorter, more familiar name.
//  2. Worker emission sites stay compact.
//
// The subpackage is the source of truth; changing a field there changes
// it here. New upstream events should be added to events/ first, then
// re-exported here only if callers need the unqualified name.
//
// re-exported below so a migration to qualified names is a pure find &
// replace.

// --- Job queueing (push side) -----------------------------------------

type JobQueueing = events.JobQueueing

type JobQueued = events.JobQueued

// --- Job processing (worker side) -------------------------------------

type JobPopping = events.JobPopping

type JobPopped = events.JobPopped

type JobProcessing = events.JobProcessing

type JobAttempted = events.JobAttempted

type JobProcessed = events.JobProcessed

type JobFailed = events.JobFailed

type JobExceptionOccurred = events.JobExceptionOccurred

type JobReleasedAfterException = events.JobReleasedAfterException

type JobTimedOut = events.JobTimedOut

type JobRetryRequested = events.JobRetryRequested

// --- Backend state ------------------------------------------------------

type Looping = events.Looping

type Busy = events.Busy

type Paused = events.Paused

type Resumed = events.Resumed

type FailedOver = events.FailedOver

// --- Worker lifecycle -------------------------------------------------

type WorkerStarting = events.WorkerStarting

type WorkerStopping = events.WorkerStopping

type WorkerPausing = events.WorkerPausing

type WorkerResuming = events.WorkerResuming
