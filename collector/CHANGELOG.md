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
Core Eshu pins the `v0.2.0` prerelease line (`v0.2.0-rc.2`), so new
candidates continue there; the latest published `v0.1.0` candidate is
`v0.1.0-rc.5`. See
[SDK Compatibility](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/public/extend/sdk-compatibility.md) for how
this module version lines up with core Eshu releases and the wire protocol.

Tags for the new module use the Go subdirectory-module format after separate release approval:
`collector/vX.Y.Z`.

## [Unreleased]

## [0.2.0-rc.3] - Prepared (unpublished)

This candidate source is not tagged. Publishing it needs a separate release
ruling against the merged commit, and no proxy verification is claimed for it.
It continues the `v0.2.0` prerelease line because core Eshu pins
`collector/v0.2.0-rc.2`; a `v0.1.0-rc.N` tag would sort below that pin. The
`Claim.ScopeQueries` entry first shipped in `collector/v0.2.0-rc.2`, whose tag
points at `a663423`, a commit that was never merged to `main`; this source
carries that change as a cherry-pick. The Fixed and Changed entries first
shipped in `collector/v0.1.0-rc.5` and `collector/v0.2.0-rc.2`. The new item is
the additive `collector-sdk/v1alpha3` pagination contract, so the module bump
is a prerelease step within the 0.2 line (rc.2 to rc.3). No required field,
existing field, or v1alpha1/v1alpha2 wire format changes. Schema file digests
(SHA-256) at this source:
`schema/collector-sdk-v1alpha3.schema.json` `c8a6a10926bfc3cf99c1c0f7d6c857cfd3e1637feed364154c06ebc8059124a0`,
`schema/collector-request-v1alpha3.schema.json` `68dc5c58beb1da8f80782301838c29a80d0f58def5928e2b18e5f4768e454490`,
and `schema/collector-sdk-v1alpha1.schema.json`, unchanged since `v0.2.0-rc.2`,
`c665efe4b9021fe25791822a5990c19460ac8189f8707733513d579c8d5bb2b1`. The release
workflow's `expect_file` pair is
`collector/schema/collector-sdk-v1alpha3.schema.json` with the first digest.

### Breaking (Go API)

- `Claim` is no longer comparable, because `Claim.ScopeQueries` is a slice.
  `a.Claim != b.Claim` and `a.Claim == b.Claim` stop compiling. This first
  shipped in `collector/v0.2.0-rc.2` and cannot change there; this candidate
  adds no further break. Migrate by comparing field by field and using
  `slices.Equal` on `ScopeQueries`, or use `reflect.DeepEqual`. The
  `collector-sdk/v1alpha1` schema artifacts were also changed in place under
  the same `$id` in that release.

### Added

