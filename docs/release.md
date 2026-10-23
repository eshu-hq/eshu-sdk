# Release and cutover checklist

The first versions planned for the new module identities are
`collector/v0.1.0-rc.1` and `factschema/v0.1.0-rc.1`. Neither tag is published.
Source preparation, candidate availability, consumer/core qualification, and
deployed acceptance are separate states. See the [compatibility policy](compatibility.md)
for the exact old and new paths and current scope.

1. Confirm an independently reviewed mapping of inherited schemas, fixtures,
   public APIs and decoder bindings, and review all new SDK behavior.
2. Run standalone module CI. Bind the decision packet and separate tag approval
   to the exact reviewed and merged source commit with green CI. Do not hardcode
   a source SHA in these docs, because a docs edit changes the commit.
3. Tag each module independently with its directory-prefixed candidate tag only
   after that approval. Preserve all published tags, including prereleases:
   repair an `rc.1` candidate with a reviewed `rc.2`, never a retarget.
4. From a fresh external consumer, resolve each exact candidate version with
   `GOWORK=off`, no `replace`, and no `@latest`. Record proxy checksums, the
   tag origin SHA, and a successful build. Candidate availability alone does
   not establish a qualified consumer or production compatibility.
5. Migrate consumers to exact published versions and run their contract and
   runtime gates. Grant-aware conformance blocks reserved-kind collector
   qualification. Reuse the existing core envelope adapter; retire duplicate
   generic helpers and identity derivation only at their documented adoption
   steps.
6. Validate the designated deployment and rehearse rollback without overlapping
   producer authority or lost facts. Mark a pair deployed only with matching
   runtime, reducer, graph, API, MCP, retry and operator evidence.
7. Remove superseded in-tree sources only after consumer production acceptance;
   preserve the existing `github.com/eshu-hq/eshu/sdk/go/{collector,factschema}`
   `v0.2.0` support and security policy during migration. Review a final stable
   `v0.1.0` separately; do not infer it from candidate publication.

The SDK now prepares transport, identity, generic contract assertions, a bounded
stdio runner and a neutral entrypoint generator. The core host still owns
admission, durable claims, persistence and producer authority. Its legacy hosted
manifests, templates and generated commands stay unchanged until each owning
collector has an accepted cutover; they are not SDK templates.

Provider request/result budgets require representative qualification. A local
synthetic command test is not release, provider or deployed acceptance.
