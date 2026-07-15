package container

// ForgetInstance removes the cached instance for the given abstract.
func (c *App) ForgetInstance(abstract string) {
	c.mu.Lock()

	defer c.mu.Unlock()

	delete(c.instances, abstract)
}

// ForgetInstances removes all cached instances.
func (c *App) ForgetInstances() {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.instances = make(map[string]any)
}

// ForgetScopedInstances removes only the cached instances for scoped bindings.
func (c *App) ForgetScopedInstances() {
	c.mu.Lock()

	defer c.mu.Unlock()

	for abstract, b := range c.bindings {
		if b.scoped {
			delete(c.instances, abstract)
		}
	}
}

// GetBindings returns a copy of all registered bindings.
func (c *App) GetBindings() map[string]Binding {
	c.mu.RLock()

	defer c.mu.RUnlock()

	out := make(map[string]Binding, len(c.bindings))

	for k, v := range c.bindings {
		out[k] = v
	}

	return out
}