- `Claim.ScopeQueries` carries the per-scope query list (`ScopeQuery`
  with `ecosystem`, `package_name`, `version`; at most `MaxScopeQueries`
  entries) for scope ids that name a shard rather than its contents
  (pinned virtual-slot scopes, eshu-hq/eshu#9432). Empty on every legacy
  scope id and omitted from the wire then, so old satellites read
  byte-identical claims. The wire change is additive: the
  `collector-sdk/v1alpha1` required sets are unchanged.
  `validateClaim` now rejects more than `MaxScopeQueries` entries and any entry
  whose `package_name` or `version` is empty or only whitespace. The schemas
  already require non-empty `package_name` and `version` but carry no cap on
  the entry count, so the cap is a Go-side check the host also applies. The Go change is not fully additive:
  the slice field makes `Claim` incomparable, so `a.Claim != b.Claim`
  no longer compiles. Compare with `reflect.DeepEqual` or field by field.
  Several satellite test suites use `==` on `Claim` and need that edit
  when they move to `v0.2.0-rc.2` or later.
- `ProtocolVersionV1Alpha3` (`collector-sdk/v1alpha3`): third collector
  extension wire contract with versioned multi-call pagination
  (eshu-hq/eshu#9385), ported from the Eshu host's in-tree mirror
  (eshu-hq/eshu#9395, eshu-hq/eshu#9635). Hosts admit it only when explicitly
  configured; the v1alpha1 default is unchanged. `Result.Page`
  (`ResultPage`) carries the resume `cursor` (required while `more` is true,
  empty on the final page), the zero-based `page_index`, and the optional
  declared `total_facts`; only `complete` and `partial` results may request
  continuation. `Request.Continuation` (`PageContinuation`) carries the
  cursor back on resumed calls and is admitted only on v1alpha3 with a
  non-blank cursor. `PageBoundForStdout`, `FactWireBytes`, and `SplitFacts`
  size every page against the host stdout limit minus 64 KiB of envelope
  headroom; a single fact above the bound fails closed naming its kind and
  stable key, never dropped. `command.Run`, `entrypoints` manifests, and the
  conformance manifest admit the new protocol when listed. New artifacts:
  `schema/collector-sdk-v1alpha3.schema.json` (`JSONSchemaV1Alpha3`) and
  `schema/collector-request-v1alpha3.schema.json`
  (`RequestJSONSchemaV1Alpha3`). Compatibility: this change leaves the v1alpha1 result and
  request schema artifacts byte-identical to `v0.2.0-rc.2`, pageless results and
  first-call requests serialize unchanged, and a strict decoder of the
  older shapes rejects a page or continuation member, so older hosts and
  collectors fail closed instead of dropping the continuation. Adding the
  exported `Result.Page` and `Request.Continuation` fields can break external
  unkeyed `Result` or `Request` literals; keyed literals stay source
  compatible.

### Fixed

- Payload validation admits the redaction-safe Vault auth-role metadata keys
  `token_ttl_seconds` (integer TTL duration), `token_policy_count` (integer
  policy count), and `token_policy_name_fingerprints` (HMAC redaction
  markers, never raw names) on `vault_auth_role` facts
  (eshu-hq/eshu#8869): the substring heuristic rejected the first key and
  the fail-closed walk would reject the other two, ending collection with
  `invalid_result` and zero emitted facts.

### Changed

- Go directive `1.26.5` → `1.26.10` (toolchain floor only; clears the 11
  standard-library advisories govulncheck reports as reachable from this
  module at the old floor — `GO-2026-6809`, `GO-2026-6617`, `GO-2026-6613`,
  `GO-2026-6612`, `GO-2026-6611`, `GO-2026-6610`, `GO-2026-6608`,
  `GO-2026-6607`, `GO-2026-6605`, `GO-2026-6604`, `GO-2026-6603` —
  with no language or library break for consumers).

## [0.1.0-rc.4] - 2026-10-27

This tag is published and proxy-verified; it does not qualify a core consumer
or deployed pair. It carries the RC3 source below plus the `env_from_secret`
allowlist fix.

### Fixed

- Payload validation admits the redaction-safe `env_from_secret` boolean on
  `kubernetes_live.pod_template` containers (issue #26, PR #28): the flag
  records only whether a container references secret-backed env and never
  carries a value. Non-boolean values under that key are still rejected, and
  the pod template valid fixture now carries the key.

## [0.1.0-rc.3] - 2026-10-25

This tag is published and proxy-verified; it does not qualify a core consumer
or deployed pair. This candidate source includes the strict native JSON
duplicate-member guard and the Azure tag and Kubernetes posture validation
corrections below. The guard does not freeze the caller's payload; custom
serialization must remain stable through subsequent use.

### Fixed

- Native fact payloads with custom JSON serialization reject duplicate decoded
  member names at every object depth before later values can overwrite earlier
  raw values. Validation does not mutate or freeze the caller's result; callers
  must keep serialization stable through subsequent use.
- Fact validation admits Kubernetes posture fields only at schema `1.0.0`
  roots: `k8s_service_account` and `k8s_service_account_token_posture` allow
  `secret_ref_count` and `image_pull_secret_ref_count` as JSON numbers or null,
  and `automount_token` as `"true"`, `"false"`, `"unknown"`, or null.
  `k8s_workload_identity_use` allows `projected_service_account_token` only as
  a boolean. Generic sensitive-key checks remain strict; payload schemas
  separately own integer validity.
- Fact validation admits original tag names only in the root fingerprint map
  of `azure_tag_observation` schema `1.0.0`, and requires every value to be a
  complete HMAC-SHA256 redaction marker. Native and JSON-decoded payloads
  reject credential-shaped keys everywhere else.

## [0.1.0-rc.2] - 2026-10-24

Preferred collector candidate on the initial `v0.1.0` line. This source includes
the neutral `marker` API already present in the published
`collector/v0.2.0-rc.1` candidate.

The tag is published. Verify its public-proxy origin before use; publication
alone does not qualify a consumer/core pair.

## [0.2.0-rc.1] - 2026-10-24

Published historical candidate. The owner subsequently chose the `v0.1.0`
prerelease line for the preferred next candidate; this immutable tag remains
published but is not the preferred pin or a qualified consumer/core pair.

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
