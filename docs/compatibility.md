# SDK module compatibility and candidate status

The repository contains two independent Go module identities. The exact Git
refs available in this checkout include collector `v0.1.0-rc.3` and factschema
`v0.1.0-rc.5`; earlier candidate tags remain immutable. This branch prepares
factschema `v0.1.0-rc.6` with optional EPSS model-version preservation but
does not publish its tag or qualify an Eshu consumer. Recheck tag origin and
fresh public-proxy resolution before claiming candidate availability.

| Existing supported module | New module identity | Latest checked Git tag | Prepared next candidate |
| --- | --- | --- | --- |
| `github.com/eshu-hq/eshu/sdk/go/collector@v0.2.0` | `github.com/eshu-hq/eshu-sdk/collector` | `collector/v0.1.0-rc.3` | none in this change |
| `github.com/eshu-hq/eshu/sdk/go/factschema@v0.2.0` | `github.com/eshu-hq/eshu-sdk/factschema` | `factschema/v0.1.0-rc.5` | `factschema/v0.1.0-rc.6` (unpublished) |

The old and new import paths name different Go types. The new `v0.1.0`
line is a new module identity, not a continuation of the old path's `v0.2.0`
line or a drop-in replacement. Both modules require Go `1.26.5` or newer.
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

The collector RC3, factschema RC4, and factschema RC5 source changes have
separate immutable Git tags. This RC6 preparation adds only the optional EPSS
`model_version` field to the factschema payload contract. FIRST API score rows
do not provide it, so consumer proof must test conditional retention without
claiming live source acquisition.

## Acceptance states

The recorded status below is a source-preparation snapshot from 2026-10-27.
Git tag presence was checked locally; fresh public-proxy proof for RC4, RC5,
and the future RC6 remains NOT_CHECKED here.

| State | Evidence required | Recorded snapshot |
| --- | --- | --- |
| Candidate available | Merged core compatibility mapping, exact reviewed and merged source commit, green CI, separate tag approval, immutable directory-prefixed tag, and fresh proxy resolution | Factschema RC5 and collector RC3 Git tags exist; RC6 is prepared, not tagged or proxy-verified |
| Qualified consumer/core pair | Exact candidate pin plus contract and runtime gates for that consumer and core version | RC6 consumer NOT_CHECKED |
| Deployed accepted pair | Matching deployment, reducer, graph, API/MCP, retry, operator, and rollback evidence | RC6 deployment NOT_CHECKED |

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
The [migration evidence](sdk-migration-evidence.md) also records the original
module tag origins and Go checksums. See [release and cutover](release.md) for
the order of approval and qualification.

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
