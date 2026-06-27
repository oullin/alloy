package foundation

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// StreamedEvent represents a single server-sent event (SSE).
type StreamedEvent struct {
	Event string // event type (optional)
	Data  string // event payload
	ID    string // event ID (optional)
	Retry int    // reconnection time in milliseconds (optional, 0 = omit)
}

// String formats the event according to the SSE specification.
func (e *StreamedEvent) String() string {
	var b strings.Builder

	if e.ID != "" {
		fmt.Fprintf(&b, "id: %s\n", e.ID)
	}

	if e.Event != "" {
		fmt.Fprintf(&b, "event: %s\n", e.Event)
	}

	if e.Retry > 0 {
		fmt.Fprintf(&b, "retry: %d\n", e.Retry)
	}

	// Data may contain multiple lines; each line must be prefixed with "data: ".
	lines := strings.Split(e.Data, "\n")

	for _, line := range lines {
		fmt.Fprintf(&b, "data: %s\n", line)
	}

	b.WriteString("\n")

	return b.String()
}

// StreamEvents writes a series of server-sent events from the channel to the
// http.ResponseWriter. It sets the appropriate SSE headers, flushes after each
// event, and returns when the channel is closed or the client disconnects.
func StreamEvents(ctx context.Context, w http.ResponseWriter, events <-chan StreamedEvent) error {
	flusher, ok := w.(http.Flusher)

	if !ok {
		return fmt.Errorf("httpx: response writer does not support flushing")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	if ctx == nil {
		ctx = context.Background()
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-events:
			if !ok {
				return nil
			}

			_, err := fmt.Fprint(w, event.String())

			if err != nil {
				return err
			}

			flusher.Flush()
		}
	}
}

// StreamEventsFunc writes server-sent events produced by a callback function.
// The callback receives a send function; call it for each event. Return from
// the callback to end the stream.
func StreamEventsFunc(ctx context.Context, w http.ResponseWriter, fn func(send func(StreamedEvent))) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ch := make(chan StreamedEvent)

	go func() {
		defer close(ch)

		fn(func(event StreamedEvent) {
			select {
			case <-ctx.Done():
				return
			case ch <- event:
			}
		})
	}()

	return StreamEvents(ctx, w, ch)
}
