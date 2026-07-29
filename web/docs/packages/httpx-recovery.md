# HTTP Recovery Middleware

`HandleRecovery` is panic-recovery middleware for `net/http` handlers. It
keeps an unexpected panic from tearing down the request path, logs the panic
with a stack trace through `slog`, and returns a JSON 500 response when the
response has not already started.

Use it as an outer middleware around application routes, usually near the top
of the global middleware stack.

## Example

```go
package main

import (
    "fmt"
    "net/http"

    "hara.sh/alloy/httpx/middleware"
)

func main() {
    routes := http.NewServeMux()
    routes.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        panic("boom")
    })

    handler := middleware.NewHandleRecovery().Wrap(routes)

    fmt.Println("listening on http://127.0.0.1:8080")
    if err := http.ListenAndServe(":8080", handler); err != nil {
        panic(err)
    }
}
```

Requesting `/` logs the recovered panic and returns:

```json
{"error":"internal server error"}
```

## Semantics

- Panics other than `http.ErrAbortHandler` are recovered.
- Recovered panics are logged with message `recovery: panic recovered` and
  attributes `error` and `stack`.
- The middleware writes status `500` and `Content-Type: application/json` only
  when the wrapped handler has not already written the response.
- `http.ErrAbortHandler` is re-panicked so Go's HTTP server abort semantics are
  preserved.
- Normal responses pass through unchanged.

## Testing

Wrap the handler under test and call it with `httptest`:

```go
handler := middleware.NewHandleRecovery().Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    panic("boom")
}))

rec := httptest.NewRecorder()
handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
```

