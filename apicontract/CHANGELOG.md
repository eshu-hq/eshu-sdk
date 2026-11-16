# Changelog

The new `github.com/eshu-hq/eshu-sdk/apicontract` path starts an
independent version line; it is not a continuation of the in-tree
`github.com/eshu-hq/eshu/sdk/go/apicontract` module. Candidate entries
describe source changes; tag publication and production consumer support
are separate states. Check Git tags and the Go proxy for current
availability. See the [compatibility policy](../docs/compatibility.md).

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/):
changing the frozen `EnvelopeMIMEType` string is a breaking wire change
(major bump); adding a backward-compatible constant is a minor bump; a
docs-only change is a patch for stable releases.

Tags for the new module use the Go subdirectory-module format after
separate release approval: `apicontract/vX.Y.Z`.

## [Unreleased]

### Added

- Frozen `EnvelopeMIMEType` (`application/eshu.envelope+json`), copied from
  Eshu core `go/internal/query/querycontract` (`http.go`), pinned
  byte-for-byte by `contract_test.go`. The module uses only the Go
  standard library and imports nothing core-internal.
