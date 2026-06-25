# Alloy Auth Go

Alloy Auth provides headless Laravel Fortify / Jetstream style backend security
primitives for Go applications.

The package is intentionally UI-free. Frontends talk to JSON HTTP handlers,
store WebAuthn ceremony state server-side, and render their own screens.

See:

- [Headless security contracts](docs/headless-security.md)
- [Storage contracts](docs/storage-contracts.md)

Production apps should validate `security.ProductionDefaults()` after setting
deployment-specific app key and WebAuthn relying-party values.

Use `github.com/oullin/alloy/cache` when an auth or routing component needs a
shared TTL cache or rate-limit backend.
