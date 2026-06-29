package multisteps_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"alloy.dev/backend/workflow/multisteps"
)

// Wait briefly so both async siblings overlap, confirming parallelism.

// Coordinate the two async jobs so the test exercises the fail-fast
// path deterministically: "slow" signals once it is inside its
// select (subscribed to ctx.Done), then "fast" returns its error.
// Without this, "fast" can finish so quickly that the engine cancels
// the group context before "slow" is ever invoked, and the test
// fails for a timing reason unrelated to the fail-fast semantics.

// Defensive: don't deadlock the test if "slow" never started.

type syncDriver struct{}

type goroutineDriver struct {
	maxConcurrency int
}

func TestRun_SignupDAG_SyncThenAsyncFanOut(t *testing.T) {
	var (
		createInvocations int32
		emailInvocations  int32
		notifyInvocations int32
		notifyStarted     = make(chan struct{}, 1)
		emailStarted      = make(chan struct{}, 1)
	)

	create := func(in multisteps.JobInput) (any, error) {
		atomic.AddInt32(&createInvocations, 1)

		name, _ := in.Resolved["name"].(string)

		return map[string]any{"id": "u-" + name, "name": name}, nil
	}

	email := func(in multisteps.JobInput) (any, error) {
		atomic.AddInt32(&emailInvocations, 1)
		emailStarted <- struct{}{}

		select {
		case <-notifyStarted:
		case <-time.After(time.Second):
			return nil, errors.New("notify never started")
		}

		uid, _ := in.Resolved["userId"].(string)

		return map[string]any{"sent_to": uid}, nil
	}

	notify := func(in multisteps.JobInput) (any, error) {
		atomic.AddInt32(&notifyInvocations, 1)
		notifyStarted <- struct{}{}

		select {
		case <-emailStarted:
		case <-time.After(time.Second):
			return nil, errors.New("email never started")
		}

		uid, _ := in.Resolved["userId"].(string)

		return map[string]any{"notified": uid}, nil
	}

	wf := multisteps.Machine("signup",
		multisteps.Sync("create", create, multisteps.Args(multisteps.A{
			"name": multisteps.Variable("name"),
		})),
		multisteps.Async("email", email, multisteps.Args(multisteps.A{
			"userId": multisteps.Response("create", "id"),
		})),
		multisteps.Async("notify", notify, multisteps.Args(multisteps.A{
			"userId": multisteps.Response("create", "id"),
		})),
	)

	eng := multisteps.NewEngine(multisteps.WithDriver(newGoroutineDriver(0)))

	res, err := eng.Run(context.Background(), wf, map[string]any{"name": "Jane"})

	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if got := atomic.LoadInt32(&createInvocations); got != 1 {
		t.Fatalf("create invoked %d times, want 1", got)
	}

	if got := atomic.LoadInt32(&emailInvocations); got != 1 {
		t.Fatalf("email invoked %d times, want 1", got)
	}

	if got := atomic.LoadInt32(&notifyInvocations); got != 1 {
		t.Fatalf("notify invoked %d times, want 1", got)
	}

	createOut, _ := res.Responses["create"].(map[string]any)

	if createOut["id"] != "u-Jane" {
		t.Fatalf("create response = %v", res.Responses["create"])
	}

	emailOut, _ := res.Responses["email"].(map[string]any)

	if emailOut["sent_to"] != "u-Jane" {
		t.Fatalf("email response = %v", res.Responses["email"])
	}
}

func TestRun_ResponseFieldOnStruct(t *testing.T) {
	type Account struct {
		ID   string
		Name string
	}

	wf := multisteps.Machine("struct-resp",
		multisteps.Sync("create", func(in multisteps.JobInput) (any, error) {
			return Account{ID: "abc", Name: "Jane"}, nil
		}),
		multisteps.Sync("consume", func(in multisteps.JobInput) (any, error) {
			return in.Resolved["id"], nil
		}, multisteps.Args(multisteps.A{
			"id": multisteps.Response("create", "ID"),
		})),
	)

	res, err := multisteps.Run(context.Background(), wf, nil)

	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if res.Responses["consume"] != "abc" {
		t.Fatalf("consume response = %v, want abc", res.Responses["consume"])
	}
}

