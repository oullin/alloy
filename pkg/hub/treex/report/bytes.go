package report

import "fmt"

// Bytes is a byte count that knows how to print itself.
type Bytes int64

const unit = 1024

// String renders a human-readable size using binary units, matching what du -h
// and ls -h print so a treex figure can be compared with them directly.
func (b Bytes) String() string {
	if b < unit {
		return fmt.Sprintf("%d B", int64(b))
	}

	value := float64(b)
	suffixes := [...]string{"KB", "MB", "GB", "TB", "PB"}

	index := -1

	for value >= unit && index < len(suffixes)-1 {
		value /= unit
		index++
	}

	if value >= 100 {
		return fmt.Sprintf("%.0f %s", value, suffixes[index])
	}

	return fmt.Sprintf("%.1f %s", value, suffixes[index])
}

// Int64 returns the raw count.
func (b Bytes) Int64() int64 {
	return int64(b)
}
