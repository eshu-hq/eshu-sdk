# Collector SDK

This module defines the public Go contracts for out-of-tree Eshu collector
extensions. It intentionally does not import `github.com/eshu-hq/eshu/go/internal`
packages. Core Eshu hosts validate these records before mapping them into the
internal durable fact envelope.

## Compatibility

- Go module path: `github.com/eshu-hq/eshu-sdk/collector`
- Release version: not yet selected for this relocated module
- Initial wire protocol: `collector-sdk/v1alpha1`
- JSON Schema artifact: `schema/collector-sdk-v1alpha1.schema.json`
- Cassette format JSON Schema (replay fixtures): `schema/cassette-format.v1.schema.json`
  — a generated mirror of the host's cassette envelope contract
  (`go/internal/replay/schema`), so credential-free replay cassettes can be
  validated offline against the same schema the host enforces.

The module version may add helpers without changing the wire protocol. The wire
protocol changes only when host and extension records would serialize
differently.

## Contracts

The public surface includes:

- `Claim`, `Scope`, and `Generation` for core-owned work identity.
- `Fact`, `SourceRef`, and `Redaction` for source evidence.
- `Status` and `Result` for complete, unchanged, partial, retryable, and
  terminal outcomes.
- `Contract`, `FactDeclaration`, `Validator`, and `ValidationReport` for
  fail-closed host validation.
- `Request` and `RequestJSONSchema` for the core-owned command input.

`Request.Validate()` checks known and matching request/contract protocols and
the existing structural claim contract without consulting the clock. Opaque
`Config` numbers decoded by the command runner are `json.Number`; handlers
should convert them explicitly. The checked-in request schema and fixture are
under `schema/` and `testdata/fixtures/`. See [command boundary](../docs/command.md)
for byte budgets, deadline behavior, and host responsibilities.

Collectors emit facts only. They do not write graph truth, mutate queues, run
DDL, or import Eshu internals.

`marker/` is a standard-library-only helper for deterministic keyed redaction
markers. Its zero `Key` preserves legacy behavior but production collectors
should construct a nonzero key before acquiring source data. See
[redaction markers](marker/README.md).

`transport/` is a separate, standard-library-only helper package for bounded
HTTP requests, TLS trust, retry-status decisions, and safe provider failures.
Its inherited `Version = "v0.1.0"` is an internal contract marker, not the
version of this unreleased collector module. Core collector call sites have not
migrated to this package.

## Validation

`NewValidator(contract).ValidateResult(result)` rejects:

- unsupported protocol versions or result states;
- undeclared fact kinds, schema versions, source confidence values, and
  tombstones;
- `source_confidence=unknown`;
- blank or control-character stable keys;
- source references outside the claimed scope/generation/fact key;
- credential-bearing or host-local source URIs;
- payload keys that look like credentials and were not redacted before emission;
- conflicting duplicates for the same fact kind and stable key.

When a component manifest declares `payloadSchemaRef`, the in-tree host and
`eshu component conform` additionally validate that fact's payload against the
referenced `factschema/fixturepack` schema before publication or hosted
activation. Namespaced component facts without a schema ref remain
provenance-only and are not payload-shape validated.

Exact duplicate facts are accepted as idempotent and reported in
`ValidationReport.DuplicateCount`. Conflicting duplicates fail before host
commit.

## Fixtures

`testdata/fixtures` contains golden result fixtures for complete, unchanged,
partial, retryable, terminal, duplicate, conflict, tombstone, and redaction
cases. The package tests validate those fixtures against the public contract and
assert the generated JSON Schema stays in lockstep with the checked-in artifact.

The SDK emits no telemetry. The command runner and generated commands execute
on the collector runtime path; providers and hosts own operational metrics,
traces, and logs. Safe command failure stages are described in
[the command guide](../docs/command.md).
