// Package redistest provides an in-memory fake of the redis driver's Client
// contract.
//
// It lives here rather than in redis's own _test.go files because the driver
// wrappers (background, failover) test delegation against a real, working
// driver, and a fake defined in redis_test would not be importable from
// drivers_test. Mirrors the auth/internal/authtest pattern.
package redistest

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Client struct {
	mu        sync.Mutex
	lists     map[string][]string
	sorted    map[string][]sortedEntry
	EvalCalls []EvalCall
	PushErr   error
	EvalErr   error
}

type sortedEntry struct {
	score  float64
	member string
}

type EvalCall struct {
	Script string
	Keys   []string
	Args   []any
}

func NewClient() *Client {
	return &Client{
		lists:  make(map[string][]string),
		sorted: make(map[string][]sortedEntry),
	}
}

func (c *Client) LPush(_ context.Context, key string, values ...any) error {
	c.mu.Lock()

	defer c.mu.Unlock()

	if c.PushErr != nil {
		return c.PushErr
	}

	for _, v := range values {
		c.lists[key] = append([]string{fmt.Sprint(v)}, c.lists[key]...)
	}

	return nil
}

func (c *Client) RPop(_ context.Context, key string) (string, error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	list := c.lists[key]

	if len(list) == 0 {
		return "", errors.New("empty")
	}

	val := list[len(list)-1]
	c.lists[key] = list[:len(list)-1]

	return val, nil
}

func (c *Client) ZAdd(_ context.Context, key string, score float64, member string) error {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.sorted[key] = append(c.sorted[key], sortedEntry{score: score, member: member})

	return nil
}

func (c *Client) ZRangeByScore(_ context.Context, key string, min, max float64) ([]string, error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	var result []string

	for _, e := range c.sorted[key] {
		if e.score >= min && e.score <= max {
			result = append(result, e.member)
		}
	}

	return result, nil
}

func (c *Client) ZRem(_ context.Context, key string, members ...any) error {
	c.mu.Lock()

	defer c.mu.Unlock()

	memberSet := make(map[string]bool)

	for _, m := range members {
		memberSet[fmt.Sprint(m)] = true
	}

	var remaining []sortedEntry

	for _, e := range c.sorted[key] {
		if !memberSet[e.member] {
			remaining = append(remaining, e)
		}
	}

	c.sorted[key] = remaining

	return nil
}

func (c *Client) Eval(_ context.Context, script string, keys []string, args ...any) (any, error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.EvalCalls = append(c.EvalCalls, EvalCall{Script: script, Keys: append([]string(nil), keys...), Args: append([]any(nil), args...)})

	if c.EvalErr != nil {
		return nil, c.EvalErr
	}

	if len(keys) < 2 || len(args) == 0 {
		return int64(0), nil
	}

	max, ok := evalScore(args[0])

	if !ok {
		return int64(0), nil
	}

	var due []string

	var remaining []sortedEntry

	for _, entry := range c.sorted[keys[0]] {
		if entry.score <= max {
			due = append(due, entry.member)

			continue
		}

		remaining = append(remaining, entry)
	}

	c.sorted[keys[0]] = remaining

	for _, member := range due {
		c.lists[keys[1]] = append([]string{member}, c.lists[keys[1]]...)
	}

	return int64(len(due)), nil
}

func (c *Client) LLen(_ context.Context, key string) (int64, error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	return int64(len(c.lists[key])), nil
}

func (c *Client) ZCard(_ context.Context, key string) (int64, error) {
	c.mu.Lock()

	defer c.mu.Unlock()

	return int64(len(c.sorted[key])), nil
}

// evalScore normalises the Lua cutoff argument, which the driver passes as
// an int64 but a caller could reasonably supply as int or float64.
func evalScore(value any) (float64, bool) {
	switch v := value.(type) {
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// List returns a copy of the list stored at key. It backs test fixtures that
// layer the optional ListRanger capability over this fake.
func (c *Client) List(key string) []string {
	c.mu.Lock()

	defer c.mu.Unlock()

	return append([]string(nil), c.lists[key]...)
}

// SortedMembers returns the members of the sorted set at key, in insertion
// order. It backs test fixtures that layer the optional SortedSetRanger
// capability over this fake.
func (c *Client) SortedMembers(key string) []string {
	c.mu.Lock()

	defer c.mu.Unlock()

	out := make([]string, 0, len(c.sorted[key]))

	for _, e := range c.sorted[key] {
		out = append(out, e.member)
	}

	return out
}
