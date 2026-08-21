package diskwalk

import (
	"io/fs"
	"sync"
)

// FileKey identifies a file by device and inode, which is how two paths are
// recognised as the same bytes on disk.
type FileKey struct {
	Dev uint64
	Ino uint64
}

// linkSet remembers which multiply-linked files have already been counted.
//
// Only files with more than one link are ever inserted, so the set stays small
// even on a tree of a million files, and the overwhelmingly common
// single-linked case never takes a lock at all.
type linkSet struct {
	enabled bool
	shards  [linkShards]linkShard
}

type linkShard struct {
	mu   sync.Mutex
	seen map[FileKey]struct{}
}

// Zero reports whether the key carries no identity, which happens on platforms
// where the underlying stat is unavailable.
func (k FileKey) Zero() bool {
	return k.Dev == 0 && k.Ino == 0
}

// linkShards spreads the dedupe set across independent mutexes so the workers
// that do consult it rarely contend.
const linkShards = 64

func newLinkSet(enabled bool) *linkSet {
	return &linkSet{enabled: enabled}
}

// claim reports whether this key is being counted for the first time. A
// disabled set claims everything, which makes the caller branch-free.
func (l *linkSet) claim(key FileKey) bool {
	if !l.enabled || key.Zero() {
		return true
	}

	shard := &l.shards[key.Ino%linkShards]

	shard.mu.Lock()

	defer shard.mu.Unlock()

	if shard.seen == nil {
		shard.seen = make(map[FileKey]struct{})
	}

	if _, ok := shard.seen[key]; ok {
		return false
	}

	shard.seen[key] = struct{}{}

	return true
}

// FileKeyOf returns a file's device/inode identity and its link count. It is
// exported so callers can record the identity of a path at scan time and check
// it again immediately before deleting, catching a path that was replaced in
// between.
func FileKeyOf(info fs.FileInfo) (FileKey, uint64, bool) {
	return fileKeyOf(info)
}
