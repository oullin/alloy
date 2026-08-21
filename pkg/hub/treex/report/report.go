package report

import "time"

// Mode is what produced a report, which decides how it reads: a scan describes
// what is there, a dry run describes what would happen, an apply describes what
// did.
type Mode string

// Risk mirrors the policy verdict without importing it, keeping this package
// free of internal dependencies.
type Risk string

// Report is a whole run.
type Report struct {
	Mode      Mode       `json:"mode"`
	Version   string     `json:"version"`
	Config    string     `json:"config,omitempty"`
	Providers []Provider `json:"providers"`
	Totals    Totals     `json:"totals"`
	Repairs   []Repair   `json:"repairs,omitempty"`
	Warnings  []string   `json:"warnings,omitempty"`
	Elapsed   string     `json:"elapsed"`
}

// Provider groups a report by the agent the debris belongs to.
type Provider struct {
	Name       string     `json:"name"`
	Root       string     `json:"root"`
	Bytes      Bytes      `json:"bytes"`
	Categories []Category `json:"categories"`
	Entries    []Entry    `json:"entries"`
}

// Category is one class of reclaimable thing within a provider.
type Category struct {
	Name        string `json:"name"`
	Bytes       Bytes  `json:"bytes"`
	Count       int    `json:"count"`
	Reclaimable Bytes  `json:"reclaimable"`
	Blocked     Bytes  `json:"blocked"`
}

// Entry is one candidate.
type Entry struct {
	Path     string    `json:"path"`
	Label    string    `json:"label"`
	Category string    `json:"category"`
	Bytes    Bytes     `json:"bytes"`
	Risk     Risk      `json:"risk"`
	Reasons  []string  `json:"reasons,omitempty"`
	Kind     string    `json:"kind,omitempty"`
	Branch   string    `json:"branch,omitempty"`
	Repo     string    `json:"repo,omitempty"`
	Newest   time.Time `json:"newest,omitempty"`

	// Removed is set on an apply, and false on an entry that was selected but
	// failed. Error carries why.
	Removed bool   `json:"removed,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Repair is a registry entry pointing at a working tree that is already gone —
// damage from an earlier manual delete, which doctor fixes.
type Repair struct {
	Repo    string   `json:"repo"`
	Entries []string `json:"entries"`
	Fixed   bool     `json:"fixed,omitempty"`
}

// Totals is the bottom line.
type Totals struct {
	Scanned     Bytes `json:"scanned"`
	Reclaimable Bytes `json:"reclaimable"`
	Blocked     Bytes `json:"blocked"`
	Removed     Bytes `json:"removed,omitempty"`
	Files       int64 `json:"files"`
	Entries     int   `json:"entries"`
	BlockedRows int   `json:"blocked_entries"`
}

const (
	// ModeScan is a read-only measurement.
	ModeScan Mode = "scan"

	// ModeDry is a clean that was not given --apply.
	ModeDry Mode = "dry-run"

	// ModeApply is a clean that acted.
	ModeApply Mode = "apply"
)

const (
	// RiskSafe means nothing would be lost.
	RiskSafe Risk = "safe"

	// RiskBlocked means treex refused.
	RiskBlocked Risk = "blocked"
)

// ExitCode is the process status a report implies. An apply that left entries
// blocked is not a failure, but it is not a clean success either: the distinct
// code lets a scripted caller notice without parsing output.
func (r Report) ExitCode() int {
	if r.Mode == ModeApply && r.Totals.BlockedRows > 0 {
		return 3
	}

	return 0
}
