package proc_test

import (
	"os"
	"testing"

	"hara.sh/alloy/treex/internal/proc"
)

func TestAliveFindsTheRunningProcess(t *testing.T) {
	t.Parallel()

	if !proc.Alive(os.Getpid()) {
		t.Fatal("Alive(self) = false, want true")
	}
}

func TestAliveRejectsAnImpossiblePID(t *testing.T) {
	t.Parallel()

	for _, pid := range []int{0, -1} {
		if proc.Alive(pid) {
			t.Fatalf("Alive(%d) = true, want false", pid)
		}
	}
}
