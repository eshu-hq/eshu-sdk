# Trusted core fact context for collector conformance

`conformance.Request.CoreFacts` is an optional, trusted, per-run snapshot that
lets `Run` evaluate a component manifest and result fixtures that emit reserved
core fact kinds. The field has `json:"-" yaml:"-"` tags: a manifest or fixture
cannot supply its own authorization by decoding into a request. An out-of-tree
caller without trusted host context leaves it nil and keeps the existing
namespaced, source-only conformance behavior and report bytes.

The context binds `ComponentID` and `ComponentVersion` exactly to the validated
manifest. Each `CoreFactAuthorization` names a core kind in the caller's complete
`ReservedFactKinds` inventory, lists permitted `CoreFactEmission` collector
kinds with their schema versions, and supplies trusted reducer phases and an
optional payload schema reference. The SDK rejects missing, duplicate,
undeclared, or noncanonical coverage. Every declared version of a reserved kind
needs coverage under at least one declared collector kind. Each fixture,
including a zero-fact fixture, must have the manifest component and a declared
collector kind. Each emitted reserved fact must match one exact
collector-kind/fact-kind/schema-version tuple; separate permissions do not
combine into a new tuple.

The manifest must declare the union of trusted reducer phases. It may also
declare `source_evidence_only:no_graph_truth`, and no other extra phase.
Optional namespaced kinds retain the source-only rule when no context is
supplied. A trusted `PayloadSchemaRef` requires caller-supplied
`PayloadSchemas[kind]` even when the manifest omits its own reference; a
conflicting manifest reference or invalid schema fails closed. The existing
payload validator checks fixture data against supplied schemas. A manifest
cannot invent a schema reference for a core kind whose trusted context has
none; such a kind remains explicitly untyped.

The host is the authority for the durable grants, expiry, revocation,
compatible registry versions, complete reserved inventory, canonical reducer
phase mapping, and payload schema bytes. It must derive a fresh context from
one grant snapshot and explicit evaluation time for each run. The SDK neither
loads grants nor decides whether they remain valid later. Its portable tests
use `proof.synthetic:context` solely to prove phase matching; that token is
not a core registry mapping. This API is a conformance mechanism, not a
release, deployed provider, or runtime permission proof.

`Run` remains pure and performs no file or network I/O. It adds no telemetry,
global state, locks, clock checks, or retry loop. Existing `Report`, `Finding`,
`Summary`, finding codes, and JSON tags remain unchanged.
