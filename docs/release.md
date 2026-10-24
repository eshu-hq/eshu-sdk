# Release and cutover checklist

Both first `v0.1.0-rc.1` module candidates are available through immutable
directory-prefixed tags and verified fresh public-proxy downloads from the
source in [SDK PR #9](https://github.com/eshu-hq/eshu-sdk/pull/9). The later `collector/v0.2.0-rc.1` marker and `factschema/v0.1.0-rc.2`
numeric-decoder tags are also published and immutable. The owner selected
`collector/v0.1.0-rc.2` as the preferred next collector candidate on the
initial prerelease line; that tag is prepared, not published. The published
collector `v0.2.0-rc.1` is a historical candidate, not the preferred pin.
Check Git tags and fresh proxy origins before acting on availability;
qualification and deployed acceptance remain separate. See the [compatibility policy](compatibility.md).

1. Confirm an independently reviewed mapping of inherited schemas, fixtures,
   public APIs and decoder bindings, and review all new SDK behavior. Before
   the preferred next collector tag, confirm the path mapping in the core
   [SDK compatibility matrix](https://github.com/eshu-hq/eshu/blob/main/docs/public/extend/sdk-compatibility.md).
   The standalone SDK mapping alone does not satisfy that core prerequisite.
2. Confirm separately reviewed versioned metadata for each intended tag and
   its core compatibility-mapping row. The collector changelog prepares
   `0.1.0-rc.2`; the factschema changelog records its published `0.1.0-rc.2`.
   Preserve the first candidates and the published historical collector
   `0.2.0-rc.1` entry. The optional
   trusted `conformance.Request.CoreFacts` API remains part of collector RC1;
   its grant and supported-pair requirements still apply. Source preparation
   is not tag approval or a consumer qualification.
3. Run standalone module CI. Bind the decision packet and separate tag approval
   to the exact reviewed and merged source commit with green CI. Do not hardcode
   a source SHA in these docs, because a docs edit changes the commit.
4. Tag the preferred collector `collector/v0.1.0-rc.2` only after that
   approval. The factschema `factschema/v0.1.0-rc.2` tag is already published;
   verify its exact origin rather than republishing it. Preserve all published
   tags, including both first RC1s and collector `v0.2.0-rc.1`. Never retarget
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
