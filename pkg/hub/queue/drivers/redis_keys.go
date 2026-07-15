package drivers

import "strings"

// queuePrefix namespaces every Redis key the queue driver touches.
const queuePrefix = "queues:"

// queueKey returns the Redis list key holding a queue's ready jobs. An empty
// name resolves to the default queue.
func queueKey(q string) string {
	if q == "" {
		q = "default"
	}

	return queuePrefix + q
}

// delayedKey returns the sorted-set key holding a queue's delayed jobs, scored
// by the timestamp at which each becomes due.
func delayedKey(q string) string {
	return queueKey(q) + ":delayed"
}

// failedKey returns the list key holding a queue's failed-job envelopes.
func failedKey(q string) string {
	return queueKey(q) + ":failed"
}

// queueNameFromKey recovers the logical queue name from a scanned Redis key.
// It is the inverse of queueKey, so the two must change together — that is why
// they live in the same file.
//
// It strips the queuePrefix, drops any per-queue ":delayed" or ":failed"
// suffix so each logical queue is reported once, and unwraps a Redis Cluster
// hash tag if one is present.
//
// The hash-tag unwrapping is deliberately asymmetric: this driver never writes
// `queues:{name}` keys, so nothing it creates needs unwrapping. It stays
// because a key can predate this driver or come from a cluster-aware producer
// sharing the same Redis, and a scan must report those under their plain name
// rather than as a separate queue called "{name}". Do not "simplify" it away.
func queueNameFromKey(key string) string {
	name := strings.TrimPrefix(key, queuePrefix)

	if i := strings.IndexByte(name, ':'); i >= 0 {
		name = name[:i]
	}

	return strings.TrimPrefix(strings.TrimSuffix(name, "}"), "{")
}
