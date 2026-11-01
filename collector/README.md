# Collector SDK

This module defines the public Go contracts for out-of-tree Eshu collector
extensions. It intentionally does not import `github.com/eshu-hq/eshu/go/internal`
packages. Core Eshu hosts validate these records before mapping them into the
internal durable fact envelope.

## Compatibility

- Go module path: `github.com/eshu-hq/eshu-sdk/collector`
- Release version: not yet selected for this relocated module
- Initial wire protocol: `collector-sdk/v1alpha1`
- Paged wire protocol: `collector-sdk/v1alpha3` (optional `Result.Page`
  continuation and `Request.Continuation` resume cursor; v1alpha1 and v1alpha2
  single-shot records are unchanged)
- JSON Schema artifact: `schema/collector-sdk-v1alpha1.schema.json`
- JSON Schema artifacts (paged): `schema/collector-sdk-v1alpha3.schema.json`
  (result) and `schema/collector-request-v1alpha3.schema.json` (request)
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
- `ResultPage`, `PageContinuation`, and the byte-accounting helpers
  `PageBoundForStdout`, `FactWireBytes`, and `SplitFacts` for versioned
  (`collector-sdk/v1alpha3`) multi-call pagination of results that would not
  fit one bounded stdout write.
- `Contract`, `FactDeclaration`, `Validator`, and `ValidationReport` for
  fail-closed host validation.
- `Request` and `RequestJSONSchema` for the core-owned command input.

`Request.Validate()` checks known and matching request/contract protocols and
the existing structural claim contract without consulting the clock. Opaque
`Config` numbers decoded by the command runner are `json.Number`; handlers
should convert them explicitly. The checked-in request schema and fixture are
under `schema/` and `testdata/fixtures/`. See [command boundary](../docs/command.md)
for byte budgets, deadline behavior, and host responsibilities.

## Pagination (`collector-sdk/v1alpha3`)

A host admits `collector-sdk/v1alpha3` only when it is explicitly configured.
A collector that would exceed the host stdout limit splits its facts with
`SplitFacts(facts, bound)`, where `bound` comes from
`PageBoundForStdout(stdoutLimit)` (the limit minus 64 KiB of envelope
headroom). Each page repeats the claim and generation and sets `Result.Page`:
`more` and a non-blank `cursor` while pages remain, then a final page with
`more` false and no cursor. Only `complete` and `partial` results may request
continuation. `page_index` counts from zero, and `total_facts`, when declared,
must agree on every page. The host re-invokes the collector with
`Request.Continuation.Cursor` set to the previous cursor. The cursor format
belongs to the collector. `SplitFacts` never drops a fact. One fact above the
bound fails, naming its kind and stable key.

`ValidateResult` rejects a page on v1alpha1 or v1alpha2, a continuation
without a cursor, a final page with a cursor, and negative counters. Rules that
span pages (no gaps or replays, a stable claim and generation, the declared
total) need every page and belong to the reassembling host. A collector built
on an older SDK decodes requests strictly, so it rejects a resumed call
instead of restarting at page zero.

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
- duplicate decoded JSON member names in a native payload's canonical
  representation, including identical duplicates and escaped equivalent names;
- conflicting duplicates for the same fact kind and stable key.

Native payloads with custom JSON marshaling are checked against their serialized
JSON, including member-name uniqueness at every object depth. The caller must
provide deterministic serialization and keep the result stable from validation
through subsequent host use. `ValidationReport` does not freeze or copy the
payload; concurrent mutation or stateful marshalers can change what is later
emitted.

For `azure_tag_observation` schema `1.0.0` only, the root
`tag_value_fingerprints` object may retain source tag names, including names
such as `api_key`. Every value must be a `redacted:hmac-sha256:` marker with
exactly 64 lowercase hexadecimal digits. Null and malformed values fail
validation. This exception does not apply to nested objects, other fact kinds,
other schema versions, or the context-free `ValidateShareSafeKeys` check.
Native Go payloads and JSON-decoded payloads receive the same validation.

For Kubernetes posture facts at schema `1.0.0`, the validator admits only these
exact root fields through the sensitive-key check:

| Fact kind | Root field | Key-safe JSON value |
| --- | --- | --- |
| `k8s_service_account`, `k8s_service_account_token_posture` | `secret_ref_count`, `image_pull_secret_ref_count` | Number or null |
| Same two kinds | `automount_token` | `"true"`, `"false"`, `"unknown"`, or null |
| `k8s_workload_identity_use` | `projected_service_account_token` | Boolean |

These fields remain optional. This is key-safety admission, not payload-schema
validation: signed, unsigned, fractional, negative, exponent, and large JSON
numbers pass this check after successful JSON serialization. A configured
payload schema separately decides whether a count is an integer. The exception
does not apply to other kinds, schema versions, nested keys, or the context-free
`IsSensitiveKeyName` and `ValidateShareSafeKeys` helpers. Native aliases,
pointers, typed nil values, and custom marshalers are checked against their
canonical JSON representation.

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
cases. `paged_first.json` and `paged_final.json` are one two-page v1alpha3
result, and `request_continuation.json` is a resumed v1alpha3 request. The package tests validate those fixtures against the public contract and
assert the generated JSON Schema stays in lockstep with the checked-in artifact.

The SDK emits no telemetry. The command runner and generated commands execute
on the collector runtime path; providers and hosts own operational metrics,
traces, and logs. Safe command failure stages are described in
[the command guide](../docs/command.md).
