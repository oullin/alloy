package container

import "fmt"

// Alias creates an alias that resolves to the given abstract. It returns
// ErrSelfAlias when the alias and the abstract are the same name.
func (c *App) Alias(abstract, name string) error {
	if abstract == name {
		return fmt.Errorf("%w: %q", ErrSelfAlias, name)
	}

	c.mu.Lock()

	defer c.mu.Unlock()

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
