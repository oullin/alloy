package security

import "time"

// Timebox executes fn while ensuring it takes at least minDuration.
func Timebox(minDuration time.Duration, fn func()) {
	start := time.Now()
	fn()
	elapsed := time.Since(start)

	if elapsed < minDuration {
		time.Sleep(minDuration - elapsed)
	}
}
