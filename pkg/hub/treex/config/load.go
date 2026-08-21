package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// Loader discovers and reads configuration. Home and Env are injected rather
// than read from the process so tests can drive discovery without touching the
// real environment, and so the Docker entrypoint can point discovery at a
// mounted root.
type Loader struct {
	Home     string
	Env      func(string) string
	Explicit string
}

// EnvVar names an environment variable holding a config path, checked after an
// explicit --config and before the XDG location.
const EnvVar = "TREEX_CONFIG"

// DiscoverPath returns the config file that would be read, and whether one
// exists at all. An empty path with no error means "no file; use the built-in
// defaults", which is the expected state for a first run.
func (l Loader) DiscoverPath() (string, bool, error) {
	if l.Explicit != "" {
		path := l.expand(l.Explicit)

		if !exists(path) {
			return "", false, fmt.Errorf("%w: %s", ErrNotFound, path)
		}

		return path, true, nil
	}

	if fromEnv := strings.TrimSpace(l.env(EnvVar)); fromEnv != "" {
		path := l.expand(fromEnv)

		if !exists(path) {
			return "", false, fmt.Errorf("%w: %s (from %s)", ErrNotFound, path, EnvVar)
		}

		return path, true, nil
	}

	for _, candidate := range l.searchPaths() {
		if exists(candidate) {
			return candidate, true, nil
		}
	}

	return "", false, nil
}

// DefaultPath is where treex config init writes, and what treex config path
// prints when no file exists yet.
func (l Loader) DefaultPath() string {
	return l.searchPaths()[0]
}

func (l Loader) searchPaths() []string {
	base := strings.TrimSpace(l.env("XDG_CONFIG_HOME"))

	if base == "" {
		base = filepath.Join(l.Home, ".config")
	}

	return []string{
		filepath.Join(base, "treex", "config.yml"),
		filepath.Join(base, "treex", "config.yaml"),
		filepath.Join(l.Home, ".treex.yml"),
	}
}

// Load reads the discovered configuration layered onto Default(), and returns
// the merged result along with the path it read (empty when none was found).
func (l Loader) Load() (Config, string, error) {
	path, found, err := l.DiscoverPath()

	if err != nil {
		return Config{}, "", err
	}

	base := Default()

	if !found {
		if err := Validate(&base); err != nil {
			return Config{}, "", err
		}

		return base, "", nil
	}

	file, seen, err := parse(path)

	if err != nil {
		return Config{}, "", err
	}

	if file.Version != 0 && file.Version != Version {
		return Config{}, "", fmt.Errorf("%w: %d (treex understands %d)", ErrUnsupportedVersion, file.Version, Version)
	}

	merged := Merge(base, file, seen)

	if err := Validate(&merged); err != nil {
		return Config{}, "", err
	}

	return merged, path, nil
}

func parse(path string) (Config, KeySet, error) {
	v := viper.New()

	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return Config{}, nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var file Config

	decode := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.TextUnmarshallerHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))

	if err := v.Unmarshal(&file, decode); err != nil {
		return Config{}, nil, fmt.Errorf("parse config %s: %w", path, err)
	}

	return file, NewKeySet(v.AllSettings()), nil
}

func (l Loader) env(name string) string {
	if l.Env == nil {
		return os.Getenv(name)
	}

	return l.Env(name)
}

// expand resolves a leading ~ and any environment variables. It is deliberately
// narrow: only a leading ~ is a home reference, so a path that legitimately
// contains a tilde mid-string survives intact.
func (l Loader) expand(path string) string {
	expanded := os.Expand(path, func(name string) string {
		return l.env(name)
	})

	if expanded == "~" {
		return l.Home
	}

	if strings.HasPrefix(expanded, "~/") {
		expanded = filepath.Join(l.Home, expanded[2:])
	}

	return filepath.Clean(expanded)
}

// Expand resolves a leading ~ and environment variables against this loader's
// home, and returns an absolute, cleaned path.
func (l Loader) Expand(path string) string {
	return l.expand(path)
}

func exists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
