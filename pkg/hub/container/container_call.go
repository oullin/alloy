package container

import "fmt"

// Call invokes the given callable, passing the container and parameters.
func (c *App) Call(callable MethodCallable, parameters map[string]any) (any, error) {
	return callable(c, parameters)
}

// Wrap returns a closure that invokes the given callable with the container and
// parameters when called.
func (c *App) Wrap(callable MethodCallable, parameters map[string]any) func() (any, error) {
	return func() (any, error) {
		return c.Call(callable, parameters)
	}
}

// BindMethod registers a callable for a named method binding.
func (c *App) BindMethod(method string, callback MethodCallable) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.methodBindings[method] = callback
}

// HasMethodBinding reports whether a method binding is registered.
func (c *App) HasMethodBinding(method string) bool {
	c.mu.RLock()

	defer c.mu.RUnlock()

	_, ok := c.methodBindings[method]

	return ok
}

// CallMethodBinding invokes the registered method binding. The instance is
// passed in the parameters map under the key "_instance".
func (c *App) CallMethodBinding(method string, instance any) (any, error) {
	c.mu.RLock()
	cb, ok := c.methodBindings[method]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrMethodNotBound, method)
	}

	return cb(c, map[string]any{"_instance": instance})
}
