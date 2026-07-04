# HTTP Request Log Middleware

`HandleRequestLog` records completed HTTP requests through `slog`. It wraps the
response writer to capture the final status code and number of bytes written,
then logs after the downstream handler returns.

Use it around HTTP routes when you want uniform request logs without changing
individual handlers.

## Example

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/oullin/alloy/packages/foundation/httpx/middleware"
)

func main() {
    routes := http.NewServeMux()
    routes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusCreated)
        _, _ = w.Write([]byte("created"))
    })
    routes.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNoContent)
    })

    logger := middleware.NewHandleRequestLog(middleware.RequestLogOptions{
        SkipPaths: []string{"/health"},
    })

    fmt.Println("listening on http://127.0.0.1:8080")
    if err := http.ListenAndServe(":8080", logger.Wrap(routes)); err != nil {
        panic(err)
    }
}
```

Requests to `/` emit one `slog` record with message `http: request` and the
attributes `method`, `path`, `status`, `duration`, and `bytes`. Requests to
`/health` pass through without a request-log record.

## Semantics

- `SkipPaths` is an exact path match against `r.URL.Path`.
- If a handler writes a body without calling `WriteHeader`, the logged status
  is `200`.
- The logged byte count is the number of bytes successfully written by the
  wrapped response writer.
- The middleware logs after `next.ServeHTTP` returns, so duration includes the
  downstream handler.
- When the original writer implements `http.Flusher`, the wrapped writer also
  implements `http.Flusher` and forwards `Flush`.
- The wrapped writer implements `Unwrap() http.ResponseWriter` so callers can
  recover the original writer when they need to.

## Testing

Install a test `slog` handler, call the wrapped handler, and inspect the
captured record attributes:

```go
handler := middleware.NewHandleRequestLog(middleware.RequestLogOptions{}).Wrap(
    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte("ok"))
    }),
)

handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items", nil))
```

