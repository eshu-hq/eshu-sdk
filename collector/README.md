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
- `ValidateFactPayload` for producers that have a complete `Fact` payload but
  no host claim yet. It applies the same fact-kind and schema-version scoped
  payload safety rules used by `Validator.ValidateResult`; it does not validate
  payload presence, envelope fields, contract declarations, or source scope.
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
- duplicate decoded JSON member names in a native payload's canonical
  representation, including identical duplicates and escaped equivalent names;
- conflicting duplicates for the same fact kind and stable key.

Native payloads with custom JSON marshaling are checked against their serialized
JSON, including member-name uniqueness at every object depth. The caller must
provide deterministic serialization and keep the result stable from validation
through subsequent host use. `ValidationReport` does not freeze or copy the
payload; concurrent mutation or stateful marshalers can change what is later
emitted. If marshaling fails, validation reports the failure stage without
forwarding the producer's error text, which may contain a credential.

For `azure_tag_observation` schema `1.0.0` only, the root
`tag_value_fingerprints` object may retain source tag names, including names
such as `api_key`. Every value must be a `redacted:hmac-sha256:` marker with
exactly 64 lowercase hexadecimal digits. Null and malformed values fail
validation. This exception does not apply to nested objects, other fact kinds,
other schema versions, or the context-free `ValidateShareSafeKeys` check.
Native Go payloads and JSON-decoded payloads receive the same validation.

For `aws_resource` schema `1.0.0`, only `aws_lambda_function` with
`service_kind=lambda` and `aws_ecs_task_definition` with `service_kind=ecs`
(when the service kind is present) admit the legacy AWS environment shapes.
Lambda `attributes.environment` is an object keyed by variable name; ECS
`attributes.containers[].environment` is an array of `{name,value}` objects.
Each value must be an object with exactly `marker`, `reason`, `source`, and
`ruleset_version`: a lowercase 64-hex HMAC-SHA256 marker, the legacy
`known_sensitive_key` or `unknown_provider_schema` reason, the source path
derived from its variable name, and `aws-launch-2026-05-14`. ECS
`attributes.containers[].secrets` admits only an array of `{name,value_from}`
string pairs. The legacy absent, null, and empty states remain valid. Other
fields still receive recursive sensitive-key validation. A marker-shaped
object proves its JSON shape, not that a producer used a particular key or
raw input. Native maps reject malformed UTF-8 member names, which the Go JSON
encoder would otherwise replace and potentially turn into duplicate names.
Raw JSON and custom marshalers are decoded with duplicate-member detection
before validating their serialized keys. The context-free
`ValidateShareSafeKeys` policy is unchanged.

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
cases. The package tests validate those fixtures against the public contract and
assert the generated JSON Schema stays in lockstep with the checked-in artifact.

The SDK emits no telemetry. The command runner and generated commands execute
on the collector runtime path; providers and hosts own operational metrics,
traces, and logs. Safe command failure stages are described in
[the command guide](../docs/command.md).
