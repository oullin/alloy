package failed

import "time"

// Job is the decoded failed-job record returned by providers.
type Job struct {
	ID         string
	UUID       string
	Connection string
	Backend    string
	Payload    string
	Exception  string
	FailedAt   time.Time
}

// Provider persists permanently failed jobs.
type Provider interface {
	Log(connection, queue, payload string, exception error) (string, error)
	IDs(queueFilter string) ([]string, error)
	All() ([]Job, error)
	Find(id string) (*Job, error)
	Forget(id string) (bool, error)
	Flush(hours int) error
}

// Countable is implemented by providers that support efficient counting.
type Countable interface {
	Count(connection, queueFilter string) (int64, error)
}

// Prunable is implemented by providers that can prune old rows.
type Prunable interface {
	Prune(before time.Time) (int64, error)
}
