package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/internal/console"
)

// runProviders lists what treex would sweep and where, which is the fastest way
// to check a configuration without running a scan.
func (d *deps) runProviders(_ context.Context, args []string) int {
	opts, err := parseOptions(args)

	if err != nil {
		return d.usageError(err)
	}

	loader := config.Loader{Home: d.home, Env: d.env, Explicit: opts.configPath}

	cfg, path, err := loader.Load()

	if err != nil {
		return d.reportError(configError(err, opts.configPath))
	}

	printer := console.NewPrinter(d.stdout, console.DetectColor(d.stdout))

	printer.Section("providers")

	if path != "" {
		printer.Detail("config", path)
	} else {
		printer.Detail("config", "(built-in defaults)")
	}

	printer.Blank()

	for _, provider := range cfg.Providers {
		resolved := provider.Resolve(d.home, "")
		state := "disabled"

		if provider.Enabled {
			state = "enabled"
		}

		present := "missing"

		if info, err := os.Stat(resolved.Root); err == nil && info.IsDir() {
			present = "present"
		}

		printer.Line(fmt.Sprintf("%-16s %-9s %-8s %-6s %s",
			provider.Name, state, provider.Kind, present, resolved.Root))
	}

	return exitOK
}

// runConfig handles the config subcommands.
func (d *deps) runConfig(_ context.Context, args []string) int {
	if len(args) == 0 {
		return d.usageError(fmt.Errorf("config needs one of - {%q}", "show|path|init"))
	}

	loader := config.Loader{Home: d.home, Env: d.env}

	switch args[0] {
	case "path":
		return d.configPath(loader)
	case "show":
		return d.configShow(loader)
	case "init":
		return d.configInit(loader)
	default:
		return d.usageError(fmt.Errorf("unknown config subcommand - {%q}", args[0]))
	}
}

func (d *deps) configPath(loader config.Loader) int {
	path, found, err := loader.DiscoverPath()

	if err != nil {
		return d.reportError(err)
	}

	if !found {
		_, _ = fmt.Fprintf(d.stdout, "%s (does not exist yet)\n", loader.DefaultPath())

		return exitOK
	}

	_, _ = fmt.Fprintln(d.stdout, path)

	return exitOK
}

func (d *deps) configShow(loader config.Loader) int {
	_, path, err := loader.Load()

	if err != nil {
		return d.reportError(err)
	}

	if path == "" {
		_, _ = fmt.Fprint(d.stdout, config.Template)

		return exitOK
	}

	contents, err := os.ReadFile(path)

	if err != nil {
		return d.reportError(fmt.Errorf("read %s: %w", path, err))
	}

	_, _ = d.stdout.Write(contents)

	return exitOK
}

// configInit writes a starter file. It refuses to overwrite: a config that
// disables a provider is exactly the kind of thing a user would be annoyed to
// lose to a stray command.
func (d *deps) configInit(loader config.Loader) int {
	path := loader.DefaultPath()

	if _, err := os.Stat(path); err == nil {
		return d.reportError(fmt.Errorf("%s already exists", path))
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return d.reportError(fmt.Errorf("create config directory: %w", err))
	}

	if err := os.WriteFile(path, []byte(config.Template), 0o644); err != nil {
		return d.reportError(fmt.Errorf("write %s: %w", path, err))
	}

	_, _ = fmt.Fprintf(d.stdout, "wrote %s\n", path)

	return exitOK
}
