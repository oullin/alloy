package observability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"alloy.dev/go/auth/observability"
	"alloy.dev/go/auth/user"
	cauth "alloy.dev/go/contracts/auth"
	clog "alloy.dev/go/contracts/auth/log"
)

type staticGuard struct {
	user cauth.User
}

type recordingLogger struct {
	entries []map[string]any
}

func TestMiddlewareAddsRequestContextAndLogs(t *testing.T) {
	user := user.NewGenericUser(map[string]any{"id": "1"})
	guard := staticGuard{user: user}
	logger := &recordingLogger{}
	middleware := observability.Middleware(logger, guard, func(*http.Request) string { return "req-1" })

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		meta, ok := observability.RequestContextFromContext(r.Context())

		if !ok {
			t.Fatal("expected request context")
		}

		if meta.RequestID != "req-1" || meta.UserID != "1" || meta.Path != "/secure" {
			t.Fatalf("meta = %#v", meta)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.Header.Set("X-Real-IP", "127.0.0.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}

	if len(logger.entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(logger.entries))
	}

	if logger.entries[0]["request_id"] != "req-1" {
		t.Fatalf("log context = %#v", logger.entries[0])
	}
}

func (g staticGuard) User(context.Context) (cauth.User, error) { return g.user, nil }
func (g staticGuard) Check(context.Context) bool               { return g.user != nil }
func (g staticGuard) Guest(context.Context) bool               { return g.user == nil }
func (g staticGuard) ID(context.Context) any {
	if g.user == nil {
		return nil
	}

	return g.user.GetAuthIdentifier()
}

func (l *recordingLogger) Emergency(message string, context ...map[string]any) {
	l.Log(clog.LevelEmergency, message, context...)
}
func (l *recordingLogger) Alert(message string, context ...map[string]any) {
	l.Log(clog.LevelAlert, message, context...)
}
func (l *recordingLogger) Critical(message string, context ...map[string]any) {
	l.Log(clog.LevelCritical, message, context...)
}
func (l *recordingLogger) Error(message string, context ...map[string]any) {
	l.Log(clog.LevelError, message, context...)
}
func (l *recordingLogger) Warning(message string, context ...map[string]any) {
	l.Log(clog.LevelWarning, message, context...)
}
func (l *recordingLogger) Notice(message string, context ...map[string]any) {
	l.Log(clog.LevelNotice, message, context...)
}
func (l *recordingLogger) Info(message string, context ...map[string]any) {
	l.Log(clog.LevelInfo, message, context...)
}
func (l *recordingLogger) Debug(message string, context ...map[string]any) {
	l.Log(clog.LevelDebug, message, context...)
}
func (l *recordingLogger) Log(_ clog.Level, _ string, context ...map[string]any) {
	if len(context) == 0 {
		l.entries = append(l.entries, map[string]any{})

		return
	}

	l.entries = append(l.entries, context[0])
}
