package gitwork

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Runner executes git commands with a fixed, hostile-input-safe environment.
type Runner struct {
	// Timeout bounds each invocation. Zero means DefaultTimeout.
	Timeout time.Duration

	// Binary is the git executable, for tests and unusual installs.
	Binary string
}

// DefaultTimeout bounds a single git invocation. A repository on a stalled
// network mount must not hang the whole scan.
const DefaultTimeout = 30 * time.Second

// Run executes git in dir and returns its stdout.
//
// The environment is pinned rather than inherited for two reasons that both
// matter on a machine the user is actively working on: GIT_OPTIONAL_LOCKS=0
// stops a status query from taking the index lock or rewriting the index of a
// repository someone has open, and GIT_TERMINAL_PROMPT=0 guarantees a
// credential prompt can never block a scan waiting for input nobody will type.
func (r Runner) Run(ctx context.Context, dir string, args ...string) (string, error) {
	timeout := r.Timeout

	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	bounded, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	binary := r.Binary

	if binary == "" {
		binary = "git"
	}

	cmd := exec.CommandContext(bounded, binary, append([]string{"-C", dir}, args...)...)

	cmd.Env = append(cmd.Environ(),
		"GIT_OPTIONAL_LOCKS=0",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"GCM_INTERACTIVE=never",
	)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		if errors.Is(bounded.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("%w: git %s in %s", ErrTimeout, strings.Join(args, " "), dir)
		}

		return "", fmt.Errorf("git %s in %s: %w: %s", strings.Join(args, " "), dir, err, strings.TrimSpace(stderr.String()))
	}

	return stdout.String(), nil
}

// Line runs git and returns the first line of output, trimmed. Most plumbing
// queries this package makes answer with exactly one line.
func (r Runner) Line(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := r.Run(ctx, dir, args...)

	if err != nil {
		return "", err
	}

	line, _, _ := strings.Cut(strings.TrimSpace(out), "\n")

	return line, nil
}

// Count runs a git command that answers with a single integer.
func (r Runner) Count(ctx context.Context, dir string, args ...string) (int, error) {
	line, err := r.Line(ctx, dir, args...)

	if err != nil {
		return 0, err
	}

	if line == "" {
		return 0, nil
	}

	count := 0

	if _, err := fmt.Sscanf(line, "%d", &count); err != nil {
		return 0, fmt.Errorf("parse count from %q: %w", line, err)
	}

	return count, nil
}
