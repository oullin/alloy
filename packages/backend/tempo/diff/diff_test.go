package diff_test

import (
	"testing"

	"alloy.dev/backend/tempo"
)

func TestDiffInDaysUsesUnixSeconds(t *testing.T) {
	later, err := tempo.Parse("2024-01-03T00:00:00+00:00")

	if err != nil {
		t.Fatalf("parse later: %v", err)
	}

	earlier, err := tempo.Parse("2024-01-01T00:00:00+00:00")

	if err != nil {
		t.Fatalf("parse earlier: %v", err)
	}

	if got := later.DiffInDays(earlier); got != 2 {
		t.Fatalf("DiffInDays() = %d, want 2", got)
	}
}
