# filesystem

<!-- ref: @alloy/code-0071 -->
<!-- ref: @alloy/code-0068 -->
<!-- ref: @alloy/code-0072 -->
<!-- ref: @alloy/code-0073 -->
<!-- ref: @alloy/code-0074 -->
<!-- ref: @alloy/code-0069 -->
<!-- ref: @alloy/code-0070 -->

<!-- ALLOY:HAND -->

## Introduction

The filesystem package gives every Alloy app one API for files and
directories on local disk, so product code does not reach for `os`
directly.

It exposes two types:

- **`Local`** — path-based operations, anywhere the process can reach.
  Use it for paths your own code controls.
- **`Rooted`** — the same operations confined to a root directory. Use
  it for any path that came from outside the program.

There is no service provider and no container binding: construct what
you need where you need it.

## Basic Usage

`Local` is stateless — `New()` takes no arguments, and every method that
touches the disk takes a `context.Context` first.

```go
fs := filesystem.New()
ctx := context.Background()

// Write, then read back.
err := fs.Put(ctx, "uploads/avatar-42.png", data)
raw, err := fs.Get(ctx, "uploads/avatar-42.png")

// Stream, for sources of unknown size.
err = fs.PutStream(ctx, "uploads/big.iso", reader)

// Copy and delete. Put, PutStream and Copy create missing parents.
err = fs.Copy(ctx, "uploads/avatar-42.png", "backup/avatar-42.png")
err = fs.Delete("backup/avatar-42.png")

// Move does not create them, so make the destination first.
err = fs.MakeDirectory("archive")
err = fs.Move("uploads/avatar-42.png", "archive/avatar-42.png")
```

