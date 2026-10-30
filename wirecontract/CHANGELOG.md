# Changelog

This file records the new `github.com/eshu-hq/eshu-sdk/wirecontract` module
path. It starts an independent version line; there is no prior in-tree tag
for this path. Candidate entries describe source changes; tag publication
and production consumer support are separate states. Check Git tags and
the Go proxy for current availability. See the
[compatibility policy](../docs/compatibility.md).

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this module follows the versioning rule in
[Contract System v1 §5](https://github.com/eshu-hq/eshu/blob/main/docs/internal/design/contract-system-v1.md#5-versioning-and-compatibility-policy):
a renamed field, retag, narrowed type, or changed meaning is a major bump;
an additive optional field or entry is a minor bump; docs-only is a patch.
The preferred prerelease line remains `v0.1.0` until stable qualification.

Tags for the new module use the Go subdirectory-module format after separate
release approval: `wirecontract/vX.Y.Z`.

## [Unreleased]

### Added

- Initial module (Refs eshu-hq/eshu#9509, part of #6707): publishes the API
  wire shapes a thin Tier-2 CLI needs — response/truth/error envelopes,
  query profiles, graph backends, freshness causes, the MCP tool-definition
  shape, and the client-scoped environment registry with its pure validator
  — with JSON goldens pinned byte-for-byte to the core `sdk/go/apicontract`
  types and parity tests on the SDK side. No `go/internal` imports; stdlib
  only. The `wirecontract/v0.1.0-rc.1` tag is NOT cut here — a maintainer
  cuts it per [release and cutover](../docs/release.md) after merge.

## Convention for future entries

Add a new `## [Unreleased]` entry at the top of this file for every
merge-worthy user-facing change to this module, grouped under `### Added`,
`### Changed`, `### Deprecated`, `### Removed`, `### Fixed`, or `### Security`
as needed (omit empty groups). When a maintainer cuts a release:

1. Rename the top `[Unreleased]` section to `[X.Y.Z] - YYYY-MM-DD` using the
   version chosen per the breaking/additive/patch rule above.
2. Add a fresh empty `## [Unreleased]` section above it.
3. Follow the [release and cutover checklist](../docs/release.md) before
   tagging a new module release with `wirecontract/vX.Y.Z`.

A wire-protocol-breaking change additionally needs a new protocol version
per Contract System v1 §5; do not reuse the existing wire shape for an
incompatible change.
