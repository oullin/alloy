package container

// BeforeResolving registers a callback that fires before the given abstract is
// resolved.
func (c *App) BeforeResolving(abstract string, callback BeforeResolvingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.before.Add(abstract, callback)
}

// BeforeResolvingAny registers a global before-resolving callback.
func (c *App) BeforeResolvingAny(callback BeforeResolvingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.before.AddGlobal(callback)
}

// Resolving registers a callback that fires when the given abstract is being
// resolved.
func (c *App) Resolving(abstract string, callback BindingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.resolv.Add(abstract, callback)
}

// ResolvingAny registers a global resolving callback.
func (c *App) ResolvingAny(callback BindingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.resolv.AddGlobal(callback)
}

// AfterResolving registers a callback that fires after the given abstract is
// resolved.
func (c *App) AfterResolving(abstract string, callback BindingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.after.Add(abstract, callback)
}

// AfterResolvingAny registers a global after-resolving callback.
func (c *App) AfterResolvingAny(callback BindingCallback) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.after.AddGlobal(callback)
}

// fireCallbacks invokes each callback with the given instance and container.
// Callers must not hold the lock: callbacks re-enter the container.
func fireCallbacks(callbacks []BindingCallback, instance any, c *App) {
	for _, cb := range callbacks {
		cb(instance, c)
	}
}

// fireBeforeCallbacks invokes each before-resolving callback. Callers must not
// hold the lock: callbacks re-enter the container.
func fireBeforeCallbacks(callbacks []BeforeResolvingCallback, abstract string, params map[string]any, c *App) {
	for _, cb := range callbacks {
		cb(abstract, params, c)
	}
}
