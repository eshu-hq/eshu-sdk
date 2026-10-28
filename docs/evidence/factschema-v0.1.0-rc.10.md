# Factschema `v0.1.0-rc.10` public proxy proof

The `factschema/v0.1.0-rc.10` tag is a lightweight tag on
`f921babcf8dffc9b04d84f3fc412a7cd76580937`, the merge commit of
[SDK PR #60](https://github.com/eshu-hq/eshu-sdk/pull/60) (merged
2026-10-28T20:32:15Z). The push-to-`main` CI run
[40764167499](https://github.com/eshu-hq/eshu-sdk/actions/runs/40764167499) on
that exact commit completed with `success`, including `ci-complete`. The tag
was pushed on 2026-10-28 with the owner's approval naming the tag and commit.

On 2026-10-28, two fresh external Go consumers resolved the exact
`github.com/eshu-hq/eshu-sdk/factschema@v0.1.0-rc.10` version through
`https://proxy.golang.org` with `sum.golang.org`, `GOWORK=off`, empty module and
build caches, and no `replace` directive. Neither used `@latest`. The first run
was the post-tag fetch; the second, at 21:16Z, repeated it independently in a
new scratch directory. Both recorded the same checksums and origin.

| Check | Observed result |
| --- | --- |
| `git ls-remote origin refs/tags/factschema/v0.1.0-rc.10` | `f921babcf8dffc9b04d84f3fc412a7cd76580937`, equal to the local tag (lightweight tag, exit 0) |
| `go list -m -versions` for the module via the proxy | Lists `v0.1.0-rc.1` through `v0.1.0-rc.10` (exit 0) |
| `go mod download -json .../factschema@v0.1.0-rc.10` (cold cache) | Zip `h1:+8IpD4At2XmaSePQ5FdYi6BFAXQA9a1RRnrAYfMbdcc=`, mod `h1:GysO1GSvcUYSSl7ydDx34MYRSw4yI2DgADT0twajVXY=`, Origin `https://github.com/eshu-hq/eshu-sdk` subdir `factschema`, hash `f921babc...` at `refs/tags/factschema/v0.1.0-rc.10` (exit 0) |
| `go get .../factschema@v0.1.0-rc.10` and `go mod tidy` (cold cache, no `replace`) | Exit 0 on the first attempt; `go.sum` carries the same two `h1:` values |
| `go build -buildvcs=false` and run the scratch consumer | Exit 0; `fixturepack.SchemaFor("vulnerability.source_snapshot")` returned `true` and 2388 bytes |
| Typed decode in the second consumer | Exit 0; `json.Unmarshal` into `vulnerability/v1.SourceSnapshot` filled `CoverageDescriptor` (mode `cve_ids`), `CollectorInstanceID`, and `ResultCount` |
| `go vet ./...` in the scratch consumer | Exit 0 |
| Cached fixture-pack schema | `vulnerability.source_snapshot.v1.schema.json` declares `coverage_descriptor` |

The consumers pinned `factschema v0.1.0-rc.10` with `go 1.26.10`. The
`-buildvcs=false` flag only disabled local binary stamping and did not change
module resolution or verification. The mod `h1:` equals RC8's because the
module's `go.mod` file is unchanged since RC8.

## What the release carries

RC10 adds three optional fields to `vulnerability.source_snapshot` (core
eshu-hq/eshu#9227):

- `coverage_descriptor`: a versioned, normalized record of an NVD snapshot's
  target: the explicit CVE set (trimmed, upper case, deduplicated, sorted),
  query mode and window, page size, configured endpoint set, and the endpoint
  actually used. It carries no credential or cadence material.
- `collector_instance_id`: the source instance that committed the snapshot.
- `result_count`: the number of source rows the observation reported.

These are additive optional fields, a minor change. `required` stays `source`
and `complete`, and no stable key changes. The canonical schema and the
fixture-pack copy match byte for byte.

## Requirements for consumers

- Pin the exact version `v0.1.0-rc.10` with the checksums above. Do not use
  `@latest`, a workspace, or a `replace` directive.
- Use Go `1.26.10` or newer.
- Treat a snapshot without `coverage_descriptor` as a valid source
  observation that cannot certify zero advisory coverage. Older snapshots
  stay valid.

## What this does not prove

This proves public candidate resolution and consumer builds for RC10. Core
PR eshu-hq/eshu#9435 pins this immutable candidate; until it merges, core
`main` still pins `factschema v0.1.0-rc.7`. Reducer, query, and deployed
acceptance remain separate proof: no core pair has run its contract and runtime
gates against RC10, and deployment is `NOT_CHECKED`.
