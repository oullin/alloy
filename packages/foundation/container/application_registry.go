package container

import (
	"fmt"
	"sync"
)

var (
	applicationMu sync.RWMutex
	application   *Application
)

// SetApp installs the given Application as the process-wide instance.
// Pass nil to clear it (useful for tests).
func SetApp(app *Application) {
	applicationMu.Lock()

	defer applicationMu.Unlock()

	application = app
}

// Global returns the process-wide Application, or ErrNoApplication if SetApp
// has not been called.
func Global() (*Application, error) {
	applicationMu.RLock()

	defer applicationMu.RUnlock()

	if application == nil {
		return nil, ErrNoApplication
	}

	return application, nil
}

// MustGlobal returns the process-wide Application. It panics if SetApp has
// not been called; reserve it for wiring code that runs at startup, where a
// missing application is a programming error.
func MustGlobal() *Application {
	app, err := Global()

	if err != nil {
		panic(err)
	}

	return app
}

// HasApp reports whether a global Application has been installed.
func HasApp() bool {
	applicationMu.RLock()

	defer applicationMu.RUnlock()

	return application != nil
}

// Make resolves an abstract from the global application.
func Make(abstract string) (any, error) {
	app, err := Global()

	if err != nil {
		return nil, err
	}

	return app.Make(abstract)
}

// MustMake resolves an abstract from the global application. It panics on
// failure; reserve it for wiring code that runs at startup.
func MustMake(abstract string) any {
	v, err := Make(abstract)

	if err != nil {
		panic(fmt.Sprintf("container: MustMake(%q): %v", abstract, err))
	}

	return v
}

// Resolve is a generic, typed resolver. It returns ErrNoApplication when no
// global application is installed, the resolution error when the abstract
// cannot be made, or a type-mismatch error when the resolved value is not T.
func Resolve[T any](abstract string) (T, error) {
	var zero T

	raw, err := Make(abstract)

	if err != nil {
		return zero, err
	}

	v, ok := raw.(T)

	if !ok {
		return zero, fmt.Errorf("container: Resolve[%T](%q): wrong type %T", zero, abstract, raw)
	}

	return v, nil
}

// MustResolve is the panicking variant of Resolve. Reserve it for wiring
// code that runs at startup.
func MustResolve[T any](abstract string) T {
	v, err := Resolve[T](abstract)

	if err != nil {
		panic(err)
	}

	return v
}
