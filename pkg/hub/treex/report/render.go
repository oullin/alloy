package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Format is an output encoding.
type Format string

// Renderer writes a report.
type Renderer struct {
	// Verbose lists every entry rather than the largest few. A scan of a
	// developer machine routinely has hundreds, so the default is a summary.
	Verbose bool

	// Explain includes the reasons blocked entries were refused, which is the
	// information that tells the user what to do next.
	Explain bool

	// Limit caps how many entries a category lists when not verbose.
	Limit int
}

const (
	// FormatText is the aligned, human-readable rendering.
	FormatText Format = "text"

	// FormatJSON is the machine-readable rendering, and the reason this
	// package is a stable contract.
	FormatJSON Format = "json"
)

// ParseFormat reads a --format value.
func ParseFormat(raw string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownFormat, raw)
	}
}

// Render writes the report in the requested format.
func (r Renderer) Render(w io.Writer, format Format, report Report) error {
	switch format {
	case FormatJSON:
		return r.renderJSON(w, report)
	case FormatText:
		return r.renderText(w, report)
	default:
		return fmt.Errorf("%w: %q", ErrUnknownFormat, format)
	}
}

func (r Renderer) renderJSON(w io.Writer, report Report) error {
	encoder := json.NewEncoder(w)

	encoder.SetIndent("", "  ")

	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("encode report: %w", err)
	}

	return nil
}

// Sorted returns a provider's entries largest first, so an interrupted or
// limited run still shows what matters most.
func (p Provider) Sorted() []Entry {
	out := make([]Entry, len(p.Entries))

	copy(out, p.Entries)

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Bytes > out[j].Bytes
	})

	return out
}
