package container

import (
	"fmt"
	"slices"
)

// Alias creates an alias that resolves to the given abstract. It returns
// ErrSelfAlias when the alias and the abstract are the same name.
func (c *App) Alias(abstract, alias string) error {
	if abstract == alias {
		return fmt.Errorf("%w: %q", ErrSelfAlias, alias)
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	c.aliases[alias] = abstract
	c.abstractAliases[abstract] = append(c.abstractAliases[abstract], alias)

	return nil
}

// GetAlias resolves an alias chain to the actual abstract name.
func (c *App) GetAlias(abstract string) string {
	c.mu.RLock()

	defer c.mu.RUnlock()

	return c.getAlias(abstract)
}

// IsAlias reports whether the given name is a registered alias.
func (c *App) IsAlias(name string) bool {
	c.mu.RLock()

	defer c.mu.RUnlock()

	_, ok := c.aliases[name]

	return ok
}

// getAlias resolves the full alias chain without locking. Caller must hold the
// lock.
func (c *App) getAlias(abstract string) string {
	for {
		target, ok := c.aliases[abstract]

		if !ok {
			return abstract
		}

		abstract = target
	}
}

// removeAlias removes the abstract from all alias mappings. Caller must hold
// the write lock.
func (c *App) removeAlias(abstract string) {
	for abs, aliases := range c.abstractAliases {
		for i, alias := range aliases {
			if alias == abstract {
				c.abstractAliases[abs] = slices.Delete(aliases, i, i+1)

				break
			}
		}
	}

	delete(c.aliases, abstract)
}
