package container

import (
	"fmt"
	"slices"
)

// Make resolves the given abstract from the container.
func (c *App) Make(abstract string) (any, error) {
	return c.resolve(abstract, nil)
}

// MakeWith resolves the given abstract, passing parameters to the factory via
// the parameter override stack.
func (c *App) MakeWith(abstract string, parameters map[string]any) (any, error) {
	return c.resolve(abstract, parameters)
}

// Build executes the given factory directly. It is useful for one-off
// instantiation without registering a binding.
func (c *App) Build(factory Factory) (any, error) {
	return factory(c)
}

// Get resolves the abstract. If the abstract is not bound, it returns
// ErrNotBound (PSR-11 parity).
func (c *App) Get(abstract string) (any, error) {
	if c.Has(abstract) {
		return c.resolve(abstract, nil)
	}

	return nil, fmt.Errorf("%w: %q", ErrNotBound, abstract)
}

// FactoryFunc returns a closure that resolves the abstract each time it is
// called.
func (c *App) FactoryFunc(abstract string) func() (any, error) {
	return func() (any, error) {
		return c.Make(abstract)
	}
}

// resolve is the core resolution engine.
func (c *App) resolve(abstract string, parameters map[string]any) (any, error) {
	c.mu.Lock()

	original := abstract
	abstract = c.aliases.Resolve(abstract)

	// Snapshot before-resolving callbacks. This must stay above the cached
	// instance early return below, which fires these and only these.
	beforeGlobal, beforeSpecific := c.before.Snapshot(abstract)

	// Check contextual binding first (takes precedence over cached instances).
	concrete := c.getContextualConcrete(abstract)

	if concrete == nil && original != abstract {
		concrete = c.getContextualConcrete(original)
	}

	// Check for cached instance when no parameters are given and no
	// contextual override exists.
	if concrete == nil && len(parameters) == 0 {
		if inst, ok := c.instances[abstract]; ok {
			c.mu.Unlock()

			fireBeforeCallbacks(beforeGlobal, abstract, parameters, c)
			fireBeforeCallbacks(beforeSpecific, abstract, parameters, c)

			return inst, nil
		}
	}

	// Determine the concrete factory.
	var factory Factory

	if concrete != nil {
		if f, ok := concrete.(Factory); ok {
			factory = f
		} else {
			val := concrete
			factory = func(_ *App) (any, error) { return val, nil }
		}
	} else if b, ok := c.bindings[abstract]; ok {
		factory = b.factory
	}

	if factory == nil {
		c.mu.Unlock()

		return nil, fmt.Errorf("%w: %q", ErrNotBound, abstract)
	}

	// Circular dependency detection.
	if slices.Contains(c.buildStack, abstract) {
		// Snapshot the stack while still holding the lock; otherwise
		// the deferred fmt.Errorf read would race with concurrent
		// resolve() calls mutating c.buildStack at lines 265/283.
		stackSnapshot := slices.Clone(c.buildStack)
		c.mu.Unlock()

		return nil, fmt.Errorf("%w: %q (build stack: %v)", ErrCircularDependency, abstract, stackSnapshot)
	}

	// Capture binding metadata before unlocking.
	b := c.bindings[abstract]
	extenders := slices.Clone(c.extenders[abstract])
	resolvGlobal, resolvSpecific := c.resolv.Snapshot(abstract)
	afterGlobal, afterSpecific := c.after.Snapshot(abstract)

	c.buildStack = append(c.buildStack, abstract)

	// Always push parameters (even nil) so nested Make calls get their own
	// scope and don't inherit the parent's parameters.
	c.with = append(c.with, parameters)

	c.mu.Unlock()

	fireBeforeCallbacks(beforeGlobal, abstract, parameters, c)
	fireBeforeCallbacks(beforeSpecific, abstract, parameters, c)

	// Execute factory.
	instance, err := factory(c)

	c.mu.Lock()

	// Pop build stack.
	if len(c.buildStack) > 0 {
		c.buildStack = c.buildStack[:len(c.buildStack)-1]
	}

	if len(c.with) > 0 {
		c.with = c.with[:len(c.with)-1]
	}

	c.mu.Unlock()

	if err != nil {
		return nil, err
	}

	// Apply extenders.
	for _, ext := range extenders {
		instance, err = ext(instance, c)

		if err != nil {
			return nil, err
		}
	}

	// Cache shared instances.
	if b.shared && len(parameters) == 0 {
		c.mu.Lock()
		c.instances[abstract] = instance
		c.mu.Unlock()
	}

	c.mu.Lock()
	c.resolved[abstract] = true
	c.mu.Unlock()

	// Fire resolving callbacks.
	fireCallbacks(resolvGlobal, instance, c)
	fireCallbacks(resolvSpecific, instance, c)

	// Fire after-resolving callbacks.
	fireCallbacks(afterGlobal, instance, c)
	fireCallbacks(afterSpecific, instance, c)

	return instance, nil
}

// Parameters returns the current parameter override map from the top of the
// with stack. Factories can call this to access parameters passed via MakeWith.
func (c *App) Parameters() map[string]any {
	c.mu.RLock()

	defer c.mu.RUnlock()

	if len(c.with) == 0 {
		return nil
	}

	return c.with[len(c.with)-1]
}
