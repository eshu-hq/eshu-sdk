# Release and cutover checklist

Both first `v0.1.0-rc.1` module candidates are available through immutable
directory-prefixed tags and verified fresh public-proxy downloads from the
source in [SDK PR #9](https://github.com/eshu-hq/eshu-sdk/pull/9). Source now
prepares `collector/v0.2.0-rc.1` for the additive marker API and
`factschema/v0.1.0-rc.2` for the typed numeric decoder correction. The next
tags are not published by this preparation. Check Git tags and proxy origins
before acting on availability; qualification and deployed acceptance remain
separate. See the [compatibility policy](compatibility.md).

1. Confirm an independently reviewed mapping of inherited schemas, fixtures,
   public APIs and decoder bindings, and review all new SDK behavior. Before
   either new-path tag, merge the path mapping into the core
   [SDK compatibility matrix](https://github.com/eshu-hq/eshu/blob/main/docs/public/extend/sdk-compatibility.md).
   The standalone SDK mapping alone does not satisfy that core prerequisite.
2. Confirm separately reviewed versioned metadata for each intended tag and
   its core compatibility-mapping row. The next collector changelog prepares
   `0.2.0-rc.1`; the next factschema changelog prepares `0.1.0-rc.2`.
   Preserve the first candidates and their historical entries. The optional
   trusted `conformance.Request.CoreFacts` API remains part of collector RC1;
   its grant and supported-pair requirements still apply. Source preparation
   is not tag approval or a consumer qualification.
3. Run standalone module CI. Bind the decision packet and separate tag approval
   to the exact reviewed and merged source commit with green CI. Do not hardcode
   a source SHA in these docs, because a docs edit changes the commit.
4. Tag each module independently with its directory-prefixed candidate tag only
   after that approval. Preserve all published tags, including prereleases:
   repair the factschema candidate with reviewed `rc.2` bytes; additive
   collector APIs use their separately reviewed minor candidate. Never retarget
   a published tag.
5. From a fresh external consumer, resolve each exact candidate version with
   `GOWORK=off`, no `replace`, and no `@latest`. Record proxy checksums, the
   tag origin SHA, and a successful build. Candidate availability alone does
   not establish a qualified consumer or production compatibility.
6. Migrate consumers to exact published versions and run their contract and
   runtime gates. The caller must derive fresh, authoritative grant coverage,
   expiry/revocation, scope, version, reducer phases, and payload schemas for
   `CoreFacts`; the SDK cannot qualify a reserved-kind collector on its own.
   Reuse the existing core envelope adapter; retire duplicate
   generic helpers and identity derivation only at their documented adoption
   steps.
7. Validate the designated deployment and rehearse rollback without overlapping
   producer authority or lost facts. Mark a pair deployed only with matching
   runtime, reducer, graph, API, MCP, retry and operator evidence.
8. Remove superseded in-tree sources only after consumer production acceptance;
   preserve the existing `github.com/eshu-hq/eshu/sdk/go/{collector,factschema}`
   `v0.2.0` support and security policy during migration. Review a final
   stable module version separately; do not infer it from candidate publication.

The SDK now prepares transport, identity, generic contract assertions, a bounded
stdio runner and a neutral entrypoint generator. The core host still owns
admission, durable claims, persistence and producer authority. Its legacy hosted
manifests, templates and generated commands stay unchanged until each owning
collector has an accepted cutover; they are not SDK templates.

Provider request/result budgets require representative qualification. A local
synthetic command test is not release, provider or deployed acceptance.
