# Database Migrations

`Migrate` is a small wrapper around `golang-migrate`. It reads migration files
from any `fs.FS`, including an application `embed.FS`, and applies every pending
`up` migration through a database driver supplied by the caller.

Use it when an application wants Alloy to own migration execution without
making the foundation package depend on a specific database driver.

## File Convention

Migration files follow the `golang-migrate` convention:

```text
migrations/
  1_create_widgets.up.sql
  1_create_widgets.down.sql
  2_seed_widgets.up.sql
  2_seed_widgets.down.sql
```

The important shape is `{version}_{title}.up.sql`, with matching `.down.sql`
files when the application supports rollbacks.

## Example

```go
package main

import (
    "database/sql"
    "embed"
    "fmt"

    "github.com/golang-migrate/migrate/v4/database/sqlite"
    "github.com/oullin/alloy/pkg/hub/database"
    _ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
    db, err := sql.Open("sqlite", "file:app.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    driver, err := sqlite.WithInstance(db, &sqlite.Config{})
    if err != nil {
        panic(err)
    }

    if err := database.Migrate(migrations, "migrations", driver, "sqlite"); err != nil {
        panic(err)
    }

    fmt.Println("migrations applied")
}
```

## Semantics

- `sourceFS` can be any `fs.FS`; `embed.FS` is the common production shape.
- `dir` is the directory inside `sourceFS` that contains migration files.
- The database driver is supplied by the caller. For example, use
  `sqlite.WithInstance`, `postgres.WithInstance`, or another
  `golang-migrate` database driver in the application.
- `databaseName` must match the supplied driver name, such as `sqlite` or
  `postgres`.
- `Migrate` returns nil when `golang-migrate` reports `ErrNoChange`, so running
  it against an up-to-date schema is idempotent.
- Broken SQL, missing migration directories, and migrator setup failures are
  returned as wrapped errors.
- `Migrate` does not close the supplied driver or its underlying `*sql.DB`.
  The caller owns that lifecycle.

## Driver-Agnostic Wiring

The wrapper intentionally imports only `golang-migrate` core packages. Keep
database-specific imports in the application or service package that owns the
actual connection.

