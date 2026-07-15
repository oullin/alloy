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
		{"name containing a colon is left alone", "a:b", "queues:a:b"},
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

func TestRedisDelayedAndFailedKeysSuffixTheQueueKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in          string
		wantDelayed string
		wantFailed  string
	}{
		{"emails", "queues:emails:delayed", "queues:emails:failed"},
		{"", "queues:default:delayed", "queues:default:failed"},
	}

	for _, tc := range cases {
		if got := drivers.ExportDelayedKey(tc.in); got != tc.wantDelayed {
			t.Fatalf("delayedKey(%q) = %q, want %q", tc.in, got, tc.wantDelayed)
		}

		if got := drivers.ExportFailedKey(tc.in); got != tc.wantFailed {
			t.Fatalf("failedKey(%q) = %q, want %q", tc.in, got, tc.wantFailed)
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
		{"failed suffix stripped", "queues:default:failed", "default"},
		{"hash tag unwrapped", "queues:{cluster}", "cluster"},
		{"hash tag with suffix", "queues:{cluster}:delayed", "cluster"},
		{"prefix only yields empty", "queues:", ""},
		{"unprefixed key is returned as-is", "emails", "emails"},
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

// The two must compose: a name that survives a round trip through queueKey and
// back is what lets QueueNames report what Push created.
func TestRedisQueueKeyRoundTrip(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"default", "emails", "high-priority"} {
		if got := drivers.ExportQueueNameFromKey(drivers.ExportQueueKey(name)); got != name {
			t.Fatalf("round trip of %q produced %q", name, got)
		}
	}

	// The delayed and failed keys must round-trip to the same logical queue,
	// which is what stops QueueNames reporting them as separate queues.
	for _, key := range []string{drivers.ExportDelayedKey("emails"), drivers.ExportFailedKey("emails")} {
		if got := drivers.ExportQueueNameFromKey(key); got != "emails" {
			t.Fatalf("queueNameFromKey(%q) = %q, want %q", key, got, "emails")
		}
	}
}
