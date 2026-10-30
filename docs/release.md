# Release and cutover checklist

## Current wirecontract RC1 preparation (Refs eshu-hq/eshu#9509)

The `wirecontract` module is source-prepared in this change with no tag cut.
It ports core `sdk/go/apicontract` byte-for-byte (envelopes, profiles,
backends, freshness, tool shape, client env registry plus validator) with
`wirecontract/testdata/golden/*.json` pinned to the core bytes and
in-module parity tests green (`cd wirecontract && GOWORK=off go test
./... -count=1`). It has no `go/internal` imports and stdlib only. Follow
the checklist below for `wirecontract/v0.1.0-rc.1` up to but not including
the tag push: reviewed mapping (this branch), versioned metadata
(`wirecontract/CHANGELOG.md` Unreleased entry plus `docs/compatibility.md`
row and module-pair qualification), standalone module CI green
(`make build test vet lint tidy-check file-cap race govulncheck shape`),
then separate tag approval and a fresh external `GOWORK=off` consumer probe
before any tag. Tier-2 CLI repoint (Box 3) follows the published tag and is
out of scope for this lane. No proxy verification, consumer qualification,
or deployed acceptance is claimed here.

## Current RC4/RC10 state

Git tags in this checkout reach `collector/v0.1.0-rc.4` and
`factschema/v0.1.0-rc.10`. RC10 adds the optional NVD `coverage_descriptor`,
`collector_instance_id`, and `result_count` fields to
`vulnerability.source_snapshot` (core eshu-hq/eshu#9227, PR #60) with no
stable-key change; RC9 unexports the `terraformstate` drift allowlist table
(PR #53); RC8 publishes that allowlist as Go API only (PR #50) with no payload
schema or fixture change; RC7 carries the shared NVD paging input-to-payload
golden (core #9225); RC4 carries the `env_from_secret` allowlist fix (issue
#26, PR #28) on top of the RC3 validation corrections. These tags match their
`origin` refs and are listed by the public Go proxy. Fresh external
`GOWORK=off` consumers with no `replace` or `@latest` resolved, downloaded,
and built three of them: the RC4/RC7 proof records one scratch consumer
pinning `collector v0.1.0-rc.4` plus `factschema v0.1.0-rc.7`, the RC8 proof
records a second consumer pinning `factschema v0.1.0-rc.8`, and the RC10 proof
records consumers pinning `factschema v0.1.0-rc.10`. RC9 has no recorded
consumer proof.
Consumer qualification and deployed acceptance remain separate: the pinned
satellite still runs collector RC3 plus factschema RC6, core `main` still pins
factschema RC7, and core PR eshu-hq/eshu#9435 carries the RC10 pin. No core
pair has run its contract and runtime gates against RC4/RC10. Candidate
publication alone is not consumer or production acceptance.

## Historical RC3 preparation

The checklist below records the earlier RC3 preparation and cutover plan. Its
candidate-status statements predate the published RC3 through RC6 Git tags;
use the current section above for RC10 actions and recheck exact tag and proxy
state before any release act.

Both first `v0.1.0-rc.1` module candidates are available through immutable
directory-prefixed tags and verified fresh public-proxy downloads from the
source in [SDK PR #9](https://github.com/eshu-hq/eshu-sdk/pull/9). The later `collector/v0.2.0-rc.1` marker and `factschema/v0.1.0-rc.2`
numeric-decoder tags are also published and immutable. The owner selected
`collector/v0.1.0-rc.2` as the preferred collector candidate on the initial
prerelease line; its tag is published. The next prepared candidates are
`collector/v0.1.0-rc.3` and `factschema/v0.1.0-rc.3`. Neither RC3 candidate is
published by this source preparation. The checklist below governs their
publication and later qualification. Verify availability from each exact Git
tag and public-proxy origin. The published collector `v0.2.0-rc.1` remains a
historical candidate.
The [factschema RC2 public-consumer proof](evidence/factschema-v0.1.0-rc.2.md)
records its proxy origin and checksum. Check the collector historical tag
separately before acting on its availability;
qualification and deployed acceptance remain separate. See the [compatibility policy](compatibility.md).

1. Confirm an independently reviewed mapping of inherited schemas, fixtures,
   public APIs and decoder bindings, and review all new SDK behavior. Before
   either RC3 tag, confirm the path mapping in the core
   [SDK compatibility matrix](https://github.com/eshu-hq/eshu/blob/main/docs/public/extend/sdk-compatibility.md).
   The standalone SDK mapping alone does not satisfy that core prerequisite.
2. Confirm separately reviewed versioned metadata for each intended tag and
   its core compatibility-mapping row. The changelogs prepare
   `collector/v0.1.0-rc.3` for scoped Azure tag and Kubernetes posture
   validation plus the native custom-JSON duplicate-member guard, and
   `factschema/v0.1.0-rc.3` for the corrected Azure tag observation fixtures.
   The guard does not freeze a caller's payload; serialization must stay stable
   through subsequent use. The factschema schema, typed decoder, and collector
   wire protocol are unchanged by these RC3 source corrections.
   Preserve the first candidates and the published historical collector
   `0.2.0-rc.1` entry. The optional
   trusted `conformance.Request.CoreFacts` API remains part of collector RC1;
   its grant and supported-pair requirements still apply. Source preparation
   is not tag approval or a consumer qualification.
3. Run standalone module CI. Bind the decision packet and separate tag approval
   to the exact reviewed and merged source commit with green CI. Do not hardcode
   a source SHA in these docs, because a docs edit changes the commit.
4. After separate approval, tag the reviewed and merged release commit with
   `collector/v0.1.0-rc.3` and `factschema/v0.1.0-rc.3` as independently
   versioned modules. Confirm the intended commit immediately before each tag;
   do not substitute this preparation branch's SHA. Preserve all published
   tags, including both first RC1s, collector `v0.1.0-rc.2` and
   `v0.2.0-rc.1`, and factschema `v0.1.0-rc.2`. Never retarget a published tag.
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
