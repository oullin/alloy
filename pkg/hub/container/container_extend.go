package container

// Extend registers a callback that modifies a resolved instance. If the
// abstract already has a cached instance, the extender is applied immediately.
func (c *App) Extend(abstract string, extender ExtenderFunc) {
	c.mu.Lock()

	abs := c.aliases.Resolve(abstract)

	if inst, ok := c.instances[abs]; ok {
		c.mu.Unlock()

		newInst, err := extender(inst, c)

		if err == nil {
			c.mu.Lock()
			c.instances[abs] = newInst
			c.mu.Unlock()

			c.rebound(abs)
		}

		return
	}

	c.extenders[abs] = append(c.extenders[abs], extender)

	c.mu.Unlock()

	if c.Resolved(abs) {
		c.rebound(abs)
	}
}

// ForgetExtenders removes all extension callbacks for the given abstract.
func (c *App) ForgetExtenders(abstract string) {
	c.mu.Lock()

	defer c.mu.Unlock()

	delete(c.extenders, c.aliases.Resolve(abstract))
}
