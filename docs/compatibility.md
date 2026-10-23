# SDK module compatibility and candidate status

The new repository contains two independent Go module identities. Source for
their first candidates is prepared as `v0.1.0-rc.1`. At preparation time, no
candidate tag had been published; check Git tags and the Go proxy for current
availability. A candidate tag makes bytes available for testing; it does not
qualify a core consumer or establish a deployed pair.

| Existing published module | Existing version | New module identity | Prepared first candidate |
| --- | --- | --- | --- |
| `github.com/eshu-hq/eshu/sdk/go/collector` | `v0.2.0` | `github.com/eshu-hq/eshu-sdk/collector` | `collector/v0.1.0-rc.1` |
| `github.com/eshu-hq/eshu/sdk/go/factschema` | `v0.2.0` | `github.com/eshu-hq/eshu-sdk/factschema` | `factschema/v0.1.0-rc.1` |

The old and new import paths name different Go types. The new `v0.1.0`
line is a new module identity, not a continuation of the old path's `v0.2.0`
line or a drop-in replacement. Both modules require Go `1.26.5` or newer.
The old `v0.2.0` paths, their source, and their support and security policy
remain in place until consumer and deployed acceptance authorize a cutover.

## Scope of the prepared candidate

The candidate includes the current collector contracts and tooling: transport,
conformance, public `Request`, bounded `command.Run`, neutral entrypoint
generation, and the optional `TrustedDeletion` attestation field. It includes
the current factschema typed contracts, schemas, fixture pack, identity helper,
and generic contract assertions. The unimplemented CoreFacts API is excluded.
Core still owns grant decisions, admission, graph truth, queues, leases, and
durable state. A grant-aware conformance result is required before a reserved
fact kind can qualify a collector; this does not narrow availability of the
candidate modules for other consumer tests.

The comparison against the supported old `factschema@v0.2.0` has 176 old
schemas and 179 current schemas: three added, none removed, and four existing
schemas with optional nullable field additions. This is an additive schema
comparison, not proof that old core accepts new SDK output. The
[migration evidence](sdk-migration-evidence.md) records the source counts,
per-file changes, and commands to reproduce them.

## Acceptance states

| State | Evidence required | Status at source preparation |
| --- | --- | --- |
| Candidate available | Exact reviewed and merged source commit, green CI on that commit, separate tag approval, immutable directory-prefixed tags, and fresh proxy resolution | Source prepared; publication requires live Git/proxy checks |
| Qualified consumer/core pair | Exact candidate pins, contract and runtime gates for that consumer and core version, including grant-aware conformance for reserved kinds | Not qualified |
| Deployed accepted pair | Matching deployment, reducer, graph, API/MCP, retry, operator, and rollback evidence | Not accepted |

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

The [migration evidence](sdk-migration-evidence.md) also records the original
module tag origins and Go checksums. See [release and cutover](release.md) for
the order of approval and qualification.
