package container

import (
	"errors"
	"fmt"
	"sync"
)

// call represents an in-flight or completed single-flight resolution.
type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

// singleflight manages concurrent duplicate resolutions so that the factory
// only executes once.
type singleflight struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do executes and returns the results of the given function, making sure that
// only one execution is in-flight for a given key at a time. If a duplicate
// comes in, the duplicate caller waits for the original to complete and receives
// the same results.
func (g *singleflight) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()

		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	completed := false

	defer func() {
		if completed {
			return
		}

		// fn never returned: either it panicked, or the goroutine is
		// unwinding via runtime.Goexit (e.g. t.FailNow inside a test
		// factory). Keying cleanup off a completion flag instead of
		// recover() alone guarantees the key is removed and waiters are
		// released in both cases; otherwise every future Do for this
		// key would block forever.
		r := recover()

		if r != nil {
			c.err = fmt.Errorf("factory panicked: %v", r)
		} else {
			c.err = errors.New("factory did not complete (runtime.Goexit)")
		}

		g.mu.Lock()
		delete(g.m, key)
		g.mu.Unlock()
		c.wg.Done()

		if r != nil {
			// Re-raise real panics so the leader's caller still observes
			// them. A Goexit must not be converted into a panic.
			panic(r)
		}
	}()

	c.val, c.err = fn()
	completed = true

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	c.wg.Done()

	return c.val, c.err
}
