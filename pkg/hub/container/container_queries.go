package container

// Bound reports whether an abstract has a binding, instance, or alias.
func (c *App) Bound(abstract string) bool {
	c.mu.RLock()

	defer c.mu.RUnlock()

	return c.isBound(abstract)
}

// isBound is the lock-free version of Bound. Caller must hold the lock.
func (c *App) isBound(abstract string) bool {
	_, hasBind := c.bindings[abstract]
	_, hasInst := c.instances[abstract]
	_, hasAlias := c.aliases[abstract]

	return hasBind || hasInst || hasAlias
}

// Has is an alias for Bound (PSR-11 parity).
func (c *App) Has(abstract string) bool {
	return c.Bound(abstract)
}

// Resolved reports whether the given abstract has been resolved at least once.
func (c *App) Resolved(abstract string) bool {
	c.mu.RLock()

	defer c.mu.RUnlock()

	abs := c.getAlias(abstract)

	if _, ok := c.instances[abs]; ok {
		return true
	}

	return c.resolved[abs]
}

// IsShared reports whether the given abstract is a singleton or scoped binding.
func (c *App) IsShared(abstract string) bool {
	c.mu.RLock()

	defer c.mu.RUnlock()

	abs := c.getAlias(abstract)

	if _, ok := c.instances[abs]; ok {
		return true
	}

	if b, ok := c.bindings[abs]; ok {
		return b.shared
	}

	return false
}

// CurrentlyResolving returns the abstract at the top of the build stack, or an
// empty string if nothing is being resolved.
func (c *App) CurrentlyResolving() string {
	c.mu.RLock()

	defer c.mu.RUnlock()

	if c.resolution == nil || len(c.resolution.buildStack) == 0 {
		return ""
	}

	return c.resolution.buildStack[len(c.resolution.buildStack)-1]
}
