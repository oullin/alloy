# HTTP Server Runner

`server.Run` and `server.RunListener` run an `http.Server` until serving fails
or a context is canceled. On context cancellation they attempt graceful
shutdown, enforce a timeout, and fall back to `Close` when graceful shutdown
cannot complete.

Use these helpers for application entrypoints that should stop cleanly on
process shutdown signals.

## Run Example

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/oullin/alloy/packages/foundation/httpx/server"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    srv := &http.Server{
        Addr:    ":8080",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            _, _ = w.Write([]byte("ok"))
        }),
    }

    fmt.Println("listening on http://127.0.0.1:8080")
    if err := server.Run(ctx, srv, 10*time.Second); err != nil {
        panic(err)
    }
}
```

## RunListener Example

```go
ln, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
    panic(err)
}

srv := &http.Server{Handler: handler}
err = server.RunListener(ctx, srv, ln, 5*time.Second)
```

`RunListener` is useful in tests or supervisors that create the listener
before handing it to the HTTP server.

## Semantics

- Passing a nil `*http.Server` returns `server: nil http server`.
- `Run` serves with `srv.ListenAndServe`.
- `RunListener` serves with `srv.Serve(ln)`.
- A non-positive shutdown timeout uses the default timeout of 5 seconds.
- If serving returns before context cancellation, the serve error is returned
  after normalization.
- `http.ErrServerClosed` is treated as nil.
- On context cancellation, the helper calls `srv.Shutdown` with a fresh
  timeout context.
- If graceful shutdown times out or fails, the helper calls `srv.Close` as a
  fallback and returns a wrapped `server: graceful shutdown failed` error.
- If shutdown succeeds, in-flight requests can complete before the helper
  returns.

## Testing

Use `RunListener` with a loopback listener to get a real URL without binding a
fixed port:

```go
ln, _ := net.Listen("tcp", "127.0.0.1:0")
ctx, cancel := context.WithCancel(context.Background())

srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    _, _ = w.Write([]byte("ok"))
})}

go func() {
    _ = server.RunListener(ctx, srv, ln, time.Second)
}()

cancel()
```

