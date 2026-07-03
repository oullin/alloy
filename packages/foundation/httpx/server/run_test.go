package server_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/oullin/alloy/packages/foundation/httpx/server"
)

func TestRunListenerServesRequestAndStopsOnCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	runErr, baseURL := runListener(t, ctx, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}), time.Second)

	body := getBody(t, baseURL)
	if body != "ok" {
		t.Fatalf("body = %q, want %q", body, "ok")
	}

	cancel()

	if err := waitRun(t, runErr, time.Second); err != nil {
		t.Fatalf("RunListener error = %v, want nil", err)
	}
}

func TestRunListenerLetsInFlightRequestCompleteDuringShutdown(t *testing.T) {
	t.Parallel()

	entered := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	runErr, baseURL := runListener(t, ctx, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(entered)
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte("done"))
	}), time.Second)

	response := make(chan clientResult, 1)
	go func() {
		response <- getBodyResult(baseURL)
	}()

	waitSignal(t, entered, time.Second, "handler did not start")
	cancel()

	result := waitClient(t, response, time.Second)
	if result.err != nil {
		t.Fatalf("request error = %v, want nil", result.err)
	}

	if result.body != "done" {
		t.Fatalf("body = %q, want %q", result.body, "done")
	}

	if err := waitRun(t, runErr, time.Second); err != nil {
		t.Fatalf("RunListener error = %v, want nil", err)
	}
}

func TestRunListenerClosesWhenShutdownTimeoutExpires(t *testing.T) {
	t.Parallel()

	entered := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	runErr, baseURL := runListener(t, ctx, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	}), 20*time.Millisecond)

	response := make(chan clientResult, 1)
	go func() {
		response <- getBodyResult(baseURL)
	}()

	waitSignal(t, entered, time.Second, "handler did not start")

	started := time.Now()
	cancel()

	err := waitRun(t, runErr, time.Second)
	if err == nil {
		t.Fatal("RunListener error = nil, want shutdown timeout error")
	}

	if time.Since(started) > 500*time.Millisecond {
		t.Fatalf("RunListener returned too slowly after shutdown timeout: %s", time.Since(started))
	}

	if !strings.Contains(err.Error(), "graceful shutdown failed") {
		t.Fatalf("RunListener error = %v, want wrapped shutdown error", err)
	}

	_ = waitClient(t, response, time.Second)
}

func TestRunListenerReturnsWrappedServeError(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	err = server.RunListener(context.Background(), &http.Server{}, ln, time.Second)
	if err == nil {
		t.Fatal("RunListener error = nil, want serve error")
	}

	if errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("RunListener error = %v, want serve error", err)
	}

	if !strings.Contains(err.Error(), "serve failed") {
		t.Fatalf("RunListener error = %v, want wrapped serve error", err)
	}
}

type clientResult struct {
	body string
	err  error
}

func runListener(t *testing.T, ctx context.Context, handler http.Handler, shutdownTimeout time.Duration) (<-chan error, string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := &http.Server{
		Handler: handler,
	}

	runErr := make(chan error, 1)
	go func() {
		runErr <- server.RunListener(ctx, srv, ln, shutdownTimeout)
	}()

	return runErr, fmt.Sprintf("http://%s", ln.Addr())
}

func getBody(t *testing.T, url string) string {
	t.Helper()

	result := getBodyResult(url)
	if result.err != nil {
		t.Fatalf("GET %s: %v", url, result.err)
	}

	return result.body
}

func getBodyResult(url string) clientResult {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return clientResult{err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return clientResult{err: err}
	}

	return clientResult{body: string(body)}
}

func waitRun(t *testing.T, runErr <-chan error, timeout time.Duration) error {
	t.Helper()

	select {
	case err := <-runErr:
		return err
	case <-time.After(timeout):
		t.Fatalf("RunListener did not return within %s", timeout)
		return nil
	}
}

func waitClient(t *testing.T, response <-chan clientResult, timeout time.Duration) clientResult {
	t.Helper()

	select {
	case result := <-response:
		return result
	case <-time.After(timeout):
		t.Fatalf("client did not return within %s", timeout)
		return clientResult{}
	}
}

func waitSignal(t *testing.T, signal <-chan struct{}, timeout time.Duration, message string) {
	t.Helper()

	select {
	case <-signal:
	case <-time.After(timeout):
		t.Fatal(message)
	}
}
