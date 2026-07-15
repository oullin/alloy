package container

import "slices"

// Rebinding registers a callback that fires when the given abstract is rebound.
// If the abstract is already bound, the callback is invoked immediately with
// the current instance.
func (c *App) Rebinding(abstract string, callback BindingCallback) (any, error) {
	c.mu.Lock()

	abs := c.aliases.Resolve(abstract)
	c.reboundCbs[abs] = append(c.reboundCbs[abs], callback)

	wasBound := c.isBound(abs)

	c.mu.Unlock()

	if wasBound {
		instance, err := c.Make(abstract)

		if err != nil {
			return nil, err
		}

		callback(instance, c)

		return instance, nil
	}

	return nil, nil
}

// Refresh is a convenience wrapper around Rebinding. When the abstract is
// rebound the setter is called with the new instance.
func (c *App) Refresh(abstract string, setter func(any)) {
	c.Rebinding(abstract, func(instance any, _ *App) { //nolint:errcheck
		setter(instance)
	})
}

// rebound fires all registered rebound callbacks for the given abstract.
func (c *App) rebound(abstract string) {
	c.mu.RLock()

	abs := c.aliases.Resolve(abstract)
	cbs := slices.Clone(c.reboundCbs[abs])

	c.mu.RUnlock()

	if len(cbs) == 0 {
		return
	}

	instance, err := c.Make(abstract)

	if err != nil {
		return
	}

	for _, cb := range cbs {
		cb(instance, c)
	}
}
