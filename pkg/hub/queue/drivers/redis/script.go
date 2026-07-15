package redis

import (
	"context"

	"time"
)

// migrateDueLua atomically moves every delayed job whose score has come due
// onto the ready list, in one round trip.
//
// KEYS[1] is the delayed sorted set, KEYS[2] the ready list, ARGV[1] the
// cutoff timestamp.
const migrateDueLua = `
local jobs = redis.call('zrangebyscore', KEYS[1], '-inf', ARGV[1])

for i, job in ipairs(jobs) do
	redis.call('lpush', KEYS[2], job)
end

if #jobs > 0 then
	redis.call('zrem', KEYS[1], unpack(jobs))
end

return #jobs
`

// migrateDue promotes any due delayed jobs onto the ready list before a Pop
// reads it.
//
// Errors are ignored deliberately: a failed migration leaves the jobs parked
// in the delayed set for the next Pop to retry, which is preferable to failing
// a Pop that could still serve an already-ready job.
func (d *Driver) migrateDue(ctx context.Context, queueName string) {
	_, _ = d.client.Eval(ctx, migrateDueLua, []string{delayedKey(queueName), queueKey(queueName)}, time.Now().Unix())
}
