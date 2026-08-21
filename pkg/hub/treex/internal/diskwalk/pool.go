package diskwalk

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Pool owns the workers for every walk in the process. Roots are submitted to
// it as jobs rather than each getting its own pool, because a scan routinely
// has hundreds of roots and only a handful of cores.
type Pool struct {
	jobs int
}

// walkRun is the state shared by the workers of a single Run.
type walkRun struct {
	walker Walker
	queue  *queue
	accum  []*accumulator
	seen   *linkSet

	progressFiles atomic.Int64
	progressBytes atomic.Int64
}

// maxJobs caps the worker count. Past a few dozen concurrent readdir calls the
// queue depth stops helping and starts costing context switches.
const maxJobs = 64

// NewPool builds a pool with the given worker count. Zero or negative derives
// one from the machine: four times NumCPU, because workers spend nearly all
// their time blocked on the filesystem rather than running.
func NewPool(jobs int) *Pool {
	if jobs <= 0 {
		jobs = runtime.NumCPU() * 4
	}

	return &Pool{jobs: min(jobs, maxJobs)}
}

// Jobs reports the resolved worker count.
func (p *Pool) Jobs() int {
	return p.jobs
}

// Run measures every root and returns one Result per root, in the order the
// roots were given.
//
// A cancelled context returns the partial measurements alongside
// ErrCancelled: a scan interrupted halfway still has something worth showing,
// and discarding it would make Ctrl-C feel like a failure rather than a stop.
func (p *Pool) Run(ctx context.Context, walker Walker, roots []string) ([]Result, error) {
	if len(roots) == 0 {
		return nil, nil
	}

	accumulators := make([]*accumulator, len(roots))
	seeds := make([]task, 0, len(roots))

	for index, root := range roots {
		acc := &accumulator{root: root}
		accumulators[index] = acc

		// A root is not always a directory. Orphaned sockets and pid files are
		// measured as roots in their own right, and handing one to ReadDir
		// would report it as an unreadable directory worth nothing.
		info, err := os.Lstat(root)

		if err == nil && !info.IsDir() {
			acc.addFile(-1, sizeOf(info, walker.Apparent), info.ModTime())

			continue
		}

		seeds = append(seeds, task{root: index, dir: root, leaf: -1, depth: 0})
	}

	if len(seeds) == 0 {
		return collect(accumulators), nil
	}

	q := newQueue()
	q.seed(seeds)

	// The watchdog is what makes cancellation prompt. Without it a worker
	// blocked in cond.Wait would sit there until the rest of the walk drained.
	watchdog, stopWatchdog := context.WithCancel(context.Background())

	defer stopWatchdog()

	go func() {
		select {
		case <-ctx.Done():
			q.close()
		case <-watchdog.Done():
		}
	}()

	run := &walkRun{
		walker: walker,
		queue:  q,
		accum:  accumulators,
		seen:   newLinkSet(walker.Dedupe),
	}

	var wg sync.WaitGroup

	for range p.jobs {
		wg.Add(1)

		go func() {
			defer wg.Done()

			run.work(ctx)
		}()
	}

	wg.Wait()

	results := collect(accumulators)

	if err := ctx.Err(); err != nil {
		return results, ErrCancelled
	}

	return results, nil
}

func collect(accumulators []*accumulator) []Result {
	results := make([]Result, len(accumulators))

	for index, acc := range accumulators {
		results[index] = acc.result()
	}

	return results
}

func (r *walkRun) work(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		next, ok := r.queue.pop()

		if !ok {
			return
		}

		r.visit(next)
		r.queue.done()
	}
}

func (r *walkRun) visit(current task) {
	acc := r.accum[current.root]

	entries, err := os.ReadDir(current.dir)

	if err != nil {
		acc.addError()
		r.report(current.dir, err)

		return
	}

	children := make([]task, 0, len(entries))

	for _, entry := range entries {
		path := filepath.Join(current.dir, entry.Name())

		if entry.Type()&os.ModeSymlink != 0 && !r.walker.FollowSymlinks {
			continue
		}

		if entry.IsDir() {
			child, keep := r.directory(current, acc, path, entry)

			if keep {
				children = append(children, child)
			}

			continue
		}

		r.file(acc, current.leaf, path, entry)
	}

	r.queue.push(children...)
}

func (r *walkRun) directory(current task, acc *accumulator, path string, entry os.DirEntry) (task, bool) {
	decision := DecisionDescend

	if r.walker.Pruner != nil {
		decision = r.walker.Pruner.Prune(path, entry, current.depth+1)
	}

	if decision == DecisionSkip {
		return task{}, false
	}

	if r.walker.MaxDepth > 0 && current.depth+1 > r.walker.MaxDepth {
		return task{}, false
	}

	if info, err := entry.Info(); err == nil {
		acc.addDir(info.ModTime())
	} else {
		acc.addDir(time.Time{})
	}

	leaf := current.leaf

	// Only mark when not already inside a marked subtree, so a node_modules
	// nested in a node_modules rolls up rather than being reported twice.
	if decision == DecisionMark && leaf < 0 {
		leaf = acc.mark(path, entry.Name())
	}

	return task{root: current.root, dir: path, leaf: leaf, depth: current.depth + 1}, true
}

func (r *walkRun) file(acc *accumulator, leaf int, path string, entry os.DirEntry) {
	info, err := entry.Info()

	if err != nil {
		acc.addError()
		r.report(path, err)

		return
	}

	key, links, ok := fileKeyOf(info)

	// Only multiply-linked files can be double counted, so the dedupe set is
	// consulted for those alone and the common path stays lock-free.
	if ok && links > 1 && !r.seen.claim(key) {
		return
	}

	size := sizeOf(info, r.walker.Apparent)

	acc.addFile(leaf, size, info.ModTime())
	r.tick(size)
}

func (r *walkRun) tick(size int64) {
	if r.walker.Progress == nil {
		return
	}

	files := r.progressFiles.Add(1)
	bytes := r.progressBytes.Add(size)

	// Reporting every file would cost more than the walk. One tick per few
	// thousand is plenty to keep a counter moving.
	if files%4096 == 0 {
		r.walker.Progress(files, bytes)
	}
}

func (r *walkRun) report(path string, err error) {
	if r.walker.Errors == nil {
		return
	}

	r.walker.Errors(path, err)
}
