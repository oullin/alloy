package container

import (
	"sync"

	"github.com/oullin/alloy/pkg/hub/container/internal/alias"
	"github.com/oullin/alloy/pkg/hub/container/internal/callbacks"
)

// Factory creates a value from the container.
type Factory func(c *App) (any, error)

// BindingCallback is called during resolving lifecycle events.
type BindingCallback func(instance any, c *App)

// BeforeResolvingCallback is called before resolution begins.
type BeforeResolvingCallback func(abstract string, parameters map[string]any, c *App)

// ExtenderFunc modifies a resolved instance.
type ExtenderFunc func(instance any, c *App) (any, error)

// MethodCallable represents a function that can be called with dependency
// injection via the container.
type MethodCallable func(c *App, params map[string]any) (any, error)

// Binding holds a registered factory and its lifecycle metadata.
type Binding struct {
	factory Factory
	shared  bool
	scoped  bool
}

// App is a inversion-of-control container. It manages
// service bindings, resolution, contextual bindings, tagging, extension,
// lifecycle callbacks, and method invocation. All methods are safe for
// concurrent use.
type App struct {
	mu sync.RWMutex

	bindings       map[string]Binding
	instances      map[string]any
	aliases        *alias.Table
	resolved       map[string]bool
	buildStack     []string
	with           []map[string]any
	contextual     map[string]map[string]any
	tags           map[string][]string
	extenders      map[string][]ExtenderFunc
	reboundCbs     map[string][]BindingCallback
	methodBindings map[string]MethodCallable

	before *callbacks.Registry[BeforeResolvingCallback]
	resolv *callbacks.Registry[BindingCallback]
	after  *callbacks.Registry[BindingCallback]
}

// New creates an empty, fully initialized App.
func New() *App {
	return &App{
		bindings:       make(map[string]Binding),
		instances:      make(map[string]any),
		aliases:        alias.NewTable(),
		resolved:       make(map[string]bool),
		contextual:     make(map[string]map[string]any),
		tags:           make(map[string][]string),
		extenders:      make(map[string][]ExtenderFunc),
		reboundCbs:     make(map[string][]BindingCallback),
		methodBindings: make(map[string]MethodCallable),
		before:         callbacks.NewRegistry[BeforeResolvingCallback](),
		resolv:         callbacks.NewRegistry[BindingCallback](),
		after:          callbacks.NewRegistry[BindingCallback](),
	}
}

// Flush resets the container to an empty state, clearing all bindings,
// instances, aliases, resolved flags, callbacks, and method bindings.
func (c *App) Flush() {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.bindings = make(map[string]Binding)
	c.instances = make(map[string]any)
	c.aliases.Reset()
	c.resolved = make(map[string]bool)
	c.buildStack = nil
	c.with = nil
	c.contextual = make(map[string]map[string]any)
	c.tags = make(map[string][]string)
	c.extenders = make(map[string][]ExtenderFunc)
	c.reboundCbs = make(map[string][]BindingCallback)
	c.methodBindings = make(map[string]MethodCallable)
	c.before.Reset()
	c.resolv.Reset()
	c.after.Reset()
}
