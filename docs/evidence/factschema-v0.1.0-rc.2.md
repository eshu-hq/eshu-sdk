# Factschema `v0.1.0-rc.2` public consumer proof

On 2026-10-24, a fresh external Go consumer resolved the exact
`github.com/eshu-hq/eshu-sdk/factschema@v0.1.0-rc.2` version through
`https://proxy.golang.org` with `sum.golang.org`, `GOWORK=off`, fresh module
and build caches, and no `replace` directive. It did not use `@latest`.

| Check | Observed result |
| --- | --- |
| `go mod download -json github.com/eshu-hq/eshu-sdk/factschema@v0.1.0-rc.2` | Origin `refs/tags/factschema/v0.1.0-rc.2` at `94e162b568815f4be90489d4b74ceb1a56927b3c` |
| Module zip checksum | `h1:zgPRuX0OpqOpIO23Y9sp+nMd0x4Y9sNE3cnexgJWlp4=` |
| `go.mod` checksum | `h1:CnwND1yYkq0K/gMO7k746nEKmTKvVCRYrHB3TZpsJqk=` |
| `go list -m -json all` in the external consumer | Exact factschema `v0.1.0-rc.2` selected |
| `go build -buildvcs=false` and run the external consumer binary | Exit 0; stdout `9007199254740993` |

The consumer used `fixturepack.ValidPayload` for
`FactKindWorkItemProjectMetadata`, set `issue_count` to
`json.Number("9007199254740993")`, and called
`DecodeWorkItemProjectMetadata` with schema version `1.0.0`. It asserted that
`IssueCount` retained that exact integer. The first scratch build failed while
stamping the scratch module's local VCS status (`exit status 128`). Rebuilding
with `-buildvcs=false` passed; that flag only disabled local binary stamping
and did not change module resolution or verification.

This proves public candidate resolution and one typed numeric decoder path.
It does not qualify a collector/core pair, provider behavior, or a deployment.
At the time of this factschema verification, the preferred
`collector/v0.1.0-rc.2` candidate was untagged, and separate proxy verification
for historical `collector/v0.2.0-rc.1` had not been recorded.