func TestRun_RetryUntilSuccess(t *testing.T) {
	var attempts int32

	wf := multisteps.Machine("retry",
		multisteps.Sync("flaky", func(in multisteps.JobInput) (any, error) {
			n := atomic.AddInt32(&attempts, 1)

			if n < 3 {
				return nil, errors.New("transient")
			}

			return "ok", nil
		}, multisteps.WithRetry(3, time.Millisecond, time.Second)),
	)

	res, err := multisteps.Run(context.Background(), wf, nil)

	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Fatalf("attempts = %d, want 3", got)
	}

	if res.Responses["flaky"] != "ok" {
		t.Fatalf("flaky response = %v", res.Responses["flaky"])
	}
}

func TestRun_RetryExhaustedWrapsInWorkflowError(t *testing.T) {
	wf := multisteps.Machine("fail",
		multisteps.Sync("broken", func(in multisteps.JobInput) (any, error) {
			return nil, errors.New("permanent")
		}, multisteps.WithRetry(2, time.Millisecond, time.Second)),
	)

	_, err := multisteps.Run(context.Background(), wf, nil)

	if err == nil {
		t.Fatal("expected error")
	}

	var we *multisteps.WorkflowError

	if !errors.As(err, &we) {
		t.Fatalf("expected *WorkflowError, got %T: %v", err, err)
	}

	if we.Job != "broken" || we.Attempts != 2 {
		t.Fatalf("WorkflowError = %#v", we)
	}
}

func TestRun_RunIfSkipsAndDownstreamProceeds(t *testing.T) {
	var notifyRan bool

	wf := multisteps.Machine("conditional",
		multisteps.Sync("create", func(in multisteps.JobInput) (any, error) {
			return map[string]any{"id": "u1"}, nil
		}),
		multisteps.Sync("notify", func(in multisteps.JobInput) (any, error) {
			notifyRan = true

			return "notified", nil
		}, multisteps.Args(multisteps.A{
			"id": multisteps.Response("create", "id"),
		}), multisteps.WithRunIf(func(in multisteps.JobInput) bool {
			return in.Vars["env"] == "prod"
		})),
		multisteps.Sync("after", func(in multisteps.JobInput) (any, error) {
			return "done", nil
		}, multisteps.DependsOn("notify")),
	)

	res, err := multisteps.Run(context.Background(), wf, map[string]any{"env": "staging"})

	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if notifyRan {
		t.Fatal("notify should have been skipped")
	}

	if len(res.Skipped) != 1 || res.Skipped[0] != "notify" {
		t.Fatalf("Skipped = %v, want [notify]", res.Skipped)
	}

	if res.Responses["after"] != "done" {
		t.Fatalf("after didn't run after a skipped predecessor: %v", res.Responses)
	}
}

func TestCompile_DetectsCycle(t *testing.T) {
	wf := multisteps.Machine("cyclic",
		multisteps.Sync("a", func(multisteps.JobInput) (any, error) { return nil, nil },
			multisteps.DependsOn("b")),
		multisteps.Sync("b", func(multisteps.JobInput) (any, error) { return nil, nil },
			multisteps.DependsOn("a")),
	)

	err := wf.Compile()

	if err == nil {
		t.Fatal("expected cycle to be detected")
	}

	var ce *multisteps.CompileError

	if !errors.As(err, &ce) {
		t.Fatalf("expected *CompileError, got %T", err)
	}
}

func TestCompile_RejectsDanglingResponseRef(t *testing.T) {
	wf := multisteps.Machine("dangling",
		multisteps.Sync("a", func(multisteps.JobInput) (any, error) { return nil, nil },
			multisteps.Args(multisteps.A{"x": multisteps.Response("ghost", "id")})),
	)

	if err := wf.Compile(); err == nil {
		t.Fatal("expected dangling response to be rejected")
	}
}

