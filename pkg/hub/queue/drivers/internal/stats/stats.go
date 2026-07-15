// Package stats parses the string-keyed counters that queue backends report.
//
// SQS reports queue attributes as map[string]string, and any backend whose
// stats arrive the same way can share this lenient integer parse.
package stats

import "strconv"

// ParseInt reads key from stats as an int64. A missing key or an unparsable
// value yields 0: these counters are advisory, and a backend that reports
// garbage should not turn a size query into an error.
func ParseInt(stats map[string]string, key string) int64 {
	v, ok := stats[key]

	if !ok {
		return 0
	}

	n, err := strconv.ParseInt(v, 10, 64)

	if err != nil {
		return 0
	}

	return n
}
