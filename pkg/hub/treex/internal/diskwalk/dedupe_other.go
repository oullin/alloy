//go:build !unix

package diskwalk

import "io/fs"

// sizeOf falls back to apparent size where allocated blocks are not exposed.
func sizeOf(info fs.FileInfo, _ bool) int64 {
	return info.Size()
}

// fileKeyOf reports no identity, which disables hard-link dedupe. Reporting a
// link count of 1 keeps the caller on the lock-free path.
func fileKeyOf(_ fs.FileInfo) (FileKey, uint64, bool) {
	return FileKey{}, 1, false
}
