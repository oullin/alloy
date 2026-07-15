package container

// When creates a ContextualBindingBuilder for the given concrete type(s).
func (c *App) When(concrete ...string) *ContextualBindingBuilder {
	return &ContextualBindingBuilder{
		container: c,
		concrete:  concrete,
	}
}

// AddContextualBinding registers a contextual binding. When the concrete type
// is being resolved and needs the given abstract, the implementation is used
// instead.
func (c *App) AddContextualBinding(concrete, abstract string, implementation any) {
	c.mu.Lock()

	defer c.mu.Unlock()

	if c.contextual[concrete] == nil {
		c.contextual[concrete] = make(map[string]any)
	}

	c.contextual[concrete][abstract] = implementation
}

// getContextualConcrete looks up a contextual binding for the given abstract
// based on the current build stack. Caller must hold the lock.
func (c *App) getContextualConcrete(abstract string) any {
	if c.resolution == nil || len(c.resolution.buildStack) == 0 {
		return nil
	}

	current := c.resolution.buildStack[len(c.resolution.buildStack)-1]

	if bindings, ok := c.contextual[current]; ok {
		if impl, ok := bindings[abstract]; ok {
			return impl
		}
	}

	return nil
}
