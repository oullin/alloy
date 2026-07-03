package foundation_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oullin/alloy/packages/foundation/httpx/foundation"
)

func TestStreamedEventString(t *testing.T) {
	t.Parallel()

	event := &foundation.StreamedEvent{
		Event: "message",
		Data:  "hello world",
		ID:    "1",
	}

	s := event.String()

	if !strings.Contains(s, "id: 1") {
		t.Fatal("expected id field")
	}

	if !strings.Contains(s, "event: message") {
		t.Fatal("expected event field")
	}

	if !strings.Contains(s, "data: hello world") {
		t.Fatal("expected data field")
	}
}

func TestStreamedEventMultilineData(t *testing.T) {
	t.Parallel()

	event := &foundation.StreamedEvent{
		Data: "line1\nline2\nline3",
	}

	s := event.String()

	if strings.Count(s, "data: ") != 3 {
		t.Fatalf("expected 3 data lines, got:\n%s", s)
	}
}

func TestStreamedEventWithRetry(t *testing.T) {
	t.Parallel()

	event := &foundation.StreamedEvent{
		Data:  "hello",
		Retry: 5000,
	}

	s := event.String()

	if !strings.Contains(s, "retry: 5000") {
		t.Fatal("expected retry field")
	}
}

func TestStreamedEventMinimal(t *testing.T) {
	t.Parallel()

	event := &foundation.StreamedEvent{
		Data: "just data",
	}

	s := event.String()

	if strings.Contains(s, "id:") {
		t.Fatal("expected no id field")
	}

	if strings.Contains(s, "event:") {
		t.Fatal("expected no event field")
	}

	if !strings.Contains(s, "data: just data") {
		t.Fatal("expected data field")
	}
}

func TestStreamEvents(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	events := make(chan foundation.StreamedEvent, 3)
	events <- foundation.StreamedEvent{Event: "msg", Data: "first"}
	events <- foundation.StreamedEvent{Event: "msg", Data: "second"}
	events <- foundation.StreamedEvent{Data: "third"}
	close(events)

	err := foundation.StreamEvents(context.Background(), rec, events)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatal("expected text/event-stream content type")
	}

	if rec.Header().Get("Cache-Control") != "no-cache" {
		t.Fatal("expected Cache-Control: no-cache")
	}

	body := rec.Body.String()

	if !strings.Contains(body, "data: first") {
		t.Fatal("expected first event in body")
	}

	if !strings.Contains(body, "data: second") {
		t.Fatal("expected second event in body")
	}

	if !strings.Contains(body, "data: third") {
		t.Fatal("expected third event in body")
	}
}

func TestStreamEventsFunc(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	err := foundation.StreamEventsFunc(context.Background(), rec, func(send func(foundation.StreamedEvent)) {
		send(foundation.StreamedEvent{Event: "ping", Data: "1"})
		send(foundation.StreamedEvent{Event: "ping", Data: "2"})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := rec.Body.String()

	if !strings.Contains(body, "data: 1") || !strings.Contains(body, "data: 2") {
		t.Fatal("expected both events in body")
	}
}

func TestStreamEventsStopsWhenContextCancelled(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	events := make(chan foundation.StreamedEvent)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := foundation.StreamEvents(ctx, rec, events)

	if err == nil {
		t.Fatal("expected cancellation error")
	}
}
