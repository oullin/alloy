// Package app is treex's sole composition root. It owns the writers, the
// resolved configuration, and the concrete collaborators every command needs,
// and it is the only place those are constructed.
//
// Nothing self-registers: the command table is built explicitly in New, so
// reading one function tells you the entire surface of the binary.
package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"hara.sh/alloy/treex/internal/command"
)

// deps carries everything the handlers need. Handlers are methods on it rather
// than closures over package state, so there is no global to reset between
// tests and no ordering dependency at startup.
type deps struct {
	version string
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	home    string
	env     func(string) string

	// usage is wired in after the command Set exists, because the commands need
	// to be able to print it and the Set needs the commands.
	usage func(io.Writer)
}

const header = `treex - reclaim the disk space AI coding agents leave behind

usage:
  treex <command> [options]

commands:
`

const shared = `
options:
  --config PATH        Read this configuration file
  --providers LIST     Limit to these providers (comma separated)
  --categories LIST    Limit to worktree, artifact, cache, session, orphan
  --root PATH          Treat PATH as the home directory (for mounted volumes)
  --older-than AGE     Ignore anything touched more recently (e.g. 7d)
  --min-size SIZE      Ignore anything smaller (e.g. 500MB)
  --format FORMAT      text (default) or json
  --jobs N             Worker count for the size walk
  --limit N            Act on at most N entries, largest first
  --explain            Say why each blocked entry was refused
  --verbose            List every entry instead of the largest few
`

// CLI builds the command table against the real process.
func CLI(version string, stdout, stderr io.Writer) command.Set {
	return New(version, os.Stdin, stdout, stderr, os.Getenv, homeDir())
}

// New builds the command table with every input injected, which is how tests
// drive the binary without touching process state.
func New(
	version string,
	stdin io.Reader,
	stdout, stderr io.Writer,
	env func(string) string,
	home string,
) command.Set {
	d := &deps{
		version: version,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		home:    home,
		env:     env,
	}

	set := command.Set{
		Name:    "treex",
		Header:  header,
		ErrExit: exitUsage,
		Stderr:  stderr,
		Commands: []command.Command{
			{
				Name:  "scan",
				Usage: "  scan                 Measure what could be reclaimed, changing nothing\n",
				Run:   d.runScan,
			},
			{
				Name: "clean",
				Usage: "  clean                Remove reclaimable debris (dry run unless --apply)\n" +
					"                       --apply   actually delete; prompts unless --yes\n" +
					"                       --force   accept losing uncommitted or unpushed work\n",
				Run: d.runClean,
			},
			{
				Name:  "doctor",
				Usage: "  doctor               Report stale worktree registries and blocked trees\n",
				Run:   d.runDoctor,
			},
			{
				Name:  "providers",
				Usage: "  providers            List the agents treex knows about and their roots\n",
				Run:   d.runProviders,
			},
			{
				Name:  "config",
				Usage: "  config <show|path|init>\n                       Inspect or create the configuration file\n",
				Run:   d.runConfig,
			},
			{
				Name:    "version",
				Aliases: []string{"--version", "-version"},
				Usage:   "  version              Print the version this binary was built as\n" + shared,
				Run:     d.runVersion,
			},
		},
	}

	d.usage = func(w io.Writer) {
		set.PrintUsage(w)
	}

	return set
}

func (d *deps) runVersion(_ context.Context, _ []string) int {
	_, _ = fmt.Fprintln(d.stdout, d.version)

	return exitOK
}

func homeDir() string {
	home, err := os.UserHomeDir()

	if err != nil {
		return ""
	}

	return home
}
