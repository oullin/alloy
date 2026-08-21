package app

import (
	"errors"
	"fmt"

	"hara.sh/alloy/treex/config"
)

const (
	// exitOK is a clean run.
	exitOK = 0

	// exitError is a runtime failure.
	exitError = 1

	// exitUsage is a malformed invocation. It matches the Set's ErrExit so an
	// unknown flag and an unknown subcommand report the same way.
	exitUsage = 2
)

// reportError is the single funnel from error to exit code. Having exactly one
// means every failure is printed the same way and no handler has to remember
// the convention.
func (d *deps) reportError(err error) int {
	if err == nil {
		return exitOK
	}

	// An abort is the user declining a prompt, which is a successful outcome
	// for a tool whose default posture is to not delete.
	if errors.Is(err, errAborted) {
		_, _ = fmt.Fprintln(d.stderr, "treex: aborted; nothing was removed")

		return exitOK
	}

	_, _ = fmt.Fprintf(d.stderr, "treex: %v\n", err)

	return exitError
}

// usageError reports a malformed invocation and prints the usage, because a
// user who mistyped a flag needs to see the list of real ones.
func (d *deps) usageError(err error) int {
	_, _ = fmt.Fprintf(d.stderr, "treex: %v\n\n", err)

	if d.usage != nil {
		d.usage(d.stderr)
	}

	return exitUsage
}

// configError adds the config path to a load failure, which is nearly always
// the thing the user needs to know.
func configError(err error, path string) error {
	if path == "" || errors.Is(err, config.ErrNotFound) {
		return err
	}

	return fmt.Errorf("%s: %w", path, err)
}
