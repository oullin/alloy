// Command treex reclaims the disk space AI coding agents leave behind. The
// command surface lives in internal/app; this entrypoint only carries the
// version stamped in by -X main.version and the signal handling that lets a
// multi-hundred-gigabyte scan be interrupted cleanly.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"hara.sh/alloy/treex/internal/app"
)

var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// os.Exit skips deferred calls, so release the signal handler explicitly
	// before exiting with the captured code.
	code := app.
		CLI(version, os.Stdout, os.Stderr).
		Dispatch(ctx, os.Args[1:])

	stop()
	os.Exit(code)
}
