// Package console renders treex's sectioned, ANSI-colored terminal output:
// section headers, aligned detail lines, warning and failure banners, and the
// aligned candidate tables that make up a scan report.
//
// Color detection is resolved once by the caller (see DetectColor) and handed
// to NewPrinter, so the printer itself never reads the environment and tests
// can construct a plain-text printer without touching process state.
package console
