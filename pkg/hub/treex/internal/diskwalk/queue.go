package diskwalk

import "sync"

// task is one directory waiting to be read, tagged with the root it belongs to
// and the leaf it should be attributed to (-1 for the root itself).
type task struct {
	root  int
	dir   string
	leaf  int
	depth int
}

// queue is a LIFO stack of pending directories guarded by a condition
// variable.
//
// LIFO rather than FIFO is deliberate: depth-first ordering keeps the working
// set of directory entries small and warm in the page cache, where breadth-first
// would fan out across the whole tree before finishing anything.
//
// The active counter is what makes termination detectable. A worker that finds
// the stack empty cannot conclude the walk is over, because another worker may
// be mid-ReadDir and about to push more work. Only when the stack is empty and
// no task is in flight is the walk genuinely finished.
type queue struct {
	mu     sync.Mutex
	cond   *sync.Cond
	stack  []task
	active int
	closed bool
}

func newQueue() *queue {
	q := &queue{}
	q.cond = sync.NewCond(&q.mu)

	return q
}

// push adds directories to the stack and wakes one worker per item.
func (q *queue) push(tasks ...task) {
	if len(tasks) == 0 {
		return
	}

	q.mu.Lock()

	defer q.mu.Unlock()

	if q.closed {
		return
	}

	q.stack = append(q.stack, tasks...)

	for range tasks {
		q.cond.Signal()
	}
}

// seed primes the stack with the initial roots and marks them in flight, so a
// worker that starts before the first push cannot mistake an empty stack for a
// finished walk.
func (q *queue) seed(tasks []task) {
	q.mu.Lock()

	defer q.mu.Unlock()

	q.stack = append(q.stack, tasks...)
	q.cond.Broadcast()
}

// pop blocks until a directory is available, the walk is finished, or the queue
// is closed. The second return value is false only when there is no more work.
func (q *queue) pop() (task, bool) {
	q.mu.Lock()

	defer q.mu.Unlock()

	for len(q.stack) == 0 && q.active > 0 && !q.closed {
		q.cond.Wait()
	}

	if q.closed || len(q.stack) == 0 {
		return task{}, false
	}

	last := len(q.stack) - 1
	next := q.stack[last]

	q.stack = q.stack[:last]
	q.active++

	return next, true
}

// done marks a directory finished. When it is the last one in flight and
// nothing is queued, the walk is over and every waiting worker is released.
func (q *queue) done() {
	q.mu.Lock()

	defer q.mu.Unlock()

	q.active--

	if q.active == 0 && len(q.stack) == 0 {
		q.closed = true
		q.cond.Broadcast()
	}
}

// close abandons the walk and wakes every blocked worker. It is what turns a
// cancelled context into a prompt return.
func (q *queue) close() {
	q.mu.Lock()

	defer q.mu.Unlock()

	q.closed = true
	q.stack = nil
	q.cond.Broadcast()
}
