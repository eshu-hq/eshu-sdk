# Changelog

This file preserves the release history of the original in-tree
`github.com/eshu-hq/eshu/sdk/go/collector` module in Eshu. The `[0.1.0]` and `[0.2.0]` entries and
tags below belong to that original path and repository. The new
`github.com/eshu-hq/eshu-sdk/collector` path starts an independent version
line. Candidate entries describe source changes; tag publication and
production consumer support are separate states. Check Git tags and
the Go proxy for current availability. See the
[compatibility policy](../docs/compatibility.md).

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this module follows the versioning rule in
[Contract System v1 §5](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/internal/design/contract-system-v1.md#5-versioning-and-compatibility-policy):
a breaking change to the `collector-sdk/v1alpha1` wire protocol or the public
Go API is a major bump, an additive backward-compatible change is a minor
bump, and a docs-only or internal change is a patch for stable releases.
The preferred prerelease line remains `v0.1.0` until stable qualification;
its next collector candidate is `v0.1.0-rc.2`. See
[SDK Compatibility](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/public/extend/sdk-compatibility.md) for how
this module version lines up with core Eshu releases and the wire protocol.

Tags for the new module use the Go subdirectory-module format after separate release approval:
`collector/vX.Y.Z`.

## [Unreleased]

## [0.1.0-rc.2] - Unreleased

Preferred next candidate on the initial `v0.1.0` line. This source includes
the neutral `marker` API already present in the published
`collector/v0.2.0-rc.1` candidate. Review, merge, CI, separate tag approval,
and fresh proxy resolution remain before this tag can be called available.
Neither candidate establishes a qualified consumer/core pair.

## [0.2.0-rc.1] - 2026-10-24

Published historical candidate. The owner subsequently chose the `v0.1.0`
prerelease line for the preferred next candidate; this immutable tag remains
available but is not the preferred pin or a qualified consumer/core pair.

### Added

- `marker` provides deployment-keyed legacy-compatible HMAC-SHA256 markers
  for sensitive string and byte values, with normalized classification labels.

## [0.1.0-rc.1] - 2026-10-23

Candidate source prepared for separate review, merge, CI, and tag approval.
This entry belongs to the new module path, not the historical `v0.2.0` line.

### Added

- Public `Request`, structural validation, and its JSON Schema promote the
  existing host request shape without adding wire fields.
- Optional `TrustedDeletion` is an explicit collector attestation; core still
  decides whether deletion is admitted. It is part of this candidate source.
- `command.Run` provides explicit protocol and byte budgets, exact JSON-number
  configuration, deadline propagation, staged output and safe failure stages.
- The neutral `entrypoints` generator and `collector-entrypoints-gen` CLI create
  stdio commands around provider-owned handlers, with strict manifests and
  contained write/check behavior. Legacy hosted core templates remain separate.
- Current transport and conformance tooling are included. Optional trusted
  `conformance.Request.CoreFacts` binds a caller-evaluated authorization
  snapshot to manifest identity, reserved kinds, exact collector/schema tuples,
  reducer phases, and payload schemas. Nil context keeps source-only behavior
  and report bytes; JSON/YAML request decoding and the wire protocol are
  unchanged. Adding this exported struct field can break external unkeyed
  `Request` literals; keyed literals remain source compatible. The caller
  supplies authoritative grants, expiry/revocation, scope, version, phase,
  and schema evidence. This API does not itself qualify a collector.
- `ProtocolVersionV1Alpha2` (`collector-sdk/v1alpha2`): second collector
  extension wire contract, admitted only when explicitly configured;
  v1alpha1 default unchanged.

## [0.2.0] - 2026-09-16

Historical additive post-`v0.1.0` changes in the original module only. Verified against upstream tag
`sdk/go/collector/v0.1.0`: no `collector-sdk/v1alpha1` wire-protocol change
and no removed or renamed exported Go identifier, so this stays a minor bump
per the rule above. See
[SDK Compatibility](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/public/extend/sdk-compatibility.md) for the
version row for this release.

### Added

- `payloadSchemaRef` on conformance manifest fact families, letting a
  namespaced component fact declare the fixture-pack payload schema shape the
  host and CI should validate before publication or activation.
- `IsSensitiveKeyName` and `ValidateShareSafeKeys` — thin, behavior-preserving
  exports over the fail-closed sensitive-key-name predicate and recursive walk
  `validatePayloadKeys` already applies to `Fact.Payload`
  (`collector/validation.go`). They let a caller outside a collector
  (for example `go/internal/reportbundle`'s wrong-answer report bundle) reuse
  the exact same redaction/validation rule instead of duplicating or drifting
  from it.
- Payload-enum conformance coverage (`conformance/payload_enum_test.go`) plus
  conformance benches and a credential canary test, locking the
  fail-closed redaction and validation behavior external collectors rely on.
- Regression test locking the unsupported-protocol rejection:
  `ValidateResult` fails a result carrying an unknown, blank, or
  whitespace-only `protocol_version` with the error
  `protocol_version %q is unsupported`, naming the offending value (#6708).

### Changed

- Core Eshu now generates the host-side adapter from `collector.Fact` into the
  durable fact envelope. This does not change the `collector-sdk/v1alpha1` wire
  protocol or the public Go API.
- `ValidateResult` now refuses a `source_ref.uri` whose credential hides in an
  opaque URL body. `url.Parse` only surfaces userinfo after `//`, so
  `svc:SECRET@host/x` — the shape an operator produces by omitting `https://`
  — parsed with `User == nil` and passed "must not contain credentials".
  The validator re-parses the authority spelling when the opaque body has an
  `@` ahead of its first `/` and lets net/url decide; purls and other
  slash-first opaque bodies stay accepted. Validation errors for unparseable
  URIs no longer repeat the URI value, which can carry the credential into
  collector logs — and that promise now covers the error net/url NESTS inside
  its envelope too, where `invalid port ":secret" after host` quoted the exact
  bytes an operator leaves in port position (`svc:user@host:SECRET/x`). The
  nested reason is classified into fixed text instead of wrapped verbatim,
  failing closed to a generic reason for message shapes the classifier does
  not know. No wire-protocol or Go API change.
- Go directive `1.26.0` → `1.26.5` (toolchain floor only; no language or
  library break for consumers).

## [0.1.0] - 2026-07-06

First tagged release of the original in-tree module (`sdk/go/collector/v0.1.0` at `92061fe39`, fetchable via
the Go module proxy). Content below is the module tree at that tag.

### Added

- `Claim`, `Scope`, and `Generation` types describing a core-owned work item
  handed to an out-of-tree collector.
- `Fact`, `SourceRef`, and `Redaction` types for source evidence records.
- `Status` and `Result` types for `complete`, `unchanged`, `partial`,
  `retryable`, and `terminal` collector outcomes.
- `Contract`, `FactDeclaration`, `Validator`, and `ValidationReport` for
  fail-closed host-side validation via `NewValidator(contract).ValidateResult`.
- `collector-sdk/v1alpha1` wire protocol, published as
  `schema/collector-sdk-v1alpha1.schema.json`.
- `conformance` subpackage (`collector/conformance`) — the public,
  importable conformance harness (`conformance.Run`) an out-of-tree collector
  runs in its own CI, including payload-shape validation against
  `factschema` schemas via `conformance.Request.PayloadSchemas`.
- `schema/cassette-format.v1.schema.json` — a generated mirror of the host's
  replay cassette envelope contract, so credential-free replay fixtures can be
  validated offline against the same schema the host enforces.

## Convention for future entries

Add a new `## [Unreleased]` entry at the top of this file for every
merge-worthy user-facing change to this module, grouped under `### Added`,
`### Changed`, `### Deprecated`, `### Removed`, `### Fixed`, or `### Security`
as needed (omit empty groups). When a maintainer cuts a release:

1. Rename the top `[Unreleased]` section to `[X.Y.Z] - YYYY-MM-DD` using the
   version chosen per the breaking/additive/patch rule above.
2. Add a fresh empty `## [Unreleased]` section above it.
3. Follow the [release and cutover checklist](../docs/release.md) before
   tagging a new module release with `collector/vX.Y.Z`.

A wire-protocol-breaking change additionally needs a new protocol version
(for example `collector-sdk/v1`) per Contract System v1 §5; do not reuse
`v1alpha1` for an incompatible wire shape.
