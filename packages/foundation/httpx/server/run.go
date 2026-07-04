package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const defaultShutdownTimeout = 5 * time.Second

// Run serves srv with ListenAndServe until it fails or ctx is canceled.
//
// A non-positive shutdownTimeout uses a default of 5 seconds.
func Run(ctx context.Context, srv *http.Server, shutdownTimeout time.Duration) error {
	if srv == nil {
		return errors.New("server: nil http server")
	}

	return run(ctx, srv, shutdownTimeout, srv.ListenAndServe)
}

// RunListener serves srv on ln until serving fails or ctx is canceled.
//
// A non-positive shutdownTimeout uses a default of 5 seconds.
func RunListener(ctx context.Context, srv *http.Server, ln net.Listener, shutdownTimeout time.Duration) error {
	if srv == nil {
		return errors.New("server: nil http server")
	}

	return run(ctx, srv, shutdownTimeout, func() error {
		return srv.Serve(ln)
	})
}

func run(ctx context.Context, srv *http.Server, shutdownTimeout time.Duration, serve func() error) error {
	if shutdownTimeout <= 0 {
		shutdownTimeout = defaultShutdownTimeout
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- serve()
	}()

	select {
	case err := <-serveErr:
		return normalizeServeError(err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()

		return fmt.Errorf("server: graceful shutdown failed: %w", err)
	}

	return normalizeServeError(<-serveErr)
}

func normalizeServeError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return fmt.Errorf("server: serve failed: %w", err)
}
