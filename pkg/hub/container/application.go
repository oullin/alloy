package container

import (
	"sync"

	"github.com/oullin/alloy/pkg/hub/container/contracts/provider"
)

// Application wraps App and manages service provider lifecycle,
// in the order they are added; Boot is called after all registrations.
//
// Two optional capabilities extend the lifecycle:
//
//   - provider.Deferred: providers that delay Register() until one of
//     their declared abstracts is first resolved through Application.Make
//     (or any helper that goes through it: container.Resolve, facades).
//   - provider.DependsOn: providers that declare prerequisite abstracts.
//     RegisterMany topologically sorts deps before calling Register.
//
// Deferred resolution is an Application-level feature: code that bypasses
// the Application and calls App.Make directly will see ErrNotBound
// for keys whose providers have not yet been flushed. This is intentional —
// it keeps the App itself free of provider lifecycle concerns.
type Application struct {
	*App
	mu            sync.Mutex
	providers     []provider.ServiceProvider
	deferredByKey map[string]provider.ServiceProvider
	registered    map[provider.ServiceProvider]bool
	booted        bool
	inFlight      map[provider.ServiceProvider]*sync.WaitGroup
}

// NewApplication creates an Application backed by a fresh App.
func NewApplication() *Application {
	return &Application{
		App:           New(),
		deferredByKey: make(map[string]provider.ServiceProvider),
		registered:    make(map[provider.ServiceProvider]bool),
		inFlight:      make(map[provider.ServiceProvider]*sync.WaitGroup),
	}
}
