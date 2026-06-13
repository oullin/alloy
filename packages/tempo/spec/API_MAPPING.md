# Carbon to Tempo API Mapping

Tempo keeps Carbon `v3.11.4` as the fixture oracle while exposing idiomatic Go and TypeScript APIs.

## TypeScript

- `Carbon::parse(...)` maps to `Tempo.parse(...)`.
- `CarbonImmutable::parse(...)` maps to `TempoImmutable.parse(...)`.
- Mutating Carbon methods map to chainable methods on `Tempo`.
- Immutable Carbon methods map to methods returning a new `TempoImmutable`.

## Go

- Constructors return `(Tempo, error)` when input can fail.
- Mutation is represented by returned values instead of hidden object mutation.
- PHP magic accessors are represented as explicit methods.

Unsupported PHP-only constructs must be documented here before implementation is considered complete.
