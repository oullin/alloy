# Session Context

`session.NewContext` and `session.FromContext` store and retrieve the
per-request `*session.Store` on a `context.Context`.

Use these helpers when writing custom session middleware or integrations that
need downstream handlers to reach the request's session store without passing
it as a separate function argument.

## Example

```go
package main

import (
    "context"
    "fmt"

    "github.com/oullin/alloy/packages/foundation/session"
    "github.com/oullin/alloy/packages/foundation/session/handlers"
)

func main() {
    store := session.New("session", handlers.NewArrayHandler())
    ctx := session.NewContext(context.Background(), store)

    current, ok := session.FromContext(ctx)
    if !ok {
        panic("missing session store")
    }

    current.Put("status", "ready")
    fmt.Println(current.Get("status", "missing"))
}
```

## Middleware Example

```go
func AttachSession(store *session.Store, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := session.NewContext(r.Context(), store)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

Downstream handlers can retrieve the same store instance:

```go
store, ok := session.FromContext(r.Context())
if !ok {
    http.Error(w, "session unavailable", http.StatusInternalServerError)
    return
}

store.Put("user_id", 123)
```

## Semantics

- `NewContext` returns a child context carrying the exact `*session.Store`
  pointer provided by the caller.
- `FromContext` returns `(*session.Store, true)` when a store is present.
- `FromContext` returns `(nil, false)` for a bare context.
- The context key is an unexported struct type, so it cannot collide with keys
  from other packages.
- These helpers only expose the store. Starting, saving, regenerating, and
  closing the store remain the middleware or caller's responsibility.

