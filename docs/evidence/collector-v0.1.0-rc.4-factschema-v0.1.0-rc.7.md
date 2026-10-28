# Collector `v0.1.0-rc.4` + factschema `v0.1.0-rc.7` public proxy proof

On 2026-10-28, a fresh external Go consumer resolved the exact
`github.com/eshu-hq/eshu-sdk/collector@v0.1.0-rc.4` and
`github.com/eshu-hq/eshu-sdk/factschema@v0.1.0-rc.7` versions through
`https://proxy.golang.org` with `sum.golang.org`, `GOWORK=off`, fresh module
and build caches, and no `replace` directive. It did not use `@latest`.

| Check | Observed result |
| --- | --- |
| `git ls-remote origin refs/tags/collector/v0.1.0-rc.4` | `ef800f0ed5228b13eedd1003e36abd43d9691e4e`, equal to the local tag (lightweight tag, exit 0) |
| `git ls-remote origin refs/tags/factschema/v0.1.0-rc.7` (+ `^{}` peel) | Tag object `e46fe9584350376e00491f6e2ed14523a8cc803f`, peel `62845d95d939c2a9e7fba9f7c8745b46806f2cf3`, equal to the local peel (exit 0) |
| `go list -m -versions` for both modules via the proxy | Lists `v0.1.0-rc.4` (collector) and `v0.1.0-rc.7` (factschema) among the published versions (exit 0) |
| `go mod download -json .../collector@v0.1.0-rc.4` (cold cache) | Zip `h1:V8kE6u4KJ2S93po/9uAewswm8RAt38nwe6ppz6Kawy4=`, mod `h1:ysNLZcpQe9ZGYFZnTLyYTCHNaLrPRJjb1f9aLPOXdtA=`, Origin hash `ef800f0e...` at `refs/tags/collector/v0.1.0-rc.4` (exit 0) |
| `go mod download -json .../factschema@v0.1.0-rc.7` (cold cache) | Zip `h1:20Q3x+rqSSsQ2fEc4PvfiSFDaoQLNghsaTNgv5xSf5k=`, mod `h1:CnwND1yYkq0K/gMO7k746nEKmTKvVCRYrHB3TZpsJqk=`, Origin hash `62845d95...` at `refs/tags/factschema/v0.1.0-rc.7` (exit 0) |
| `go mod tidy` in the scratch consumer (cold cache, no `replace`) | Exit 0; downloads both pinned versions from the proxy |
| `go build -buildvcs=false` and run the scratch consumer | Exit 0; stdout `collector+factschema resolve: 95b5dd5b19fb4b9d69dd8f9a70b3427ce21ca506fb609478cbb69f586cf523e5` |
| `go vet ./...` in the scratch consumer | Exit 0 |

The consumer imported `collector/conformance` (referencing `Request`) and
`factschema/identity` (calling `StableID`), pinning `collector v0.1.0-rc.4`
plus `factschema v0.1.0-rc.7` with `go 1.26.10`. The first scratch build
failed while stamping the scratch module's local VCS status (`exit status
128`). Rebuilding with `-buildvcs=false` passed; that flag only disabled
local binary stamping and did not change module resolution or verification.

This proves public candidate resolution and one cross-module consumer build
for the RC4/RC7 pair. It does not qualify a collector/core pair, provider
behavior, or a deployment. Consumer qualification (`eshu-collector-kubernetes-live`
still pins collector RC3 plus factschema RC6) and deployed acceptance remain
`NOT_CHECKED`: no core pair has run its contract and runtime gates against
RC4/RC7.
