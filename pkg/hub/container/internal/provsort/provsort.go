// Package provsort orders service providers by their declared dependencies.
//
// It is pure: it holds no state, takes no locks, and touches nothing in the
// container beyond the provider contracts. Application.RegisterMany is its
// only caller.
package provsort

import (
	"errors"

	"github.com/oullin/alloy/pkg/hub/container/contracts/provider"
)

// ErrCycle is returned when the providers' DependsOn declarations form a cycle,
// leaving no valid registration order.
var ErrCycle = errors.New("dependency cycle among service providers")

// Sort returns the providers in dependency order. Edges come from
// provider.DependsOn declarations: if A declares DependsOn("foo") and some
// other provider B declares Provides("foo"), then B is sorted before A.
// Sort is stable for entries with no edges. Cycles return ErrCycle.
//
// A dependency on a key that no provider declares is ignored rather than
// rejected: you are simply registered after every provider that does declare
// DependsOn.
func Sort(in []provider.ServiceProvider) ([]provider.ServiceProvider, error) {
	byKey := make(map[string]int)

	for i, p := range in {
		if provides, ok := p.(provider.Provides); ok {
			for _, k := range provides.Provides() {
				byKey[k] = i
			}
		}
	}

	n := len(in)
	indeg := make([]int, n)
	out := make([][]int, n)

	for i, p := range in {
		dep, ok := p.(provider.DependsOn)

		if !ok {
			continue
		}

		for _, key := range dep.DependsOn() {
			u, found := byKey[key]

			if !found {
				continue
			}

			if u == i {
				continue
			}

			out[u] = append(out[u], i)
			indeg[i]++
		}
	}

	queue := make([]int, 0, n)

	for i := 0; i < n; i++ {
		if indeg[i] == 0 {
			queue = append(queue, i)
		}
	}

	result := make([]provider.ServiceProvider, 0, n)

	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		result = append(result, in[i])

		for _, j := range out[i] {
			indeg[j]--

			if indeg[j] == 0 {
				queue = append(queue, j)
			}
		}
	}

	if len(result) != n {
		return nil, ErrCycle
	}

	return result, nil
}
