package console

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// ColorMode is whether a Printer emits ANSI escape sequences.
type ColorMode int

// Printer renders report output to a writer. The palette fields are empty
// strings when color is off, so the same format strings render plain text.
type Printer struct {
	w io.Writer

	bold   string
	dim    string
	cyan   string
	green  string
	yellow string
	red    string
	reset  string
}

const (
	// ColorAuto defers the decision to DetectColor. NewPrinter treats it as
	// no-color, so callers resolve it through DetectColor before constructing a
	// Printer rather than passing it through.
	ColorAuto ColorMode = iota

	// ColorAlways forces ANSI color on.
	ColorAlways

	// ColorNever forces ANSI color off.
	ColorNever
)

// DetectColor resolves whether color should be used when writing to w. It
// honors FORCE_COLOR (always on) and NO_COLOR (always off) before falling back
// to whether w is a terminal. This is the single place the environment is read.
func DetectColor(w io.Writer) ColorMode {
	if os.Getenv("FORCE_COLOR") != "" {
		return ColorAlways
	}

	if os.Getenv("NO_COLOR") != "" {
		return ColorNever
	}

	if file, ok := w.(*os.File); ok && isatty.IsTerminal(file.Fd()) {
		return ColorAlways
	}

	return ColorNever
}

// IsTerminal reports whether r is an interactive terminal. clean --apply uses
// it to decide between prompting and refusing: a non-interactive run that was
// not given --yes must not silently delete.
func IsTerminal(r io.Reader) bool {
	file, ok := r.(*os.File)

	if !ok {
		return false
	}

	return isatty.IsTerminal(file.Fd())
}

// NewPrinter builds a Printer writing to w. ANSI color is enabled only for
// ColorAlways; ColorAuto and ColorNever both render plain text, so callers pass
// the resolved result of DetectColor.
func NewPrinter(w io.Writer, mode ColorMode) *Printer {
	p := &Printer{w: w}

	if mode == ColorAlways {
		p.bold = "\033[1m"
		p.dim = "\033[2m"
		p.cyan = "\033[36m"
		p.green = "\033[32m"
		p.yellow = "\033[33m"
		p.red = "\033[31m"
		p.reset = "\033[0m"
	}

	return p
}

// Writer exposes the underlying writer so callers can emit raw text (a rendered
// table, a prompt) without reaching around the Printer for the destination.
func (p *Printer) Writer() io.Writer {
	return p.w
}

// Section prints a bold, cyan-arrowed section header preceded by a blank line.
func (p *Printer) Section(msg string) {
	_, _ = fmt.Fprintf(p.w, "\n%s==>%s %s%s%s\n", p.cyan, p.reset, p.bold, msg, p.reset)
}

// SectionTotal prints a section header with a right-hand value, used for the
// per-provider headers that carry the provider's total size.
func (p *Printer) SectionTotal(msg, total string) {
	_, _ = fmt.Fprintf(p.w, "\n%s==>%s %s%-52s%s %s\n", p.cyan, p.reset, p.bold, msg, p.reset, total)
}

// Detail prints an aligned label/value line under the current section.
func (p *Printer) Detail(label, value string) {
	_, _ = fmt.Fprintf(p.w, "    %s%-12s%s %s\n", p.dim, label, p.reset, value)
}

// Line prints an indented, uncolored line under the current section.
func (p *Printer) Line(msg string) {
	_, _ = fmt.Fprintf(p.w, "    %s\n", msg)
}

// Dim prints an indented, dimmed line under the current section.
func (p *Printer) Dim(msg string) {
	_, _ = fmt.Fprintf(p.w, "    %s%s%s\n", p.dim, msg, p.reset)
}

// Warning prints a yellow, indented advisory. Blocked candidates use it: they
// are an expected outcome, not an error.
func (p *Printer) Warning(msg string) {
	_, _ = fmt.Fprintf(p.w, "    %s%s%s\n", p.yellow, msg, p.reset)
}

// Success prints a green, indented confirmation.
func (p *Printer) Success(msg string) {
	_, _ = fmt.Fprintf(p.w, "    %s%s%s\n", p.green, msg, p.reset)
}

// Failure prints a red, banged failure banner preceded by a blank line.
func (p *Printer) Failure(msg string) {
	_, _ = fmt.Fprintf(p.w, "\n%s!!%s %s%s%s\n", p.red, p.reset, p.bold, msg, p.reset)
}

// Blank prints an empty line.
func (p *Printer) Blank() {
	_, _ = io.WriteString(p.w, "\n")
}

// Prompt writes a question and reads a single line of input, reporting whether
// the answer was an affirmative y/yes. Anything else, including a read error or
// EOF, is a no: the safe default when treex is about to delete.
func (p *Printer) Prompt(in io.Reader, question string) bool {
	_, _ = fmt.Fprintf(p.w, "\n%s%s%s [y/N] ", p.bold, question, p.reset)

	answer, err := readLine(in)

	if err != nil {
		_, _ = io.WriteString(p.w, "\n")

		return false
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

// readLine reads a single line without buffering beyond it. A bufio.Reader
// would over-read stdin, which matters when the caller hands the same stream
// to something else after the prompt.
func readLine(in io.Reader) (string, error) {
	var (
		builder strings.Builder
		buf     [1]byte
	)

	for {
		n, err := in.Read(buf[:])

		if n > 0 {
			if buf[0] == '\n' {
				return builder.String(), nil
			}

			builder.WriteByte(buf[0])
		}

		if err != nil {
			if builder.Len() > 0 {
				return builder.String(), nil
			}

			return "", err
		}
	}
}
