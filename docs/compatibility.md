# SDK module compatibility and candidate status

The repository contains three independent Go module identities. The exact Git
refs available in this checkout include collector `v0.1.0-rc.4` and factschema
`v0.1.0-rc.10`; earlier candidate tags remain immutable. Neither tag qualifies
an Eshu consumer on its own: recheck tag origin and fresh public-proxy
resolution before claiming candidate availability. The `wirecontract` module
is new in this change and has no published tag yet.

| Existing supported module | New module identity | Latest checked Git tag | Prepared next candidate |
| --- | --- | --- | --- |
| `github.com/eshu-hq/eshu/sdk/go/collector@v0.2.0` | `github.com/eshu-hq/eshu-sdk/collector` | `collector/v0.1.0-rc.4` | none in this change |
| `github.com/eshu-hq/eshu/sdk/go/factschema@v0.2.0` | `github.com/eshu-hq/eshu-sdk/factschema` | `factschema/v0.1.0-rc.10` | none in this change |
| `github.com/eshu-hq/eshu/sdk/go/apicontract@v0.1.0` (release-prepared in core, tag NOT cut) | `github.com/eshu-hq/eshu-sdk/wirecontract` | none (new module) | `wirecontract/v0.1.0-rc.1` |

The old and new import paths name different Go types. The new `v0.1.0`
line is a new module identity, not a continuation of the old path's `v0.2.0`
line or a drop-in replacement. All three modules require Go `1.26.10` or newer.
The old `v0.2.0` paths, their source, and their support and security policy
remain in place until consumer and deployed acceptance authorize a cutover.

## Scope of the candidates

The candidate includes the current collector contracts and tooling: transport,
conformance, public `Request`, bounded `command.Run`, neutral entrypoint
generation, and the optional `TrustedDeletion` attestation field. It includes
the current factschema typed contracts, schemas, fixture pack, identity helper,
and generic contract assertions. The collector candidate also includes the
optional `conformance.Request.CoreFacts` context for trusted, exact-tuple
coverage of reserved kinds. Core still owns grant decisions, admission, graph
truth, queues, leases, and durable state. A grant-aware conformance result is
required before a reserved fact kind can qualify a collector; candidate
availability alone does not qualify that collector.

The comparison against the supported old `factschema@v0.2.0` has 176 old
schemas and 179 current schemas: three added, none removed, and four existing
schemas with optional nullable field additions. This is an additive schema
comparison, not proof that old core accepts new SDK output. The
[migration evidence](sdk-migration-evidence.md) records the source counts,
per-file changes, and commands to reproduce them.

