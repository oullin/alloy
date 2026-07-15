package provsort_test

import (
	"errors"
	"testing"

	"github.com/oullin/alloy/pkg/hub/container/internal/provsort"

	"github.com/oullin/alloy/pkg/hub/container/contracts/provider"
)

// stub declares whatever combination of Provides/DependsOn a case needs.
// A nil provides or depends slice still satisfies the interface, so every
// stub is both a Provides and a DependsOn; cases that need a provider
// implementing *neither* use bare instead.
type stub struct {
	name     string
	provides []string
	depends  []string
}

// bare implements only ServiceProvider, exercising the type-assertion misses.
type bare struct{ name string }

func (s *stub) Register()           {}
func (s *stub) Provides() []string  { return s.provides }
func (s *stub) DependsOn() []string { return s.depends }

func (b *bare) Register() {}

func names(t *testing.T, in []provider.ServiceProvider) []string {
	t.Helper()

	out := make([]string, 0, len(in))

	for _, p := range in {
		switch v := p.(type) {
		case *stub:
			out = append(out, v.name)
		case *bare:
			out = append(out, v.name)
		default:
			t.Fatalf("unexpected provider type %T", p)
		}
	}

	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestSortDependencyComesFirst(t *testing.T) {
	t.Parallel()

	a := &stub{name: "a", depends: []string{"foo"}}
	b := &stub{name: "b", provides: []string{"foo"}}

	got, err := provsort.Sort([]provider.ServiceProvider{a, b})

	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if want := []string{"b", "a"}; !equal(names(t, got), want) {
		t.Fatalf("got %v, want %v", names(t, got), want)
	}
}

func TestSortIsStableWithoutEdges(t *testing.T) {
	t.Parallel()

	in := []provider.ServiceProvider{&bare{name: "a"}, &bare{name: "b"}, &bare{name: "c"}}

	got, err := provsort.Sort(in)

	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if want := []string{"a", "b", "c"}; !equal(names(t, got), want) {
		t.Fatalf("got %v, want %v", names(t, got), want)
	}
}

func TestSortIgnoresUnknownDependency(t *testing.T) {
	t.Parallel()

	// Depending on a key nobody provides is permitted, not an error.
	a := &stub{name: "a", depends: []string{"nobody-binds-this"}}

	got, err := provsort.Sort([]provider.ServiceProvider{a})

	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if want := []string{"a"}; !equal(names(t, got), want) {
		t.Fatalf("got %v, want %v", names(t, got), want)
	}
}

func TestSortIgnoresSelfDependency(t *testing.T) {
	t.Parallel()

	// A provider that depends on a key it provides itself must not deadlock
	// the sort by giving itself an in-degree it can never discharge.
	a := &stub{name: "a", provides: []string{"foo"}, depends: []string{"foo"}}

	got, err := provsort.Sort([]provider.ServiceProvider{a})

	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if want := []string{"a"}; !equal(names(t, got), want) {
		t.Fatalf("got %v, want %v", names(t, got), want)
	}
}

func TestSortDetectsCycle(t *testing.T) {
	t.Parallel()

	a := &stub{name: "a", provides: []string{"a"}, depends: []string{"b"}}
	b := &stub{name: "b", provides: []string{"b"}, depends: []string{"a"}}

	got, err := provsort.Sort([]provider.ServiceProvider{a, b})

	if !errors.Is(err, provsort.ErrCycle) {
		t.Fatalf("expected ErrCycle, got %v", err)
	}

	if got != nil {
		t.Fatalf("expected nil result on cycle, got %v", names(t, got))
	}
}

func TestSortTransitiveChain(t *testing.T) {
	t.Parallel()

	// Registered in reverse order; the sort must fully invert them.
	c := &stub{name: "c", provides: []string{"c"}, depends: []string{"b"}}
	b := &stub{name: "b", provides: []string{"b"}, depends: []string{"a"}}
	a := &stub{name: "a", provides: []string{"a"}}

	got, err := provsort.Sort([]provider.ServiceProvider{c, b, a})

	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if want := []string{"a", "b", "c"}; !equal(names(t, got), want) {
		t.Fatalf("got %v, want %v", names(t, got), want)
	}
}

func TestSortEmpty(t *testing.T) {
	t.Parallel()

	got, err := provsort.Sort(nil)

	if err != nil {
		t.Fatalf("Sort: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", names(t, got))
	}
}
