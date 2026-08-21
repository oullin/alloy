// Package diskwalk sizes directory trees concurrently.
//
// The workload it was built for is a few hundred roots totalling a couple of
// hundred gigabytes across well over a million inodes. That is bound by
// filesystem latency rather than CPU, so the pool oversubscribes cores heavily
// and every design choice here is about keeping workers fed:
//
//   - One pool serves the whole process. A pool per root would spawn thousands
//     of goroutines for a scan that has only a few dozen cores to run them on.
//   - Work is handed out through a condition-variable stack rather than a
//     channel. Every worker is both producer and consumer, so a channel
//     deadlocks the moment one directory yields more subdirectories than the
//     buffer can hold.
//   - Sizes come from allocated blocks, not file sizes, because that is what
//     actually frees when the tree is deleted.
//
// A walk is cancellable at directory granularity: cancelling the context wakes
// every blocked worker, so interrupting a multi-hundred-gigabyte scan returns
// in milliseconds rather than at the end of the tree.
package diskwalk
