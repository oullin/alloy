# Getting Started

Alloy is a collection of foundational Go packages for building web applications.
The Go packages are published from the `hara.sh/alloy` module, so you
import only the packages you need.

## Requirements

| Tool    | Version |
| ------- | ------- |
| Go      | ≥ 1.26  |
| Node.js | ≥ 24    |
| pnpm    | ≥ 10.33 |

## Installation

Alloy is developed and consumed privately. The `hara.sh/alloy` module
path is not served publicly, so the Go packages are consumed through a Go
workspace instead of `go get`:

```bash
git clone git@github.com:oullin/alloy.git

# In your application repository, next to the alloy checkout:
go work init .
go work use ../alloy/pkg/hub
```

Then import the packages you need as usual:

```go
import (
	"hara.sh/alloy/cache"
	"hara.sh/alloy/container"
)
```

## Project Layout

```
pkg/hub/              Go library packages
sdk/                  TypeScript packages
web/                  Documentation, web demos, and runtime storage
```

## Development Setup

Clone the repository, install Node dependencies, and run the test suite:

```bash
git clone https://github.com/oullin/alloy.git
cd alloy
pnpm install
pnpm run test
```

Run the documentation site locally:

```bash
pnpm run dev --filter=@alloy/docs
```

Build the static site:

```bash
pnpm run build --filter=@alloy/docs
# Output: web/docs/.vuepress/dist/
```

## Package Index

The packages below ship today from `hara.sh/alloy`.

### Architecture

| Package                          | Purpose                                           |
| -------------------------------- | ------------------------------------------------- |
| [container](/packages/container) | IoC service container and application bootstrap   |
| [config](/packages/config)       | Configuration repository with dot-notation access |
| [contracts](/packages/contracts) | Shared interface definitions for every package    |

### The Basics

| Package                            | Purpose                                                  |
| ---------------------------------- | -------------------------------------------------------- |
| [httpx](/packages/httpx)           | HTTP utilities, routing, middleware, and testing helpers |
| [session](/packages/session)       | Session management with multiple storage handlers        |
| [cookie](/packages/cookie)         | HTTP cookie handling                                     |
| [validation](/packages/validation) | Rule-based input validation (80+ built-in rules)         |
| [inertia](/packages/inertia)       | Server-side Inertia.js protocol adapter                  |

### Security

| Package                            | Purpose                                            |
| ---------------------------------- | -------------------------------------------------- |
| [auth](/packages/auth)             | Authentication, authorization, password management |
| [encryption](/packages/encryption) | AES encryption with CBC and GCM mode support       |
| [hashing](/packages/hashing)       | Password hashing with bcrypt and Argon2            |

### Data & Storage

| Package                            | Purpose                                      |
| ---------------------------------- | -------------------------------------------- |
| [cache](/packages/cache)           | Caching layer with multiple driver support   |
| [database](/packages/database)     | Shared database errors and support utilities |
| [filesystem](/packages/filesystem) | Local filesystem operations                  |

### Events & Jobs

| Package                    | Purpose                                           |
| -------------------------- | ------------------------------------------------- |
| [events](/packages/events) | Event dispatching and listener management         |
| [bus](/packages/bus)       | Command and event bus with pipeline support       |
| [queue](/packages/queue)   | Background job processing with pluggable drivers  |
| workflow                   | Petri-net based workflow and state-machine engine |

### Support & Utilities

| Package                            | Purpose                                           |
| ---------------------------------- | ------------------------------------------------- |
| [collection](/packages/collection) | Fluent collection helpers for slices and maps     |
| [str](/packages/str)               | String helpers, UUID/ULID generation              |
| tempo                              | Date/time library with timezones and localization |
| [money](/packages/money)           | Money and currency primitives                     |
| [seo](/packages/seo)               | SEO utilities and i18n locale handling            |

## Roadmap

The following packages are documented as design targets but are **not yet
available** in `hara.sh/alloy`:

[routing](/packages/routing), [redis](/packages/redis),
[pagination](/packages/pagination), [jobqueue](/packages/jobqueue),
[pipeline](/packages/pipeline), [mailx](/packages/mailx),
[notifications](/packages/notifications), [ai/sdk](/packages/ai/sdk),
[ai/mcp](/packages/ai/mcp), [ai/boost](/packages/ai/boost),
[support](/packages/support), [log](/packages/log),
[translation](/packages/translation), [concurrency](/packages/concurrency),
[conditionable](/packages/conditionable), [jsonx](/packages/jsonx),
[prompts](/packages/prompts), [logtail](/packages/logtail),
[remotetasks](/packages/remotetasks), [debugbar](/packages/debugbar),
[inception](/packages/inception), [authkit](/packages/authkit),
[billing](/packages/billing), [authflows](/packages/authflows)

## AI Assisted Development

Alloy is built from the ground up to be AI-friendly. By providing clear
interfaces, predictable patterns, and explicit dependencies, Alloy makes it
easy for AI agents to understand and contribute to your codebase.

## Concept Guides

Cross-cutting topics that span multiple packages:

| Guide                                        | What it covers                                                    |
| -------------------------------------------- | ----------------------------------------------------------------- |
| [Request Lifecycle](/architecture/lifecycle) | How an HTTP request flows through a Alloy application             |
| [Cross-Runtime Parity](/architecture/parity) | Which primitives have Go/TS twins and what parity each guarantees |
| [Testing](/concepts/testing)                 | Built-in test doubles and testing patterns                        |
| [Middleware](/basics/middleware)             | Global, per-route, and controller-scoped middleware               |
| [Controllers](/basics/controllers)           | Grouping handlers into types with shared middleware               |
| [URL Generation](/basics/url-generation)     | Named-route URLs, signed URLs, redirects                          |
| [CSRF Protection](/basics/csrf)              | Tokens, headers, excluding routes                                 |

## Next Steps

Once your app starts up, you'll want to know how it's actually wired
together. The Architecture Concepts guides walk through that:

| Guide                                                       | What it answers                                              |
| ----------------------------------------------------------- | ------------------------------------------------------------ |
| [Directory Structure](/getting-started/directory-structure) | Where to put your code (entry point, routes, models, config) |
| [Application Bootstrap](/architecture/application)          | How `NewApplication` + `StandardProviders` brings up the app |
| [Service Container](/architecture/service-container)        | How to bind services and resolve them in handlers            |
| [Service Providers](/architecture/service-providers)        | How to write your own provider                               |
| [Drivers](/architecture/drivers)                            | How to swap cache/queue/log backends and add custom ones     |
| [Facades](/architecture/facades)                            | The shortcut layer over the container                        |
| [Configuration](/architecture/configuration)                | How options flow into the provider stack                     |
