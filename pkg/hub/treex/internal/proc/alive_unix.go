//go:build unix

package proc

import (
	"errors"
	"syscall"
)

// Alive reports whether a process with this id exists.
//
// Signal zero performs the permission and existence checks without delivering
// anything. ESRCH means the process is genuinely gone; EPERM means it exists
// but belongs to someone else, which still counts as alive — treating another
// user's running browser as dead would delete the socket out from under it.
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}

	err := syscall.Kill(pid, 0)

	if err == nil {
		return true
	}

	return !errors.Is(err, syscall.ESRCH)
}
