package container

// Bind registers a factory for the given abstract. When shared is true the
// factory is only called once and the result is cached.
func (c *App) Bind(abstract string, factory Factory, shared bool) {
	c.mu.Lock()

	c.dropStale(abstract)

	c.bindings[abstract] = Binding{factory: factory, shared: shared}

	wasResolved := c.resolved[abstract]

	c.mu.Unlock()

	if wasResolved {
		c.rebound(abstract)
	}
}

// BindIf registers a binding only if the abstract is not already bound.
func (c *App) BindIf(abstract string, factory Factory, shared bool) {
	if !c.Bound(abstract) {
		c.Bind(abstract, factory, shared)
	}
}

// Singleton registers a shared binding. The factory is called at most once.
func (c *App) Singleton(abstract string, factory Factory) {
	c.Bind(abstract, factory, true)
}

// SingletonIf registers a singleton only if the abstract is not already bound.
func (c *App) SingletonIf(abstract string, factory Factory) {
	if !c.Bound(abstract) {
		c.Singleton(abstract, factory)
	}
}

// Scoped registers a scoped binding. Scoped bindings behave like singletons
// but can be flushed independently via ForgetScopedInstances.
func (c *App) Scoped(abstract string, factory Factory) {
	c.mu.Lock()

	c.dropStale(abstract)

	c.bindings[abstract] = Binding{factory: factory, shared: true, scoped: true}

	wasResolved := c.resolved[abstract]

	c.mu.Unlock()

	if wasResolved {
		c.rebound(abstract)
	}
}

// ScopedIf registers a scoped binding only if the abstract is not already bound.
func (c *App) ScopedIf(abstract string, factory Factory) {
	if !c.Bound(abstract) {
		c.Scoped(abstract, factory)
	}
}

// Instance registers a pre-existing value in the container. Returns the
// instance for convenience.
func (c *App) Instance(abstract string, instance any) any {
	c.mu.Lock()

	c.aliases.Remove(abstract)

	wasBound := c.isBound(abstract)
	c.instances[abstract] = instance

	c.mu.Unlock()

	if wasBound {
		c.rebound(abstract)
	}

	return instance
}

// dropStale removes the cached instance and alias entries for the given
// abstract. Caller must hold the write lock.
func (c *App) dropStale(abstract string) {
	delete(c.instances, abstract)
	c.aliases.Drop(abstract)
}
