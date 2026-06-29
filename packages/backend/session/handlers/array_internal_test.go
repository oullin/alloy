package handlers

import (
	"context"
	"testing"
	"time"
)

func TestArrayHandlerReadDeletesExpiredSessionUnderWriteLock(t *testing.T) {
	h := NewArrayHandler()
	h.maxLifetime = 1
	h.sessions["expired"] = sessionRecord{
		data:      "stale",
		writtenAt: time.Now().Add(-2 * time.Second),
	}

	data, err := h.Read(context.Background(), "expired")

	if err != nil {
		t.Fatal(err)
	}

	if data != "" {
		t.Fatalf("expected expired session to read empty, got %q", data)
	}

	if _, ok := h.sessions["expired"]; ok {
		t.Fatal("expected expired session to be deleted")
	}
}
