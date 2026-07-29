package bus_test

import (
	"context"
	"errors"
	"testing"

	"hara.sh/alloy/bus"
)

type chainJob struct {
	bus.Queueable
	Value string
}

func TestPendingChainDispatch(t *testing.T) {
	d := bus.NewDispatcher(nil, nil)

	j1 := &chainJob{Value: "first"}
	j2 := &chainJob{Value: "second"}
	j3 := &chainJob{Value: "third"}

	d.Map(j1, func(_ context.Context, cmd any) (any, error) {
		j := cmd.(*chainJob)

		return j.Value, nil
	})

	chain := bus.NewPendingChain(d, []any{j1, j2, j3})
	result, err := chain.Dispatch(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if result != "first" {
		t.Errorf("expected 'first', got %v", result)
	}

	if len(j1.ChainJobs) != 2 {
		t.Errorf("expected 2 chain jobs on first job, got %d", len(j1.ChainJobs))
	}
}

func TestPendingChainOnConnectionOnQueue(t *testing.T) {
	d := bus.NewDispatcher(nil, nil)

	j1 := &chainJob{Value: "first"}
	d.Map(j1, func(_ context.Context, cmd any) (any, error) {
		return nil, nil
	})

	chain := bus.NewPendingChain(d, []any{j1}).
		OnConnection("redis").
		OnQueue("high")

	_, err := chain.Dispatch(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if j1.Connection != "redis" {
		t.Errorf("expected connection 'redis', got %q", j1.Connection)
	}

	if j1.Backend != "high" {
		t.Errorf("expected queue 'high', got %q", j1.Backend)
	}
}

func TestPendingChainEmptyError(t *testing.T) {
	d := bus.NewDispatcher(nil, nil)
	chain := bus.NewPendingChain(d, []any{})

	_, err := chain.Dispatch(context.Background())

	if !errors.Is(err, bus.ErrEmptyChain) {
		t.Errorf("expected ErrEmptyChain, got %v", err)
	}
}

func TestPendingChainDispatchErrorsWhenFirstJobCannotAcceptChain(t *testing.T) {
	d := bus.NewDispatcher(nil, nil)
	first := struct{ Value string }{Value: "first"}

	d.Map(first, func(_ context.Context, cmd any) (any, error) {
		return cmd, nil
	})

	chain := bus.NewPendingChain(d, []any{first, &chainJob{Value: "second"}})

	if _, err := chain.Dispatch(context.Background()); err == nil {
		t.Fatal("expected error when first job cannot accept remaining chain jobs")
	}
}

func TestPendingChainCatchCallbacks(t *testing.T) {
	d := bus.NewDispatcher(nil, nil)

	j1 := &chainJob{Value: "first"}
	d.Map(j1, func(_ context.Context, cmd any) (any, error) {
		return nil, nil
	})

	catchCalled := false
	chain := bus.NewPendingChain(d, []any{j1}).
		Catch(func(_ context.Context, _ error) { catchCalled = true })

	_, err := chain.Dispatch(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if len(j1.ChainCatchCallbacks) != 1 {
		t.Errorf("expected 1 chain catch callback on first job, got %d", len(j1.ChainCatchCallbacks))
	}

	// Invoke to verify it works.
	j1.InvokeChainCatchCallbacks(context.Background(), errTestFailure)

	if !catchCalled {
		t.Error("expected catch callback to be called")
	}
}

func TestPendingChainDispatchAfterResponse(t *testing.T) {
	d := bus.NewDispatcher(nil, nil)

	j1 := &chainJob{Value: "first"}
	d.Map(j1, func(_ context.Context, cmd any) (any, error) {
		return nil, nil
	})

	chain := bus.NewPendingChain(d, []any{j1, &chainJob{Value: "second"}})

	err := chain.DispatchAfterResponse(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	if len(j1.ChainJobs) != 1 {
		t.Errorf("expected 1 chain job on first job, got %d", len(j1.ChainJobs))
	}
}

func TestPendingChainDispatchAfterResponseEmptyError(t *testing.T) {
	d := bus.NewDispatcher(nil, nil)
	chain := bus.NewPendingChain(d, []any{})

	err := chain.DispatchAfterResponse(context.Background())

	if !errors.Is(err, bus.ErrEmptyChain) {
		t.Errorf("expected ErrEmptyChain, got %v", err)
	}
}

func TestPendingChainDispatchAfterResponseErrorsWhenFirstJobCannotAcceptChain(t *testing.T) {
	d := bus.NewDispatcher(nil, nil)
	chain := bus.NewPendingChain(d, []any{struct{ Value string }{Value: "first"}, &chainJob{Value: "second"}})

	if err := chain.DispatchAfterResponse(context.Background()); err == nil {
		t.Fatal("expected error when first job cannot accept remaining chain jobs")
	}
}