Directory work uses `MakeDirectory`, `Files`, `Directories`, and
`DeleteDirectory`; see
[`filesystem_dir.go`](https://github.com/oullin/alloy/blob/main/pkg/hub/filesystem/filesystem_dir.go).

Two pairs are easy to mix up:

| Use | When |
| --- | --- |
| `MakeDirectory` | creates parents, succeeds if it already exists |
| `MakeExclusiveDirectory` | no parents, fails with `fs.ErrExist` if taken — an atomic claim on a name |
| `Delete` | files only, ignores missing ones |
| `DeleteAll` | files or trees, recursive, idempotent |

## Untrusted Paths

Never join a caller-supplied name onto a directory and pass it to
`Local`. `"../../etc/passwd"` is a valid filename, and so is a symlink
pointing anywhere. Checking the path first does not fix it: the check and
the open are two steps, and the tree can change in between.

`At` returns a `Rooted` that the operating system confines to a
directory, enforced against an open handle rather than by inspecting the
path:

```go
uploads, err := filesystem.At("/srv/app/uploads")
if err != nil {
    return err
}

defer uploads.Close()

// Refused: "..", absolute paths, and symlinks pointing outside the root.
_, err = uploads.Get(ctx, userSuppliedName)
```

A `Rooted` holds a file descriptor, so it must be closed.

Symlinks that stay inside the root are followed normally — `Rooted`
prevents escape, it does not ban links. If your policy is stricter than
that, check `LinkInfo` on the final component.

The guarantee stops at the filesystem: it does not prohibit crossing
mount points below the root, nor reading `/proc` or device files that
live inside it. Each of those needs an attacker who already controls the
root.

## Errors

Every method that can be handed a missing path reports it the same way,
whichever one you called:

```go
if _, err := fs.Get(ctx, path); errors.Is(err, fs.ErrNotExist) {
    // ...
}
```

`filesystem.ErrNotFound` wraps `fs.ErrNotExist`, so either matches, and
the underlying `*fs.PathError` is preserved — the failing path stays in
the message. Do not use `os.IsNotExist`: it predates `errors.Is` and does
not unwrap.

A `Rooted` escape attempt returns a `*fs.PathError` reading `path escapes
from parent`. It is deliberately **not** an `fs.ErrNotExist`, so a
refusal never looks like a missing file.

## Uploads

`filesystem.Local` and `filesystem.Rooted` both satisfy
`httpx/foundation.FileStore`, so either can back an uploaded file. Prefer
`Rooted` — the filename comes from the client:

```go
uploads, err := filesystem.At("/srv/app/uploads")
defer uploads.Close()

path, err := file.StoreAs(ctx, "avatars", file.HashName(), uploads)
```

## Extending

Implement the `Filesystem` interface in
[`contracts/filesystem`](https://github.com/oullin/alloy/blob/main/pkg/hub/contracts/filesystem/filesystem.go)
to back these operations with something other than local disk. `*Local`
is bound to it by a compile-time assertion, and a test keeps the two from
drifting apart.

Note the contract models **local disk**: it is context-aware and returns
`*os.File` from `MakeTempFile`. A cloud backend will fit some of it
awkwardly. If that is what you need, a narrower interface at your own
call site is usually the better tool — `FileStore` above is an example.

## Events

This package does not currently dispatch events. Wrap operations in
your own event-emitting decorator if you need observability.

## See Also

- [Service Providers](/architecture/service-providers).
<!-- /ALLOY:HAND -->

Package filesystem provides local filesystem operations including reading, writing, copying, moving, and deleting files and directories. It also supports file locking, MIME type detection, hashing, and permission management.

<div class="docs-callout docs-callout-upstream"></div>

<div class="docs-callout docs-callout-go">
  <strong>Go adaptation.</strong>
  </div>

## Installation

Install this module directly in applications that consume packages independently:

```bash
go get github.com/oullin/alloy/pkg/hub/filesystem@latest
```

When working inside this monorepo, use the repository workspace:

```bash
GOWORK=./pkg/hub/go.work go test -count=1 ./pkg/hub/filesystem/...
```

## Source Coverage

| Package      | Purpose                                                                                                                                                                                                                          |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `filesystem` | Package filesystem provides local filesystem operations including reading, writing, copying, moving, and deleting files and directories. It also supports file locking, MIME type detection, hashing, and permission management. |

## Core Concepts

The filesystem reference is organized around the exported Go surface for package `filesystem`. Start from the source coverage and public surface tables to identify the constructors, managers, interfaces, sentinel errors, and helper functions available to callers. Use the package tests as executable wiring examples for collaborators, default behavior.

### Public Surface

| Surface                    | Exported API                                                                                                                                                                                                                                                                              |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Types                      | `Local`, `Rooted`, `LockableFile`                                                                                                                                                                                                                                                         |
| Constructors and functions | `AllDirectories`, `AllFiles`, `Append`, `Basename`, `Chmod`, `CleanDirectory`, `Close`, `Copy`, `CopyDirectory`, `Delete`, `DeleteDirectories`, `DeleteDirectory`, `Directories`, `Dirname`, `EnsureDirectoryExists`, `ExclusiveLock`, `Exists`, `Extension`, `Files`, `Get`, and 38 more |
| Variables                  | `ErrHashAlgorithm`, `ErrLockFailed`, `ErrLocked`, `ErrNotDirectory`, `ErrNotFound`                                                                                                                                                                                                        |
| Constants                  | None exported from this package root.                                                                                                                                                                                                                                                     |

### Capability Matrix

| Capability                         | Documentation note                                                                                                   |
| ---------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Security-sensitive behavior        | Supported by exported API and package tests; use the API reference and parity tests below when wiring this behavior. |
| Serialization or transport formats | Supported by exported API and package tests; use the API reference and parity tests below when wiring this behavior. |

## Usage

Start with the package constructor or manager type when one is exported. Alloy keeps dependencies explicit, so callers should pass repositories, stores, handlers, dispatchers, clocks, or clients directly instead of relying on global framework state.

```go
package main

import (
    _ "github.com/oullin/alloy/pkg/hub/filesystem"
)

func main() {
    // Import the package you use, then wire the exported constructors,
    // managers, stores, handlers, or helpers required by your application.
}
```

Use package tests as executable examples when the exact constructor requires collaborators. The tests under `pkg/hub/filesystem` cover the supported creation paths, default values, and parity behavior.

## Configuration

Alloy documents behavior through Go options and constructor arguments:

| Upstream shape    | Alloy shape                                            |
| ----------------- | -------------------------------------------------------- |
| Config file keys  | Typed config structs, options, or constructor parameters |
| Facade defaults   | Explicit manager/default-driver setup                    |
| Service providers | Go service-provider structs or direct application wiring |
| Runtime helpers   | Package functions and interfaces                         |

Prefer narrow interfaces at package boundaries. When a package exposes a manager, register drivers or providers at startup, set the default once, and resolve named instances per request or job.

## Advanced Features

The package reference should be read through these parity lenses:

| Area              | Documentation coverage                                                                  |
| ----------------- | --------------------------------------------------------------------------------------- |
| Drivers/providers | Available implementations, default selection, custom registration, and failure behavior |
| Events            | Emitted structs, dispatcher hooks, listener timing, transaction or queue interaction    |
| Errors            | Exported sentinel errors, wrapping, and `errors.Is` compatibility                       |
| Context           | Which operations accept `context.Context` and how cancellation/deadlines propagate      |
| Testing           | Fakes, null implementations, assertion helpers, and deterministic clocks/stores         |

## Edge Cases

- Do not translate PHP-only behavior literally. If upstream depends on PHP traits, request globals, Template, CLI, or Orm magic, document the Alloy Go equivalent instead.
- Preserve error identity when the package exports sentinel errors; callers should be able to use `errors.Is` where the package promises it.
- Treat driver compatibility as observable behavior. Unsupported store/driver combinations should be documented as errors or explicit no-ops, never as silent omissions.
- For I/O paths, document cancellation and timeout behavior whenever the package accepts a `context.Context`.
- For test fakes, document whether assertions inspect recorded calls, stored payloads, emitted events, or rendered output.

## Testing

Run the package tests before changing examples:

```bash
GOWORK=./pkg/hub/go.work go test -count=1 ./pkg/hub/filesystem/...
```

Parity is tracked by these tests:

## API Reference

### Exported Types

| Type                        | Notes                                                                              |
| --------------------------- | ---------------------------------------------------------------------------------- |
| `Local`                     | Path-based local filesystem operations.                                            |
| `Rooted`                    | The same operations, confined to a root directory. Holds an fd; must be closed.     |
| `LockableFile`              | Source-backed public surface. See the Go package for exact signature and behavior. |

### Exported Functions

| Function                       | Notes                                                                              |
| ------------------------------ | ---------------------------------------------------------------------------------- |
| `AllDirectories`               | Source-backed public surface. See the Go package for exact signature and behavior. |
| `AllFiles`                     | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Append`                       | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Basename`                     | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Chmod`                        | Source-backed public surface. See the Go package for exact signature and behavior. |
| `CleanDirectory`               | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Close`                        | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Copy`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `CopyDirectory`                | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Delete`                       | Source-backed public surface. See the Go package for exact signature and behavior. |
| `DeleteDirectories`            | Source-backed public surface. See the Go package for exact signature and behavior. |
| `DeleteDirectory`              | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Directories`                  | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Dirname`                      | Source-backed public surface. See the Go package for exact signature and behavior. |
| `EnsureDirectoryExists`        | Source-backed public surface. See the Go package for exact signature and behavior. |
| `ExclusiveLock`                | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Exists`                       | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Extension`                    | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Files`                        | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Get`                          | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Glob`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `GuessExtension`               | Source-backed public surface. See the Go package for exact signature and behavior. |
| `HasSameHash`                  | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Hash`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `IsDirectory`                  | Source-backed public surface. See the Go package for exact signature and behavior. |
| `IsEmptyDirectory`             | Source-backed public surface. See the Go package for exact signature and behavior. |
| `IsFile`                       | Source-backed public surface. See the Go package for exact signature and behavior. |
| `IsReadable`                   | Source-backed public surface. See the Go package for exact signature and behavior. |
| `IsWritable`                   | Source-backed public surface. See the Go package for exact signature and behavior. |
| `JSON`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `LastModified`                 | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Lines`                        | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Link`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `MakeDirectory`                | Source-backed public surface. See the Go package for exact signature and behavior. |
| `MimeType`                     | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Missing`                      | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Move`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `MoveDirectory`                | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Name`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `New`                          | Source-backed public surface. See the Go package for exact signature and behavior. |
| `NewLockableFile`              | Source-backed public surface. See the Go package for exact signature and behavior. |
| `At`                           | Opens a directory as a Rooted sandbox.                                             |
| `Path`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Prepend`                      | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Put`                          | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Read`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `RelativeLink`                 | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Replace`                      | Source-backed public surface. See the Go package for exact signature and behavior. |
| `ReplaceInFile`                | Source-backed public surface. See the Go package for exact signature and behavior. |
| `SharedGet`                    | Source-backed public surface. See the Go package for exact signature and behavior. |
| `SharedLock`                   | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Size`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Truncate`                     | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Type`                         | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Unlock`                       | Source-backed public surface. See the Go package for exact signature and behavior. |
| `Write`                        | Source-backed public surface. See the Go package for exact signature and behavior. |

### Exported Errors, Variables, and Constants

| Name               | Notes                                                                              |
| ------------------ | ---------------------------------------------------------------------------------- |
| `ErrHashAlgorithm` | Source-backed public surface. See the Go package for exact signature and behavior. |
| `ErrLockFailed`    | Source-backed public surface. See the Go package for exact signature and behavior. |
| `ErrNotDirectory`  | Source-backed public surface. See the Go package for exact signature and behavior. |
| `ErrNotFound`      | Source-backed public surface. See the Go package for exact signature and behavior. |
