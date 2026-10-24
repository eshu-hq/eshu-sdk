# Changelog

This file preserves the release history of the original in-tree
`github.com/eshu-hq/eshu/sdk/go/factschema` module in Eshu. The `[0.1.0]`
and `[0.2.0]` entries and tags below belong to that original path and
repository. The new `github.com/eshu-hq/eshu-sdk/factschema` path starts an
independent version line. Candidate entries describe source changes; tag
publication and production consumer support are separate states. Check Git tags and
the Go proxy for current availability. See the
[compatibility policy](../docs/compatibility.md).

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Per-fact-kind payload schemas follow the breaking-change policy in
[Contract System v1 §5](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/internal/design/contract-system-v1.md#5-versioning-and-compatibility-policy):
removing or renaming a field, narrowing a field's type, or changing a stable
key's derivation is a major bump with a decode shim in the same change; an
additive optional field is a minor bump; docs-only is a patch. The **module**
version tracks the highest bump any fact kind needed in that release. See
[SDK Compatibility](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/public/extend/sdk-compatibility.md) for how
this module version lines up with core Eshu releases and fixture-pack pins.

Tags for the new module use the Go subdirectory-module format after separate release approval:
`factschema/vX.Y.Z`. The fixture pack
(`factschema/fixturepack`) ships inside this module and has no separate
version — pinning this module pins the fixture pack too.

## [Unreleased]

## [0.1.0-rc.2] - 2026-10-24

Published candidate from the reviewed source commit also used by collector
`v0.2.0-rc.1`. Fresh public-proxy resolution and a typed-decoder smoke check are
recorded in [candidate evidence](../docs/evidence/factschema-v0.1.0-rc.2.md). Publication does not qualify a consumer/core pair.

### Fixed

- Nested integer slices accept exact integral `json.Number` decimal and
  exponent values, retain legacy mixed-member conversion and report bounded
  invalid-input errors.

- Typed payload decoding now accepts number-preserving `json.Number` input.
  Integer fields preserve integral decimal and exponent values exactly and
  reject fractional or out-of-range values. Float fields accept finite values,
  including fractions, with float64 rounding. Malformed tokens and float64
  overflow fail as `input_invalid`. Schemas and fixture payloads are unchanged.

## [0.1.0-rc.1] - 2026-10-23

Candidate source prepared for separate review, merge, CI, and tag approval.
This entry belongs to the new module path, not the historical `v0.2.0` line.

### Added

- `identity` provides stable fact identity derivation, and `contracttest`
  provides generic contract assertions for out-of-tree consumers. Their core
  adoption remains a separately validated step.
- New `ref_overlay/v1` package with `Decode`/`Encode` seams, schemas, and
  fixture-pack entries for epic #5393 Phase 1 (contract only — no collector,
  reducer, or query consumer yet): `ref_overlay.branch_observation` (one
  deployment-referenced non-default ref with its resolution state) and
  `config_key_divergence` (one per-environment config-key divergence row
  carrying keys plus a divergence boolean only — no values, no hashes).
- `codegraph/v1` `Import` gains three optional booleans, `type_only`, `deferred`
  and `inferred` (json tags of the same names, `omitempty`). A parser sets one
  only when it is true. An additive optional field is a minor change.
- Compared with the supported old `factschema@v0.2.0`, this source contains
  three added schemas, no removed schemas, and four existing schemas with
  optional nullable fields added. The count changes from 176 to 179; it does
  not establish acceptance by an old core consumer.

## [0.2.0] - 2026-09-16

Historical additive post-`v0.1.0` changes in the original module — 46 new schemas, zero breaking schema
diffs against tag `factschema/v0.1.0`
(`bash scripts/verify-factschema-diff.sh -base-ref factschema/v0.1.0`
exits 0), so this stays a minor bump per the policy above. See
[SDK Compatibility](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/public/extend/sdk-compatibility.md) for the
version row for this release.

### Added

- Six new typed `<family>/v1` packages with `Decode<Kind>`/`Encode<Kind>`
  seams, schemas, and fixture-pack entries: `codeowners`
  (`codeowners.ownership`), `reducerderived` (governed reducer-owned findings
  such as `reducer_supply_chain_impact_finding`,
  `reducer_aws_cloud_runtime_drift_finding`,
  `reducer_multi_cloud_runtime_drift_finding`, plus package-correlation and
  terraform-drift findings), `scannerworker` (`scanner_worker.analysis`,
  `scanner_worker.warning`), `semantic` (`semantic.code_hint`,
  `semantic.documentation_observation`), `submodule` (`submodule.pin`), and
  `vulnerabilitysuppression` (`vulnerability.suppression`, shared by VEX,
  operator-policy, and provider-dismissal producers).
