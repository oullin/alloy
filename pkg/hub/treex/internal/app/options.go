package app

import (
	"fmt"
	"strconv"
	"strings"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/report"
)

// options is every flag the scan and clean commands share, parsed by hand.
//
// A framework would add a dependency and a lot of reflection to a surface this
// small and this fixed, on a binary whose job is deleting files.
type options struct {
	configPath string
	providers  []string
	categories []string
	root       string
	format     report.Format

	jobs      int
	limit     int
	minSize   config.Size
	olderThan config.Duration

	hasMinSize   bool
	hasOlderThan bool

	apply   bool
	yes     bool
	force   bool
	explain bool
	verbose bool
}

// parseOptions reads the shared flags. Unknown flags are an error rather than
// being ignored: a mistyped --force on a command that deletes must never be
// silently dropped.
func parseOptions(args []string) (options, error) {
	opts := options{format: report.FormatText}

	for index := 0; index < len(args); index++ {
		arg := args[index]

		value := func(name string) (string, error) {
			if index+1 >= len(args) {
				return "", fmt.Errorf("flag %s needs a value", name)
			}

			index++

			return args[index], nil
		}

		var err error

		switch arg {
		case "--config":
			opts.configPath, err = value(arg)
		case "--providers", "--provider":
			var raw string

			if raw, err = value(arg); err == nil {
				opts.providers = append(opts.providers, split(raw)...)
			}
		case "--categories", "--category":
			var raw string

			if raw, err = value(arg); err == nil {
				opts.categories = append(opts.categories, split(raw)...)
			}
		case "--root":
			opts.root, err = value(arg)
		case "--format":
			var raw string

			if raw, err = value(arg); err == nil {
				opts.format, err = report.ParseFormat(raw)
			}
		case "--jobs":
			var raw string

			if raw, err = value(arg); err == nil {
				opts.jobs, err = strconv.Atoi(raw)
			}
		case "--limit":
			var raw string

			if raw, err = value(arg); err == nil {
				opts.limit, err = strconv.Atoi(raw)
			}
		case "--min-size":
			var raw string

			if raw, err = value(arg); err == nil {
				opts.minSize, err = config.ParseSize(raw)
				opts.hasMinSize = true
			}
		case "--older-than":
			var raw string

			if raw, err = value(arg); err == nil {
				opts.olderThan, err = config.ParseDuration(raw)
				opts.hasOlderThan = true
			}
		case "--apply":
			opts.apply = true
		case "--yes", "-y":
			opts.yes = true
		case "--force":
			opts.force = true
		case "--explain":
			opts.explain = true
		case "--verbose", "-v":
			opts.verbose = true
		default:
			return options{}, fmt.Errorf("unknown flag - {%q}", arg)
		}

		if err != nil {
			return options{}, err
		}
	}

	return opts, nil
}

// apply folds the command-line overrides into the loaded configuration, so
// everything downstream reads one resolved source of truth.
func (o options) applyTo(cfg config.Config) config.Config {
	if o.jobs > 0 {
		cfg.Defaults.Jobs = o.jobs
	}

	if o.hasMinSize {
		cfg.Defaults.MinSize = o.minSize
	}

	if o.hasOlderThan {
		cfg.Defaults.OlderThan = o.olderThan
	}

	return cfg
}

func split(raw string) []string {
	out := make([]string, 0, 4)

	for _, item := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(item)

		if trimmed != "" {
			out = append(out, trimmed)
		}
	}

	return out
}