The preferred collector `v0.1.0-rc.2` candidate includes the neutral `marker` package from
[SDK PR #10](https://github.com/eshu-hq/eshu-sdk/pull/10), an additive Go API
change. The immutable collector `v0.2.0-rc.1` tag already includes it, but is
a historical candidate rather than the preferred pin. Factschema `v0.1.0-rc.2`
repairs number-preserving typed decoding from
[SDK PR #11](https://github.com/eshu-hq/eshu-sdk/pull/11), including exact
integral decimal/exponent values and nested integer lists. Schemas, fixtures
and wire protocols are unchanged by these two earlier changes.

The collector RC3 and RC4 and factschema RC4 through RC10 source changes have
separate immutable Git tags. RC6 adds the optional EPSS `model_version` field;
FIRST API rows do not supply it. RC7 adds the synthetic NVD paging golden
without changing a wire schema. RC8 publishes the `terraformstate` drift
attribute allowlist as Go API only, with no payload schema or fixture change.
RC9 unexports that allowlist table behind its accessor functions, a Go API
change with no payload schema or fixture change. RC10 adds the optional NVD
`coverage_descriptor`, `collector_instance_id`, and `result_count` fields to
`vulnerability.source_snapshot` (core eshu-hq/eshu#9227), an additive minor
schema change with no stable-key change.
Collector RC4 allowlists the redaction-safe
`env_from_secret` boolean on `kubernetes_live.pod_template` payloads
([SDK PR #28](https://github.com/eshu-hq/eshu-sdk/pull/28), issue #26). Core
and extracted consumers must prove the same pinned fixture bytes, payloads,
stable keys, and two-page traversal.

The new `wirecontract` candidate publishes the Tier-2 API wire shapes only:
response/truth/error envelopes, query profiles, graph backends, freshness
causes, the MCP tool-definition shape, and the client-scoped environment
registry with its pure validator. It ports core `sdk/go/apicontract`
(Refs eshu-hq/eshu#9509, port source `c606960c97`) with no `go/internal`
imports and stdlib only. The two-sided oracle runs in this repo's gates as
`scripts/ci/test-wirecontract-core-pin.sh` (also via `make shape`): it
rebuilds the marshal goldens from the core live types at pin
`c606960c97` and from the ported live types, requiring
golden == core == port on all four. Core's own suite regenerates its
goldens from live types (green at `fb4b777757ab`), and the guard carries
a seeded-violation RED/GREEN pair (a mistranscribed tag fails the port
side only). The in-module tests then pin SDK output to those bytes. Run
`cd wirecontract && GOWORK=off go test ./... -count=1` to reproduce the
SDK side.

## Acceptance states

The recorded status below is a snapshot from 2026-10-28. Tag origin was
checked against `origin` (`collector/v0.1.0-rc.4` and
`factschema/v0.1.0-rc.10` match locally), and both tags resolve, download,
and build through the public Go proxy with `GOWORK=off` and no `replace`
directive (the RC4/RC7 proof records a scratch consumer pinning
`collector v0.1.0-rc.4` plus `factschema v0.1.0-rc.7` through tidy, build,
vet, and run, and the RC10 proof records scratch consumers pinning
`factschema v0.1.0-rc.10` through the same steps).

| State | Evidence required | Recorded snapshot |
| --- | --- | --- |
| Candidate available | Merged core compatibility mapping, exact reviewed and merged source commit, green CI, separate tag approval, immutable directory-prefixed tag, and fresh proxy resolution | Collector RC4 and factschema RC10 Git tags exist and are proxy-verified as above; wirecontract has no tag yet (source prepared in this change, tag approval and proxy verification pending per [release and cutover](release.md)) |
| Qualified consumer/core pair | Exact candidate pin plus contract and runtime gates for that consumer and core version | RC10 consumer NOT_CHECKED: core PR eshu-hq/eshu#9435 carries the RC10 pin and core `main` still pins factschema RC7; `eshu-collector-kubernetes-live` still pins collector RC3 plus factschema RC6 and its `StillBlocked` pin records the pre-fix expectation; wirecontract pair parity pinned, NOT qualified (module-pair scope): `github.com/eshu-hq/eshu-sdk/wirecontract` at this source pins byte-identical JSON to core `sdk/go/apicontract` via committed `wirecontract/testdata/golden/*.json` and green `TestGoldenEnvelopeBytes`, `TestGoldenErrorEnvelope`, `TestGoldenToolDefinition`, `TestClientEntriesMatchGolden`, `TestValidateVectors` — unverified against a core consumer gate; consumer qualification and Tier-2 CLI repoint (Box 3) follow the published tag |
| Deployed accepted pair | Matching deployment, reducer, graph, API/MCP, retry, operator, and rollback evidence | RC10 deployment NOT_CHECKED; wirecontract deployment NOT_CHECKED |

Use exact versions in consumer `go.mod` files; do not use `@latest`, a workspace,
or a `replace` directive as release evidence. Before a tag, bind approval to
the exact reviewed and merged commit in the decision packet. Do not copy a
source SHA into this document: its own edits change the candidate commit.
After tagging, verify each path and version through a fresh external consumer
with `GOWORK=off`, no replacement, proxy checksums, the tag's origin SHA, and a
successful build. Never retarget a published tag, including a prerelease. If
`rc.1` needs repair, review a new commit and use `rc.2`. A stable `v0.1.0`
release needs separate evidence and approval; it does not follow automatically
from a candidate.

The [factschema RC2 public-consumer proof](evidence/factschema-v0.1.0-rc.2.md)
records its exact tag origin, checksums, and typed-decoder smoke check.
The [collector RC4 plus factschema RC7 public-consumer proof](evidence/collector-v0.1.0-rc.4-factschema-v0.1.0-rc.7.md)
records the pair's tag-origin match, proxy checksums, and scratch-consumer build.
The [factschema RC8 public-consumer proof](evidence/factschema-v0.1.0-rc.8.md)
records the RC8 tag-origin match, proxy checksums, and scratch-consumer build.
The [factschema RC10 public-consumer proof](evidence/factschema-v0.1.0-rc.10.md)
records the RC10 tag-origin match, proxy checksums, scratch-consumer builds,
and a typed decode of the new `coverage_descriptor`.
The [migration evidence](sdk-migration-evidence.md) also records the original
module tag origins and Go checksums. See [release and cutover](release.md) for
the order of approval and qualification.

## MCP setup snippet scope decision (Box 4)

`mcp setup` offline snippets MUST NOT use remote `tools/list` instead of the
published `wirecontract.ToolDefinition` shape. Evidence: core
`runMCPSetup` renders snippets offline from the static platform table plus
`SetupRequest` (mode, posture, service URL) with no network and no
`tools/list` (`go/internal/cli/mcpsetup/snippet.go` takes no tool
definitions); `RunVerification` reuses the embedded in-process
`mcp.ReadOnlyTools` catalog for the tools-visible count, and the hosted
first-query probe uses the API client (`GET /health`, `GET
/api/v0/index-status`) only when the posture holds a credential — otherwise
those stages skip. Remote `tools/list` needs a running MCP server with the
caller's credential chain (API key, scoped tokens, or OIDC resource URI;
headerless requests to non-public routes are rejected 401), and its
advertisement varies per deployment via `ESHU_MCP_TOOL_SURFACE` and
`ESHU_QUERY_PROFILE`. Offline snippet generation therefore cannot depend on
it. Recommendation: keep offline rendering on the published shape for
decoding remote `tools/list` responses (verification counts); do not replace
snippet generation with a live call. Full evidence lives in
`wirecontract/README.md`.

## Trusted conformance context in the candidate

The expanded first collector candidate includes
[`conformance.Request.CoreFacts`](conformance-grants.md). This optional Go API
binds a caller-evaluated grant snapshot to the manifest's component identity,
reserved kinds, exact collector and schema-version emission tuples, reducer
phases, and payload schema references. Nil context keeps the prior source-only
behavior and report bytes. The field is excluded from JSON and YAML decoding;
it does not change the collector wire protocol or report schema. Adding a field
to the exported `Request` struct can break external unkeyed composite literals;
callers using keyed literals are unaffected by the new field.

The caller must supply a complete reserved-kind inventory and authoritative,
fresh grant, expiry, revocation, scope, version, phase, and schema evidence.
It must keep all request inputs immutable while `Run` executes. The SDK checks
structural coverage and fixture tuples but does not load grants, assign core
phases, or authorize a live emission. Neither this candidate source nor an
eventual tag establishes a qualified collector/core or deployed pair.
