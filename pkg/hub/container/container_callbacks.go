package container

// BeforeResolving registers a callback that fires before the given abstract is
// resolved.
func (c *App) BeforeResolving(abstract string, callback BeforeResolvingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.beforeCbs[abstract] = append(c.beforeCbs[abstract], callback)
}

// BeforeResolvingAny registers a global before-resolving callback.
func (c *App) BeforeResolvingAny(callback BeforeResolvingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.globalBeforeCbs = append(c.globalBeforeCbs, callback)
}

// Resolving registers a callback that fires when the given abstract is being
// resolved.
func (c *App) Resolving(abstract string, callback BindingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.resolvCbs[abstract] = append(c.resolvCbs[abstract], callback)
}

// ResolvingAny registers a global resolving callback.
func (c *App) ResolvingAny(callback BindingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.globalResolvCbs = append(c.globalResolvCbs, callback)
}

// AfterResolving registers a callback that fires after the given abstract is
// resolved.
func (c *App) AfterResolving(abstract string, callback BindingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.afterCbs[abstract] = append(c.afterCbs[abstract], callback)
}

// AfterResolvingAny registers a global after-resolving callback.
func (c *App) AfterResolvingAny(callback BindingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.globalAfterCbs = append(c.globalAfterCbs, callback)
}

// fireCallbacks invokes each callback with the given instance and container.
func fireCallbacks(callbacks []BindingCallback, instance any, c *App) {
	for _, cb := range callbacks {
		cb(instance, c)
	}
}

// fireBeforeCallbacks invokes each before-resolving callback.
func fireBeforeCallbacks(callbacks []BeforeResolvingCallback, abstract string, params map[string]any, c *App) {
	for _, cb := range callbacks {
		cb(abstract, params, c)
	}
}
