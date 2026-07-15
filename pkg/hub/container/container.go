package container

import "sync"

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

// App is an inversion-of-control container. It manages
// service bindings, resolution, contextual bindings, tagging, extension,
// lifecycle callbacks, and method invocation. All methods are safe for
// concurrent use. Note that a zero-value App is not usable; always construct
// one using New(). Also, mutually-dependent shared bindings resolved concurrently
// from different goroutines may block rather than return ErrCircularDependency.
type App struct {
	mu *sync.RWMutex

	bindings        map[string]Binding
	instances       map[string]any
	aliases         map[string]string
	abstractAliases map[string][]string
	resolved        map[string]bool
	resolution      *resolution
	parent          *App
	sf              *singleflight
	contextual      map[string]map[string]any
	tags            map[string][]string
	extenders       map[string][]ExtenderFunc
	reboundCbs      map[string][]BindingCallback
	methodBindings  map[string]MethodCallable

	globalBeforeCbs []BeforeResolvingCallback
	beforeCbs       map[string][]BeforeResolvingCallback
	globalResolvCbs []BindingCallback
	resolvCbs       map[string][]BindingCallback
	globalAfterCbs  []BindingCallback
	afterCbs        map[string][]BindingCallback
}

// resolution holds the goroutine-local build stack and parameter stack
// for a dependency resolution chain.
type resolution struct {
	buildStack []string
	with       []map[string]any
	done       bool
}

// New creates an empty, fully initialized App.
func New() *App {
	return &App{
		mu:              new(sync.RWMutex),
		bindings:        make(map[string]Binding),
		instances:       make(map[string]any),
		aliases:         make(map[string]string),
		abstractAliases: make(map[string][]string),
		resolved:        make(map[string]bool),
		sf:              new(singleflight),
		contextual:      make(map[string]map[string]any),
		tags:            make(map[string][]string),
		extenders:       make(map[string][]ExtenderFunc),
		reboundCbs:      make(map[string][]BindingCallback),
		methodBindings:  make(map[string]MethodCallable),
		beforeCbs:       make(map[string][]BeforeResolvingCallback),
		resolvCbs:       make(map[string][]BindingCallback),
		afterCbs:        make(map[string][]BindingCallback),
	}
}

// Flush resets the container to an empty state, clearing all bindings,
// instances, aliases, resolved flags, callbacks, and method bindings.
func (c *App) Flush() {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.bindings = make(map[string]Binding)
	c.instances = make(map[string]any)
	c.aliases = make(map[string]string)
	c.abstractAliases = make(map[string][]string)
	c.resolved = make(map[string]bool)
	c.resolution = nil
	c.contextual = make(map[string]map[string]any)
	c.tags = make(map[string][]string)
	c.extenders = make(map[string][]ExtenderFunc)
	c.reboundCbs = make(map[string][]BindingCallback)
	c.methodBindings = make(map[string]MethodCallable)
	c.globalBeforeCbs = nil
	c.beforeCbs = make(map[string][]BeforeResolvingCallback)
	c.globalResolvCbs = nil
	c.resolvCbs = make(map[string][]BindingCallback)
	c.globalAfterCbs = nil
	c.afterCbs = make(map[string][]BindingCallback)
}