- `reducer_supply_chain_impact_finding` carries the optional
  `environment_evidence` field (`map[string]string`, `omitempty`), labelling
  each name in `environments` as `deploy_event` or `declared` (issue #5426),
  plus the optional `ci_declared_artifact_digest` and `ci_declared_image_ref`
  fields holding the matched `cicd_run_correlation` deployment's own declared
  artifact identity, baked only on a strong-branch match (issue #5469). All
  three are additive-optional: a finding written before them existed still
  decodes with nil fields.
- New kinds in existing families: `vulnerability.reference` and
  `vulnerability.source_snapshot`; `aws` IAM boundary/policy/attachment/trust,
  DNS, image-reference, and warning kinds; `azure` identity-observation,
  image-reference, resource-change, and tag-observation kinds; `gcp` IAM and
  image-reference kinds; `ci.deployment_event`; Kubernetes RBAC and
  service-account-token kinds; `secrets_iam_coverage_warning`; Vault
  auth-mount, identity, and secret-engine-mount kinds; and `workitem` metadata
  extensions.
- `SchemaBytes` — embedded access to the checked-in `schema/*.json` bytes for
  one fact kind, so out-of-module conformance tests can load a committed
  schema without duplicating the schema tree the way `fixturepack` must.

### Changed

- Core Eshu now generates the adapter that maps durable internal fact envelopes
  into this module's `Envelope` for Decode calls. This does not change any
  factschema payload schema, fixture-pack artifact, or public Go API.
- Go directive `1.26.0` → `1.26.5` (toolchain floor only; no language or
  library break for consumers).

## [0.1.0] - 2026-07-06

First tagged release of the original in-tree module (`factschema/v0.1.0` at `92061fe39`, fetchable via
the Go module proxy; closes the scaffold started in #4567). Content below is
the module tree at that tag: 130 generated schemas, zero breaking diffs since
only the tag itself is the baseline.

### Added

- `Envelope` — the canonical fact envelope shared by every typed fact kind.
- Twenty typed `<family>/v1` packages, each with a `Decode<Kind>`/`Encode<Kind>`
  seam: `aws`, `azure`, `gcp`, `iam`, `secretsiam`, `incident` (the first family
  with dotted wire kinds, e.g. `incident_routing.applied_pagerduty_resource`),
  `cicdrun`, `codedataflow`, `codegraph`, `documentation`, `kuberneteslive`,
  `observability`, `ociregistry`, `packageregistry`, `sbom`,
  `securityalert`, `servicecatalog`, `terraformstate`, `vulnerability`, and
  `workitem`. This is an incremental migration (Contract System v1 §7); a fact
  kind without a typed struct yet still decodes through the untyped envelope
  path.
- Generated JSON Schema artifacts under `schema/<kind>.v1.schema.json` for
  every typed fact kind, produced with `invopop/jsonschema`.
- `fixturepack` subpackage — the versioned, importable payload-conformance
  artifact bundling the checked-in schemas plus one valid and one invalid
  example payload per kind, with a drift-lock test guaranteeing the embedded
  copy matches the canonical generated schema in the same commit.
- `DecodeError` classified error type (`ClassificationInputInvalid`) so a
  missing required field on decode dead-letters visibly instead of silently
  zeroing out.

## Convention for future entries

Add a new `## [Unreleased]` entry at the top of this file for every
merge-worthy change to this module — a new typed family, a schema version
bump, a fixture-pack payload change, or a decode-seam behavior change —
grouped under `### Added`, `### Changed`, `### Deprecated`, `### Removed`,
`### Fixed`, or `### Security` as needed (omit empty groups). Name the
affected fact kind(s) explicitly; "misc fixes" is not sufficient detail for a
contracts module external collectors pin against.

When a maintainer cuts a release:

1. Run generation, module tests, and a consumer schema compatibility review
   against the intended release baseline. The new repository does not yet
   carry the original Eshu schema-diff gate.
2. Rename the top `[Unreleased]` section to `[X.Y.Z] - YYYY-MM-DD` using the
   highest-severity bump any fact kind in the release needed.
3. Add a fresh empty `## [Unreleased]` section above it.
4. Follow the [release and cutover checklist](../docs/release.md) before
   tagging a new module release with `factschema/vX.Y.Z`. This would also
   version the fixture pack; it has no separate tag.
