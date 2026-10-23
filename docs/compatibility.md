# SDK module compatibility and candidate status

The new repository contains two independent Go module identities. Its first
candidate versions are **planned, not published**. A candidate tag makes bytes
available for testing; it does not qualify a core consumer or establish a
deployed pair.

| Existing supported module | Existing version | New module identity | Planned first candidate |
| --- | --- | --- | --- |
| `github.com/eshu-hq/eshu/sdk/go/collector` | `v0.2.0` | `github.com/eshu-hq/eshu-sdk/collector` | `collector/v0.1.0-rc.1` |
| `github.com/eshu-hq/eshu/sdk/go/factschema` | `v0.2.0` | `github.com/eshu-hq/eshu-sdk/factschema` | `factschema/v0.1.0-rc.1` |

The old and new import paths name different Go types. The proposed `v0.1.0`
line is a new module identity, not a continuation of the old path's `v0.2.0`
line or a drop-in replacement. Both modules require Go `1.26.5` or newer.
The old `v0.2.0` paths, their source, and their support and security policy
remain in place until consumer and deployed acceptance authorize a cutover.

## Scope of the planned candidate

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
comparison, not proof that old core accepts new SDK output. The source counts
and per-file changes are recorded in the local candidate evidence artifacts
`sdk-old-new-schema-census.json` and `sdk-old-new-schema-diff.json`.

## Acceptance states

| State | Evidence required | Current status |
| --- | --- | --- |
| Candidate available | Exact reviewed and merged source commit, green CI on that commit, separate tag approval, immutable directory-prefixed tags, and fresh proxy resolution | Planned; not published |
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

The prior published module records are in the local evidence artifact
`sdk-supported-old-modules-proxy.json`. See [release and cutover](release.md)
for the order of approval and qualification.
