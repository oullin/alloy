package container

import "errors"

var (
	// ErrNoApplication is returned when no process-wide Application has been
	// installed via SetApp.
	ErrNoApplication = errors.New("container: no Application installed; call container.SetApp(application) first")

	// ErrNotBound is returned when an abstract has no binding, instance, or alias.
	ErrNotBound = errors.New("container: abstract not bound")

	// ErrCircularDependency is returned when a circular dependency is detected
	// during resolution.
	ErrCircularDependency = errors.New("container: circular dependency detected")

	// ErrSelfAlias is returned when an alias points to itself.
	ErrSelfAlias = errors.New("container: alias cannot be the same as the abstract")

	// ErrAliasCycle is returned when registering an alias would close a loop in
	// the alias chain, which would make resolution non-terminating.
	ErrAliasCycle = errors.New("container: alias would create a cycle")

	// ErrMethodNotBound is returned when a method binding does not exist.
	ErrMethodNotBound = errors.New("container: method binding not found")
)
