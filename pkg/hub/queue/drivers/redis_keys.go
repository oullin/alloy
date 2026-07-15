package drivers

import "strings"

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
func delayedKey(q string) string { return queueKey(q) + ":delayed" }

// reservedKey returns the sorted-set key holding a queue's in-flight jobs,
// scored by the deadline at which each becomes reclaimable.
func reservedKey(q string) string { return queueKey(q) + ":reserved" }

// failedKey returns the list key holding a queue's failed-job envelopes.
func failedKey(q string) string { return queueKey(q) + ":failed" }

// queueNameFromKey recovers the logical queue name from a scanned Redis key.
// It is the inverse of queueKey, so the two must change together — that is why
// they live in the same file.
//
// Only the exact per-queue suffixes are stripped, so a queue name that itself
// contains a colon (tenant:1) survives the round trip.
//
// Hash tags are unwrapped even though this driver never writes them: a scan can
// turn up keys written by a cluster-aware producer sharing the same Redis, and
// those should report under their plain name rather than as a queue literally
// called "{name}". Do not "simplify" this away.
func queueNameFromKey(key string) string {
	name := strings.TrimPrefix(key, queuePrefix)

	for _, suffix := range []string{":delayed", ":reserved", ":failed"} {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)

			break
		}
	}

	return strings.TrimPrefix(strings.TrimSuffix(name, "}"), "{")
}
