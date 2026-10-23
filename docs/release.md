# Release and cutover checklist

No SDK module release is selected or published. Source preparation, an immutable
candidate, supported consumer qualification, and deployed acceptance are
separate states.

1. Confirm an independently reviewed mapping of inherited schemas, fixtures,
   public APIs and decoder bindings, and review all new SDK behavior.
2. Run standalone module CI. Prove resolution of an exact immutable candidate
   without a workspace or local replacement, before a consumer depends on it.
3. Select module versions and the directory-prefixed tags with the accountable
   release owner. Record the old/new module paths, supported core/SDK pairs,
   wire versions and qualification evidence in the compatibility policy.
4. Publish only the reviewed candidate or release from a clean, fast-forwarded
   main checkout. Preserve existing immutable tags. Publication alone does not
   establish production compatibility.
5. Migrate consumers to exact published versions and run their contract and
   runtime gates. Reuse the existing core envelope adapter; retire duplicate
   generic helpers and identity derivation only at their documented adoption
   steps.
6. Validate the designated deployment and rehearse rollback without overlapping
   producer authority or lost facts. Mark a pair deployed only with matching
   runtime, reducer, graph, API, MCP, retry and operator evidence.
7. Remove superseded in-tree sources only after consumer production acceptance;
   preserve the existing published-module support policy during migration.

The SDK now prepares transport, identity, generic contract assertions, a bounded
stdio runner and a neutral entrypoint generator. The core host still owns
admission, durable claims, persistence and producer authority. Its legacy hosted
manifests, templates and generated commands stay unchanged until each owning
collector has an accepted cutover; they are not SDK templates.

Provider request/result budgets require representative qualification. A local
synthetic command test is not release, provider or deployed acceptance.
