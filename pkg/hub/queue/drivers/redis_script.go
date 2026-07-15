package drivers

import (
	"context"

	"time"
)

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

const popAndReserveLua = `
local job = redis.call('rpop', KEYS[1])
if job then
	redis.call('zadd', KEYS[2], ARGV[1], job)
end
return job
`

func (d *RedisDriver) migrateDue(ctx context.Context, queueName string) {
	_, _ = d.client.Eval(ctx, migrateDueLua, []string{delayedKey(queueName), queueKey(queueName)}, time.Now().Unix())
}

func (d *RedisDriver) migrateExpired(ctx context.Context, queueName string) {
	_, _ = d.client.Eval(ctx, migrateDueLua, []string{reservedKey(queueName), queueKey(queueName)}, time.Now().Unix())
}
