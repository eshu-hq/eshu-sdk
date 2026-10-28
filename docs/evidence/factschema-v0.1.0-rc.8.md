# Factschema `v0.1.0-rc.8` public proxy proof

On 2026-10-28, a fresh external Go consumer resolved the exact
`github.com/eshu-hq/eshu-sdk/factschema@v0.1.0-rc.8` version through
`https://proxy.golang.org` with `sum.golang.org`, `GOWORK=off`, fresh module
and build caches, and no `replace` directive. It did not use `@latest`.
Collector `v0.1.0-rc.4` is unchanged since the RC4/RC7 proof and was not
re-proven here.

| Check | Observed result |
| --- | --- |
| `git ls-remote origin refs/tags/factschema/v0.1.0-rc.8` | `d62cbeca9c99255c569e174f2494c5724e790499`, equal to the local tag (lightweight tag, exit 0) |
| `go list -m -versions` for the module via the proxy | Lists `v0.1.0-rc.8` among the published versions (exit 0) |
| `go mod download -json .../factschema@v0.1.0-rc.8` (cold cache) | Zip `h1:YJmrtyQdENb9ODWrAtqBgeUb/8gpOx/ztpSuhOg3ZFs=`, mod `h1:GysO1GSvcUYSSl7ydDx34MYRSw4yI2DgADT0twajVXY=`, Origin hash `d62cbeca...` at `refs/tags/factschema/v0.1.0-rc.8` (exit 0) |
| `go mod tidy` in the scratch consumer (cold cache, no `replace`) | Exit 0; downloads the pinned version from the proxy |
| `go build -buildvcs=false` and run the scratch consumer | Exit 0; stdout `factschema resolve: a0b91d5c69828db54f7828681499ee595e60ddcceae2cbc0f3a5a25b0d3c9be8` |
| `go vet ./...` in the scratch consumer | Exit 0 |

The consumer imported `factschema/identity` (calling `StableID`) and pinned
`factschema v0.1.0-rc.8` with `go 1.26.10`. The `-buildvcs=false` flag only
disabled local binary stamping and did not change module resolution or
verification.

This proves public candidate resolution and one consumer build for RC8. It
does not qualify a collector/core pair, provider behavior, or a deployment.
Consumer qualification (`eshu-collector-kubernetes-live` still pins collector
RC3 plus factschema RC6) and deployed acceptance remain `NOT_CHECKED`: no
core pair has run its contract and runtime gates against RC8.