func TestRun_AsyncFailFastCancelsSiblings(t *testing.T) {
	var siblingCancelled atomic.Bool

	slowEntered := make(chan struct{})

	wf := multisteps.Machine("failfast",
		multisteps.Async("fast", func(in multisteps.JobInput) (any, error) {
			select {
			case <-slowEntered:
			case <-time.After(2 * time.Second):

			}

			return nil, errors.New("boom")
		}),
		multisteps.Async("slow", func(in multisteps.JobInput) (any, error) {
			close(slowEntered)

			select {
			case <-in.Ctx.Done():
				siblingCancelled.Store(true)

				return nil, in.Ctx.Err()
			case <-time.After(2 * time.Second):
				return "completed", nil
			}
		}),
	)

	eng := multisteps.NewEngine(multisteps.WithDriver(newGoroutineDriver(0)))

	_, err := eng.Run(context.Background(), wf, nil)

	if err == nil {
		t.Fatal("expected error")
	}

	if !siblingCancelled.Load() {
		t.Fatal("sibling was not cancelled on fail-fast")
	}
}

func TestRun_ContinueOnErrorLetsSiblingsComplete(t *testing.T) {
	var completed atomic.Bool

	wf := multisteps.Machine("lenient",
		multisteps.Async("fast", func(in multisteps.JobInput) (any, error) {
			return nil, errors.New("boom")
		}),
		multisteps.Async("slow", func(in multisteps.JobInput) (any, error) {
			time.Sleep(20 * time.Millisecond)
			completed.Store(true)

			return "ok", nil
		}),
	)

	eng := multisteps.NewEngine(
		multisteps.WithDriver(newGoroutineDriver(0)),
		multisteps.WithContinueOnError(),
	)

	_, err := eng.Run(context.Background(), wf, nil)

	if err == nil {
		t.Fatal("expected first error to surface")
	}

	if !completed.Load() {
		t.Fatal("sibling did not complete in lenient mode")
	}
}

func TestRun_SyncDriverDeterministicOrdering(t *testing.T) {
	var order []string

	wf := multisteps.Machine("ordered",
		multisteps.Async("a", func(multisteps.JobInput) (any, error) {
			order = append(order, "a")

			return nil, nil
		}),
		multisteps.Async("b", func(multisteps.JobInput) (any, error) {
			order = append(order, "b")

			return nil, nil
		}),
	)

	eng := multisteps.NewEngine(multisteps.WithDriver(newSyncDriver()))

	if _, err := eng.Run(context.Background(), wf, nil); err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("expected deterministic order [a b], got %v", order)
	}
}

func newSyncDriver() multisteps.Driver {
	return &syncDriver{}
}

func (d *syncDriver) Run(ctx context.Context, tasks []multisteps.Task) ([]any, error) {
	results := make([]any, len(tasks))

	for i, task := range tasks {
		if err := ctx.Err(); err != nil {
			return results, err
		}

		val, err := task()

		if err != nil {
			return results, err
		}

		results[i] = val
	}

	return results, nil
}

func newGoroutineDriver(maxConcurrency int) multisteps.Driver {
	return &goroutineDriver{maxConcurrency: maxConcurrency}
}

func (d *goroutineDriver) Run(ctx context.Context, tasks []multisteps.Task) ([]any, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	results := make([]any, len(tasks))
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
		sem      chan struct{}
		mu       sync.Mutex
	)

	if d.maxConcurrency > 0 {
		sem = make(chan struct{}, d.maxConcurrency)
	}

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, fn multisteps.Task) {
			defer wg.Done()

			if sem != nil {
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-ctx.Done():
					once.Do(func() { firstErr = ctx.Err() })

					return
				}
			}

			if ctx.Err() != nil {
				return
			}

			val, err := fn()

			if err != nil {
				once.Do(func() {
					firstErr = err
					cancel()
				})

				return
			}

			mu.Lock()
			results[idx] = val
			mu.Unlock()
		}(i, task)
	}

	wg.Wait()

	if firstErr != nil {
		return results, firstErr
	}

	return results, nil
}
