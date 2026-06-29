package failed

import cfailed "alloy.dev/backend/contracts/queue/failed"

// Job is the decoded record returned by All and Find. It is the
// Go analogue of the stdClass rows the upstream providers yield: the public
// fields mirror failed_jobs columns, plus an ID that maps to either the
// integer primary key (database) or the job UUID (uuid/file/dynamodb).
type Job = cfailed.Job

// persistence layer the worker uses to record permanently-failed jobs.
//
// Differences from PHP:
//
//   - IDs returns []string; callers coerce where they need ints.
//   - Find returns (nil, nil) when the id does not exist (upstream returns null).
//   - Flush takes `hours int` where 0 means "flush everything". upstream
//     accepts int|null with null meaning the same.
type Provider = cfailed.Provider

// Providers that
// support efficient counting implement this optional interface.
// A nil or empty string acts as "any" for the corresponding dimension.
type Countable = cfailed.Countable

// Implementations
// return the number of rows removed.
type Prunable = cfailed.Prunable
