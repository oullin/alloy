# Directory Structure

<!-- ref: @alloy/code-0171 -->

Alloy doesn't enforce a project layout. The packages are agnostic to
where your code lives. But every Alloy app ends up with similar bones,
because the framework's lifecycle pushes you toward a small set of natural
seams: an entry point, a bootstrap module, route definitions, models, and
configuration.

The reference layout is the demo app at
[`services/demo/`](https://github.com/oullin/alloy/tree/main/services/demo).
This page explains what each piece is for and where to put your own code.

## The Reference Tree

```
services/demo/
├── cmd/
│   └── api/
│       └── main.go              ← process entry point: signals + ctx + Run()
├── api/
│   ├── bootstrap.go             ← Options struct, StandardProviders, NewApplication
│   ├── server.go                ← NewHandler + Run: builds app, mounts routes, listens
│   ├── app/
│   │   └── models/              ← domain types (User, Order, …)
│   │       └── user.go
│   ├── config/
│   │   └── config.go            ← typed app config (App.Name, .Env, .Key, .URL)
│   ├── database/
│   │   ├── migrations/          ← schema migrations run at boot
│   │   └── seeders/             ← seed data
│   └── routes/
│       └── web.go               ← RegisterWeb(router, application)
├── storage/                     ← runtime data (sqlite file, uploads)
├── go.mod
└── README.md
```

## Where to Put Each Kind of Code

### Entry point — `cmd/<binary>/main.go`

One file. It builds a context with signal handling, calls into the
runner, and exits on error:

```go
// services/demo/cmd/api/main.go
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    if err := api.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

Keep this file boring. Anything substantive belongs one layer down.

If your app produces multiple binaries (HTTP server, queue worker, CLI),
each gets its own subdirectory under `cmd/`.

### Bootstrap — `<pkg>/bootstrap.go`

Three things: the `Options` struct that configures the app, the
`StandardProviders` function that returns your provider stack, and
`NewApplication` that ties them together.

```go
// services/demo/api/bootstrap.go:177
func NewApplication(opts ...Options) *container.Application {
    application, err := newApplication(opts...)
    if err != nil { panic(err) }
    return application
}

func newApplication(opts ...Options) (*container.Application, error) {
    o := resolveOptions(opts)
    application := container.NewApplication()

    application.RegisterMany(StandardProviders(application, o))
    application.Boot()

    if err := configureSkeleton(application, o); err != nil {
        return nil, err
    }

    return application, nil
}
```

This is the file every Alloy app needs. Everything else flows from it.
See [Application Bootstrap](/architecture/application).

### Server — `<pkg>/server.go`

Mounts routes and runs `http.Server`. The interesting parts are
`NewHandler` (builds the application, resolves the router, mounts routes)
and `Run` (wraps the handler in a server that respects the context):

```go
// services/demo/api/server.go:18
func NewHandler(opts ...Options) (http.Handler, error) {
    application, err := newApplication(opts...)
    if err != nil { return nil, err }

    router := container.Resolve[*routing.Router]("router")
    routes.RegisterWeb(router, application)

    return routingx.NewHandler(router), nil
}
```

### Routes — `routes/web.go`, `routes/api.go`, …

A function per route file. Each one takes the router and the application
and registers handlers:

```go
// services/demo/api/routes/web.go:16
func RegisterWeb(router *routing.Router, application *container.Application) {
    router.Get("/",  homeHandler(application))
    router.Get("/up", healthHandler())
}
```

The convention from upstream is to split routes by audience: `web.go` for
session-using HTML routes, `api.go` for stateless JSON endpoints, `console.go`
for CLI commands. That split scales as the app grows; a brand-new app can
start with just `web.go`.

### Models — `app/models/`

Domain types. Each file owns one aggregate (User, Order, Subscription)
and the database queries against it. Alloy doesn't ship an ORM —
models are plain Go structs with explicit query methods.

The demo's only model is `services/demo/api/app/models/user.go`. Add
yours alongside it.

### Configuration — `config/`

Typed config shapes that the app reads at runtime, registered with the
container:

```go
// services/demo/api/config/config.go
type App struct {
    Name string
    Env  string
    Key  string
    URL  string
}

func DefaultApp(env, key string) App { /* ... */ }
```

Bind values into the container with `Container.Instance(...)` and read
them back where you need them. See
[Configuration](/architecture/configuration).

### Database — `database/migrations/`, `database/seeders/`

Migrations run at boot when `o.RunMigrations` is true; seeders run when
`o.Seed` is true. Both are ordinary Go files exposing a `Run(*sql.DB) error`
function. The demo wires these into `configureSkeleton`:

```go
// services/demo/api/bootstrap.go:226
if o.RunMigrations != nil && *o.RunMigrations {
    if err := demomigrations.Run(db); err != nil {
        db.Close()
        return err
    }
}
```

### Storage — `storage/`

Runtime data the app writes during normal operation: a SQLite database
file, uploaded user content, generated artifacts. The path is configured
by `Options.StoragePath` (defaults to `<basepath>/storage`). Add it to
`.gitignore`; check in only `.gitkeep`.

## Adding a Service Provider

When your app grows past the standard providers — say, you need a
recommendation engine that lives across multiple handlers — write your
own provider:

```
services/myapp/
├── api/
│   ├── bootstrap.go
│   ├── providers/
│   │   └── recommendation_provider.go    ← NewRecommendationProvider, Register, Boot
│   └── ...
```

Then append it to `StandardProviders` (or a layer above):

```go
providers := append(
    api.StandardProviders(application, o),
    providers.NewRecommendationProvider(application.Container),
)
application.RegisterMany(providers)
```

See [Service Providers](/architecture/service-providers) for what goes
in the provider's `Register()` and `Boot()`.

## Where Alloy Itself Lives

If you're contributing to Alloy or reading its source as you build:

```
alloy/
├── packages/
│   ├── foundation/   ← Go packages — auth, cache, container, log, …
│   ├── console/      ← TypeScript package
│   ├── money/        ← TypeScript package
│   ├── tempo/        ← TypeScript package and acceptance tests
│   └── workflow/     ← TypeScript package
├── web/
│   ├── inertia-demo/ ← reference Inertia demo app
│   └── storage/      ← local runtime and cache data
├── infra/            ← repo automation
└── vite.config.ts
```

The split is: `pkg/hub/` contains reusable Go packages, other
`packages/*` directories contain TypeScript packages, and `web/` contains the
documentation site, web demos, and local runtime data.

## See Also

- [Application Bootstrap](/architecture/application) — what
  `bootstrap.go` actually does.
- [Service Providers](/architecture/service-providers) — when to add a
  `providers/` directory.
- [Configuration](/architecture/configuration) — how the demo's
  `Options` struct flows through the rest of the layout.
