package drivers_test

import (
	"testing"

	"github.com/oullin/alloy/pkg/hub/queue/drivers"
)

func TestRedisQueueKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain name", "emails", "queues:emails"},
		{"empty name falls back to default", "", "queues:default"},
		{"name containing a colon is preserved", "tenant:1", "queues:tenant:1"},
		{"hash-tagged name is not treated specially", "{tenant}", "queues:{tenant}"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := drivers.ExportQueueKey(tc.in); got != tc.want {
				t.Fatalf("queueKey(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestRedisSuffixKeysBuildOnTheQueueKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in                                  string
		wantDelayed, wantReserved, wantFail string
	}{
		{"emails", "queues:emails:delayed", "queues:emails:reserved", "queues:emails:failed"},
		{"", "queues:default:delayed", "queues:default:reserved", "queues:default:failed"},
	}

	for _, tc := range cases {
		if got := drivers.ExportDelayedKey(tc.in); got != tc.wantDelayed {
			t.Fatalf("delayedKey(%q) = %q, want %q", tc.in, got, tc.wantDelayed)
		}

		if got := drivers.ExportReservedKey(tc.in); got != tc.wantReserved {
			t.Fatalf("reservedKey(%q) = %q, want %q", tc.in, got, tc.wantReserved)
		}

		if got := drivers.ExportFailedKey(tc.in); got != tc.wantFail {
			t.Fatalf("failedKey(%q) = %q, want %q", tc.in, got, tc.wantFail)
		}
	}
}

// queueNameFromKey is the inverse of queueKey. The hash-tag cases matter even
// though this driver never writes tagged keys: a scan can turn up keys written
// by a cluster-aware producer sharing the same Redis.
func TestRedisQueueNameFromKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain queue key", "queues:default", "default"},
		{"delayed suffix stripped", "queues:default:delayed", "default"},
		{"reserved suffix stripped", "queues:default:reserved", "default"},
		{"failed suffix stripped", "queues:default:failed", "default"},
		{"hash tag unwrapped", "queues:{cluster}", "cluster"},
		{"hash tag with suffix", "queues:{cluster}:delayed", "cluster"},
		{"prefix only yields empty", "queues:", ""},
		{"unprefixed key returned as-is", "emails", "emails"},
		// Only the exact per-queue suffixes are stripped, so a colon inside the
		// queue name itself survives. Truncating at the first colon would report
		// this queue as "tenant" and silently merge every tenant's queue.
		{"colon in name is preserved", "queues:tenant:1", "tenant:1"},
		{"colon in name with suffix", "queues:tenant:1:delayed", "tenant:1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := drivers.ExportQueueNameFromKey(tc.in); got != tc.want {
				t.Fatalf("queueNameFromKey(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// The two must compose: what Push creates, QueueNames must report back.
func TestRedisQueueKeyRoundTrip(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"default", "emails", "high-priority", "tenant:1"} {
		if got := drivers.ExportQueueNameFromKey(drivers.ExportQueueKey(name)); got != name {
			t.Fatalf("round trip of %q produced %q", name, got)
		}

		for _, key := range []string{
			drivers.ExportDelayedKey(name),
			drivers.ExportReservedKey(name),
			drivers.ExportFailedKey(name),
		} {
			if got := drivers.ExportQueueNameFromKey(key); got != name {
				t.Fatalf("queueNameFromKey(%q) = %q, want %q", key, got, name)
			}
		}
	}
}
