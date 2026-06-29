// Package container is Alloy's IoC container and application kernel.
//
// It provides two main types:
//
//   - App: a service container with bindings,
//     singletons, scoped lifetimes, contextual bindings, tagging, extension,
//     resolving callbacks, and method invocation. App is safe for
//     concurrent use.
//
//   - Application: a thin wrapper around App that adds service
//     provider lifecycle management. Use Application as your composition
//     root; reach into the embedded *App only when you need APIs the
//     Application doesn't shadow (e.g. Tag, Extend, AfterResolving).
//
// # Quickstart
//
//	app := container.NewApplication()
//	app.Register(events.NewEventsServiceProvider(app.App))
//	app.Register(cache.NewCacheServiceProvider(app.App, "array"))
//	app.Boot()
//
//	mgr, _ := app.Make("cache")
//
// # Service providers
//
// See package alloy.dev/api/container/contracts/provider for the
// provider contract. The Application supports five lifecycle hooks:
//
//   - Register:  called once per provider when it is added.
//   - Boot:      called once per Bootable provider after Register completes.
//   - Deferred:  opt-in lazy registration; Register runs on first Make.
//   - DependsOn: declarative ordering; RegisterMany sorts dependencies first.
//   - Provides:  introspection hint, also required for deferred providers.
//
// # Lazy resolution patterns
//
// When a provider's binding factory needs another service that may be
// registered later, resolve it lazily inside the factory closure rather
// than at registration time:
//
//	p.app.Singleton("bus", func(c *container.App) (any, error) {
//	    queue, err := c.Make("queue.connection")
//	    if errors.Is(err, container.ErrNotBound) {
//	        return NewDispatcher(nil, nil), nil
//	    }
//	    if err != nil { return nil, err }
//	    return NewDispatcher(queue.(queue.Backend), nil), nil
//	})
//
// This makes registration order forgiving: as long as the dependency is
// bound before the dependent is RESOLVED, everything works.
//
// # Application vs App
//
// Application.Make shadows App.Make to flush deferred providers
// transparently. Code that holds an *Application gets deferred resolution;
// code that holds the embedded *App does not. This is intentional —
// the App has no notion of providers and stays simple.
package container
