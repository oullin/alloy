package container

import (
	"fmt"
	"slices"
)

// withResolution returns a cloned App referencing the given resolution context.
func (c *App) withResolution(r *resolution) *App {
	c.mu.RLock()

	defer c.mu.RUnlock()

	clone := *c
	clone.resolution = r

	if c.parent == nil {
		clone.parent = c
	}

	return &clone
}

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
	c.mu.RLock()
	needsNewRes := c.resolution == nil || c.resolution.done
	c.mu.RUnlock()

	if needsNewRes {
		c = c.withResolution(&resolution{})
	}

	c.mu.Lock()

	original := abstract
	abstract = c.aliases.Resolve(abstract)

	// Fire before-resolving callbacks.
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
	if slices.Contains(c.resolution.buildStack, abstract) {
		// Snapshot the stack while still holding the lock; otherwise
		// the deferred fmt.Errorf read would race with concurrent
		// resolve() calls mutating c.resolution.buildStack.
		stackSnapshot := slices.Clone(c.resolution.buildStack)
		c.mu.Unlock()

		return nil, fmt.Errorf("%w: %q (build stack: %v)", ErrCircularDependency, abstract, stackSnapshot)
	}

	// Capture binding metadata before unlocking.
	b := c.bindings[abstract]
	extenders := slices.Clone(c.extenders[abstract])
	resolvGlobal, resolvSpecific := c.resolv.Snapshot(abstract)
	afterGlobal, afterSpecific := c.after.Snapshot(abstract)

	c.resolution.buildStack = append(c.resolution.buildStack, abstract)

	// Always push parameters (even nil) so nested Make calls get their own
	// scope and don't inherit the parent's parameters.
	c.resolution.with = append(c.resolution.with, parameters)

	c.mu.Unlock()

	fireBeforeCallbacks(beforeGlobal, abstract, parameters, c)
	fireBeforeCallbacks(beforeSpecific, abstract, parameters, c)

	// Execute factory.
	var instance any

	var err error

	runFactoryAndPop := func() (any, error) {
		inst, fErr := factory(c)

		if fErr == nil && inst == c {
			inst = c.parent
		}

		c.mu.Lock()

		if len(c.resolution.buildStack) > 0 {
			c.resolution.buildStack = c.resolution.buildStack[:len(c.resolution.buildStack)-1]
		}

		if len(c.resolution.with) > 0 {
			c.resolution.with = c.resolution.with[:len(c.resolution.with)-1]
		}

		if len(c.resolution.buildStack) == 0 {
			c.resolution.done = true
		}

		c.mu.Unlock()

		return inst, fErr
	}

	if b.shared && len(parameters) == 0 {
		instance, err = c.sf.Do(abstract, func() (any, error) {
			inst, fErr := runFactoryAndPop()

			if fErr != nil {
				return nil, fErr
			}

			// Apply extenders.
			for _, ext := range extenders {
				inst, fErr = ext(inst, c)

				if fErr != nil {
					return nil, fErr
				}
			}

			// Cache shared instances.
			c.mu.Lock()
			c.instances[abstract] = inst
			c.resolved[abstract] = true
			c.mu.Unlock()

			return inst, nil
		})

		// Pop build stack for waiting goroutines if they didn't run the inner func
		c.mu.Lock()

		if len(c.resolution.buildStack) > 0 && c.resolution.buildStack[len(c.resolution.buildStack)-1] == abstract {
			c.resolution.buildStack = c.resolution.buildStack[:len(c.resolution.buildStack)-1]

			if len(c.resolution.with) > 0 {
				c.resolution.with = c.resolution.with[:len(c.resolution.with)-1]
			}
		}

		if len(c.resolution.buildStack) == 0 {
			c.resolution.done = true
		}

		c.mu.Unlock()

		if err != nil {
			return nil, err
		}
	} else {
		instance, err = runFactoryAndPop()

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

		c.mu.Lock()
		c.resolved[abstract] = true
		c.mu.Unlock()
	}

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

	if c.resolution == nil || len(c.resolution.with) == 0 {
		return nil
	}

	return c.resolution.with[len(c.resolution.with)-1]
}
