# Database Transactions

`WithTx` runs a callback inside a SQL transaction. It centralizes the usual
begin, rollback, commit, and panic handling around `*sql.Tx`.

Use it when a unit of database work must commit as one operation or roll back
as one operation.

## Example

```go
package main

import (
    "context"
    "database/sql"
    "fmt"

    "hara.sh/alloy/database"
    _ "modernc.org/sqlite"
)

func main() {
    ctx := context.Background()

    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    if _, err := db.ExecContext(ctx, "CREATE TABLE entries (name TEXT NOT NULL)"); err != nil {
        panic(err)
    }

    err = database.WithTx(ctx, db, nil, func(tx *sql.Tx) error {
        _, err := tx.ExecContext(ctx, "INSERT INTO entries (name) VALUES ('committed')")
        return err
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("committed")
}
```

## Semantics

- `db` can be any value implementing `BeginTx(context.Context, *sql.TxOptions)`.
  A plain `*sql.DB` satisfies this interface.
- `opts` is passed directly to `BeginTx`; pass `nil` for the driver default.
- If beginning the transaction fails, the returned error is wrapped with
  `database: begin tx`.
- If the callback returns an error, `WithTx` rolls back and returns the
  callback error.
- If rollback also fails after a callback error, the returned error joins the
  callback error and a wrapped `database: rollback tx` error.
- If the callback panics, `WithTx` rolls back and then re-panics with the
  original panic value.
- If the callback returns nil and commit fails, the returned error is wrapped
  with `database: commit tx`.

## Error Handling

Because callback errors are returned directly, callers can keep using
`errors.Is` with their own sentinel errors:

```go
var errRejected = errors.New("rejected")

err := database.WithTx(ctx, db, nil, func(tx *sql.Tx) error {
    return errRejected
})

if errors.Is(err, errRejected) {
    // The transaction was rolled back.
}
```
