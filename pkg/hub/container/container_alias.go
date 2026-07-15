package container

import "fmt"

// Alias creates an alias that resolves to the given abstract.
//
// It returns ErrSelfAlias when the alias and the abstract are the same name,
// and ErrAliasCycle when registering the alias would close a loop — for
// example Alias("a", "b") followed by Alias("b", "a").
//
// Rejecting cycles here rather than defending against them in GetAlias is what
// keeps resolution cheap: the table stays acyclic by construction, so walking a
// chain needs no visited-set bookkeeping. GetAlias sits on the hot path of
// every Make, and a cyclic alias is a wiring bug that should surface at
// registration anyway.
func (c *App) Alias(abstract, name string) error {
	if abstract == name {
		return fmt.Errorf("%w: %q", ErrSelfAlias, name)
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	// The table is acyclic, so this walk terminates. Adding name -> abstract
	// closes a loop exactly when abstract's chain already ends at name.
	if c.aliases.Resolve(abstract) == name {
		return fmt.Errorf("%w: %q -> %q", ErrAliasCycle, name, abstract)
	}

	c.aliases.Add(abstract, name)

	return nil
}

// GetAlias resolves an alias chain to the actual abstract name.
func (c *App) GetAlias(abstract string) string {
	c.mu.RLock()

	defer c.mu.RUnlock()

	return c.aliases.Resolve(abstract)
}

// IsAlias reports whether the given name is a registered alias.
func (c *App) IsAlias(name string) bool {
	c.mu.RLock()

	defer c.mu.RUnlock()

	return c.aliases.Has(name)
}
