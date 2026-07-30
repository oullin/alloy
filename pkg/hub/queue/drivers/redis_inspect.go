package drivers

import (
	"context"
	"encoding/json"
	"time"

	"hara.sh/alloy/queue"
)

// QueueNames enumerates the queues currently known to the underlying
// Redis instance by scanning for keys matching `queues:*`. Names
// derived from cluster-style `queues:{name}` keys are unwrapped so the
// caller receives plain logical queue names. The driver returns
// ErrNotSupported when its client cannot SCAN — that capability is
// optional via the RedisScanner interface.
func (d *RedisDriver) QueueNames(ctx context.Context) ([]string, error) {
	scanner, ok := d.client.(RedisScanner)

	if !ok {
		return nil, queue.ErrNotSupported
	}

	keys, err := scanner.ScanMatch(ctx, queuePrefix+"*")

	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})

	var out []string

	for _, k := range keys {
		name := queueNameFromKey(k)

		if name == "" {
			continue
		}

		if _, already := seen[name]; already {
			continue
		}

		seen[name] = struct{}{}

		out = append(out, name)
	}

	return out, nil
}

// PendingJobs returns the raw payloads currently waiting on the queue
// list. The driver requires its client to satisfy the optional
// RedisListRanger interface; otherwise it returns ErrNotSupported.
func (d *RedisDriver) PendingJobs(ctx context.Context, queueName string) ([]queue.InspectedJob, error) {
	ranger, ok := d.client.(RedisListRanger)

	if !ok {
		return nil, queue.ErrNotSupported
	}

	raws, err := ranger.LRange(ctx, queueKey(queueName), 0, -1)

	if err != nil {
		return nil, err
	}

	return d.snapshotsFromPayloads(queueName, raws), nil
}

// DelayedJobs returns the raw payloads parked in the delayed sorted
// set for queueName. The driver requires its client to satisfy the
// optional RedisSortedSetRanger interface.
func (d *RedisDriver) DelayedJobs(ctx context.Context, queueName string) ([]queue.InspectedJob, error) {
	ranger, ok := d.client.(RedisSortedSetRanger)

	if !ok {
		return nil, queue.ErrNotSupported
	}

	raws, err := ranger.ZRange(ctx, delayedKey(queueName), 0, -1)

	if err != nil {
		return nil, err
	}

	return d.snapshotsFromPayloads(queueName, raws), nil
}

// ReservedJobs returns the raw payloads currently reserved (in-flight) in the
// reserved sorted set for queueName. The driver requires its client to satisfy
// the optional RedisSortedSetRanger interface; otherwise it returns ErrNotSupported.
func (d *RedisDriver) ReservedJobs(ctx context.Context, queueName string) ([]queue.InspectedJob, error) {
	ranger, ok := d.client.(RedisSortedSetRanger)

	if !ok {
		return nil, queue.ErrNotSupported
	}

	raws, err := ranger.ZRange(ctx, reservedKey(queueName), 0, -1)

	if err != nil {
		return nil, err
	}

	return d.snapshotsFromPayloads(queueName, raws), nil
}

// snapshotsFromPayloads converts raw Redis payload strings into InspectedJob
// entries, decoding the JSON envelope for displayName, uuid, and createdAt the
// same way the database driver does.
//
// InspectedJob.ReservedAt is left nil for every caller, including ReservedJobs.
// The reserved sorted set scores each member with its reclaim deadline, so the
// data exists; surfacing it needs a scored range, which the client contract
// does not expose yet.
func (d *RedisDriver) snapshotsFromPayloads(queueName string, raws []string) []queue.InspectedJob {
	if len(raws) == 0 {
		return nil
	}

	out := make([]queue.InspectedJob, 0, len(raws))

	for _, raw := range raws {
		job := queue.InspectedJob{
			Backend:    queueName,
			Connection: d.connection,
			Payload:    []byte(raw),
		}

		var decoded map[string]any

		if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
			if v, ok := decoded["displayName"].(string); ok {
				job.Name = v
			}

			if v, ok := decoded["uuid"].(string); ok {
				job.UUID = v
			}

			if v, ok := decoded["createdAt"].(float64); ok {
				job.CreatedAt = time.Unix(int64(v), 0)
			}
		}

		out = append(out, job)
	}

	return out
}
