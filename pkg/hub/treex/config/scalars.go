package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Size is a byte count written with an optional unit suffix ("500MB", "2GiB",
// "1024"). It is a named type with UnmarshalText so the same spelling works in
// a YAML file and in a --min-size flag.
type Size int64

// Duration is an age written either as a Go duration ("90m") or with a day or
// week suffix ("7d", "2w"), which time.ParseDuration does not accept but which
// is the only unit anyone reaches for when talking about stale worktrees.
type Duration time.Duration

var sizeUnits = []struct {
	suffix string
	scale  int64
}{
	{"KIB", 1 << 10},
	{"MIB", 1 << 20},
	{"GIB", 1 << 30},
	{"TIB", 1 << 40},
	{"KB", 1000},
	{"MB", 1000 * 1000},
	{"GB", 1000 * 1000 * 1000},
	{"TB", 1000 * 1000 * 1000 * 1000},
	{"K", 1 << 10},
	{"M", 1 << 20},
	{"G", 1 << 30},
	{"T", 1 << 40},
	{"B", 1},
}

var durationUnits = []struct {
	suffix string
	scale  time.Duration
}{
	{"d", 24 * time.Hour},
	{"w", 7 * 24 * time.Hour},
}

// ParseSize reads a byte count with an optional unit suffix. Decimal units (KB,
// MB) are powers of 1000 and binary units (KiB, MiB) are powers of 1024;
// single-letter suffixes are binary, matching what du and ls -h print.
func ParseSize(text string) (Size, error) {
	raw := strings.ToUpper(strings.TrimSpace(text))

	if raw == "" {
		return 0, fmt.Errorf("%w: empty value", ErrInvalidSize)
	}

	for _, unit := range sizeUnits {
		if !strings.HasSuffix(raw, unit.suffix) {
			continue
		}

		digits := strings.TrimSpace(strings.TrimSuffix(raw, unit.suffix))

		value, err := strconv.ParseFloat(digits, 64)

		if err != nil {
			return 0, fmt.Errorf("%w: %q", ErrInvalidSize, text)
		}

		if value < 0 {
			return 0, fmt.Errorf("%w: %q is negative", ErrInvalidSize, text)
		}

		return Size(value * float64(unit.scale)), nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)

	if err != nil || value < 0 {
		return 0, fmt.Errorf("%w: %q", ErrInvalidSize, text)
	}

	return Size(value), nil
}

// UnmarshalText lets viper and flag parsing share one spelling of a size.
func (s *Size) UnmarshalText(text []byte) error {
	parsed, err := ParseSize(string(text))

	if err != nil {
		return err
	}

	*s = parsed

	return nil
}

// Bytes returns the size as a plain byte count.
func (s Size) Bytes() int64 {
	return int64(s)
}

// ParseDuration reads a Go duration, or a count suffixed with d (days) or w
// (weeks). Days and weeks are the units that matter for "how stale is this
// worktree" and time.ParseDuration rejects both.
func ParseDuration(text string) (Duration, error) {
	raw := strings.ToLower(strings.TrimSpace(text))

	if raw == "" {
		return 0, fmt.Errorf("%w: empty value", ErrInvalidDuration)
	}

	for _, unit := range durationUnits {
		if !strings.HasSuffix(raw, unit.suffix) {
			continue
		}

		digits := strings.TrimSuffix(raw, unit.suffix)

		value, err := strconv.ParseFloat(digits, 64)

		if err != nil {
			continue
		}

		if value < 0 {
			return 0, fmt.Errorf("%w: %q is negative", ErrInvalidDuration, text)
		}

		return Duration(float64(unit.scale) * value), nil
	}

	parsed, err := time.ParseDuration(raw)

	if err != nil {
		return 0, fmt.Errorf("%w: %q", ErrInvalidDuration, text)
	}

	if parsed < 0 {
		return 0, fmt.Errorf("%w: %q is negative", ErrInvalidDuration, text)
	}

	return Duration(parsed), nil
}

// UnmarshalText lets viper and flag parsing share one spelling of a duration.
func (d *Duration) UnmarshalText(text []byte) error {
	parsed, err := ParseDuration(string(text))

	if err != nil {
		return err
	}

	*d = parsed

	return nil
}

// Std returns the duration as a time.Duration.
func (d Duration) Std() time.Duration {
	return time.Duration(d)
}
