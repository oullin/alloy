//go:build !unix

package proc

// Alive reports every process as running where liveness cannot be checked.
// Erring towards "alive" means a pid-gated orphan rule simply matches nothing
// rather than deleting the files of a process that is still using them.
func Alive(pid int) bool {
	return pid > 0
}
