package server_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oullin/alloy/api/routing"
	"github.com/oullin/alloy/api/server"
)

type contextKey struct{}

func TestNewHandlerDispatchesRouter(t *testing.T) {
	router := routing.NewRouter(nil, nil)
	router.Get("/health", func() any {
		return "ok"
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	server.NewHandler(router).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "ok")
	}
}

func TestNewReturnsConfiguredHTTPServer(t *testing.T) {
	baseContext := func(net.Listener) context.Context {
		return context.WithValue(context.Background(), contextKey{}, "alloy")
	}

	router := routing.NewRouter(nil, nil)
	cfg := server.Config{
		Addr:              ":9090",
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       4 * time.Second,
		BaseContext:       baseContext,
	}

	srv := server.New(router, cfg)

	if srv.Addr != cfg.Addr {
		t.Fatalf("Addr = %q, want %q", srv.Addr, cfg.Addr)
	}

	if srv.Handler == nil {
		t.Fatal("Handler is nil")
	}

	if srv.ReadTimeout != cfg.ReadTimeout {
		t.Fatalf("ReadTimeout = %s, want %s", srv.ReadTimeout, cfg.ReadTimeout)
	}

	if srv.ReadHeaderTimeout != cfg.ReadHeaderTimeout {
		t.Fatalf("ReadHeaderTimeout = %s, want %s", srv.ReadHeaderTimeout, cfg.ReadHeaderTimeout)
	}

	if srv.WriteTimeout != cfg.WriteTimeout {
		t.Fatalf("WriteTimeout = %s, want %s", srv.WriteTimeout, cfg.WriteTimeout)
	}

	if srv.IdleTimeout != cfg.IdleTimeout {
		t.Fatalf("IdleTimeout = %s, want %s", srv.IdleTimeout, cfg.IdleTimeout)
	}

	if srv.BaseContext == nil {
		t.Fatal("BaseContext is nil")
	}
}
