//go:build unix

package diskwalk

import (
	"io/fs"
	"syscall"
)

// sizeOf returns the bytes a file occupies. Allocated blocks are the default
// because they are what du reports and what actually frees when the file is
// removed: apparent size over-reports sparse files, and under-reports the
// per-file block floor that dominates a directory of many tiny files, which is
// exactly what node_modules is.
func sizeOf(info fs.FileInfo, apparent bool) int64 {
	if apparent {
		return info.Size()
	}

	stat, ok := info.Sys().(*syscall.Stat_t)

	if !ok {
		return info.Size()
	}

	return stat.Blocks * 512
}

// fileKeyOf returns the file's identity and link count. The link count is what
// tells the caller whether the dedupe set needs consulting at all.
func fileKeyOf(info fs.FileInfo) (FileKey, uint64, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)

	if !ok {
		return FileKey{}, 1, false
	}

	return FileKey{Dev: uint64(stat.Dev), Ino: stat.Ino}, uint64(stat.Nlink), true
}
