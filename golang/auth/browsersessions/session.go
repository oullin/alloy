package browsersessions

import "time"

// Session describes an authenticated browser session.
type Session struct {
	ID           string
	UserID       string
	IPAddress    string
	UserAgent    string
	LastActiveAt time.Time
	Current      bool
}
