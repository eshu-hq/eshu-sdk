# Release and cutover checklist

## Releasing a module tag

CI creates and proves SDK tags; nobody pushes one by hand. A person with write
access opens Actions, chooses the `release` workflow, and runs it on `main`
(`.github/workflows/release.yml`) with:

| Input | Value |
| --- | --- |
| `module` | `collector`, `factschema`, or `wirecontract` |
| `version` | the tag version, such as `v0.1.0-rc.11`; the tag is `<module>/<version>` |
| `commit` | the full 40-hex SHA of a commit already on `main` |
| `expect_file`, `expect_sha256` | optional pair: a file inside the module (repo-relative, starting with `<module>/`) and the SHA-256 it must have |

The workflow does not decide that a release is ready. The release ruling, the
`CHANGELOG.md` entry, and green CI on the commit are its inputs; consumer
qualification and deployed acceptance stay separate acts.

1. **verify** (read-only token). The commit is an ancestor of `origin/main`.
   The `ci-complete` check run owned by `github-actions` (the required check on
   `main`) concluded success on that exact SHA; the newest run decides. The
   module's `CHANGELOG.md` at that commit has a `## [<version without v>]`
   heading. The version is major 0 or 1 and `go.mod` declares exactly
   `github.com/eshu-hq/eshu-sdk/<module>`; major 2 or later is refused until the
   proof job supports a `/vN` module path, because Go serves the unsuffixed path
   at v0 and v1 only. Leading zeros are rejected. Neither a local nor a remote tag
   `<module>/<version>` exists. The module builds with `GOWORK=off`. If `expect_file` is given, it exists at the
   commit, is not inside a nested module, and has the expected SHA-256, so a wrong
   value stops the run here, before a tag exists. Each check prints as it passes.
2. **tag** (the only job with `contents: write`, in the `release`
   environment). It creates the lightweight tag at the verified SHA through the
   git refs API, which rejects an existing ref, then reads the ref back and
   requires the same SHA as a commit object, like the rc.8 to rc.10 tags. A ref
   that already points at the verified commit is accepted unchanged, so a rerun
   of this job after a failed read-back is safe; a ref anywhere else is refused.
3. **prove** (read-only token). A fresh consumer module with an empty module
   cache resolves the tag through `proxy.golang.org` with `GOWORK=off`, no
   `replace`, no go env file, and the public checksum database, retrying while
   the proxy catches up. The log records the tag's object SHA, the proxy's zip
   and `go.mod` checksums, the origin commit the proxy reports (the job fails
   unless it equals the verified commit), the consumer's `go.sum` lines, the
   optional file hash, and a successful `go vet`, build, and run of a program
   that imports the module. The log is uploaded as the artifact
   `release-proof-<module>-<version>` (90 days) and copied to the run summary.

A tag is immutable once published: never move or delete one, and fix forward
with the next version. If verify fails, no tag exists and the run can be started
again once the cause is fixed. If tag or prove fails after the tag exists (for
example the proxy is slow), re-run only the failed jobs from the run page;
starting a new run fails at verify because the tag exists. Two limits to know:
GitHub keeps one pending run per concurrency group, so a third dispatch for the
same module while one runs can silently drop the second; and a green
`ci-complete` does not by itself show every module lane ran on that SHA (a
skipped lane counts as success), so the build in verify is the per-SHA check.
The logic lives in `scripts/ci/release-*.sh`; `make shape` runs their self-tests
and the workflow-shape checks (`test-release-verify.sh`,
`test-release-tag-prove.sh`).

### What this does and does not protect

The jobs refuse to run from any ref but `main`. That stops an accidental
dispatch of an unedited copy from a branch; it is not a security boundary,
because a branch can edit the workflow file and drop the guard. The `release`
environment does not change that: its rules gate only jobs that reference it, so
an edited branch copy can drop `environment:` and keep `contents: write`. What
the environment gives is an approval step (required reviewers) and a
deployment-branch limit for runs of this workflow as merged on `main`.

Today no tag ruleset exists and no `release` environment is configured, so
anyone with write access can already push a tag directly: this workflow adds an
audited path that verifies and proves each tag, not a new barrier. Only a tag
ruleset restricting who can create `*/v*` tags, or narrower write access, limits
who can create tags. Recommended:

- Create the `release` environment **before the first run** (Settings,
  Environments), limit its deployment branches to `main`, and optionally add
  required reviewers so the tag job waits for approval. A run that names an
  environment that does not exist creates it without protection.
- Add a tag ruleset restricting creation of `*/v*` tags if hand-pushed tags
  should be blocked; whether the Actions token needs to be a bypass actor for it
  has not been checked.

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
`factschema/v0.1.0-rc.11`. RC10 adds the optional NVD `coverage_descriptor`,
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

`factschema/v0.1.0-rc.11` is published at commit `d966966`, created by the
release workflow on 2026-10-31 ([run 41231620015](https://github.com/eshu-hq/eshu-sdk/actions/runs/41231620015)) (it adds the
fixture-pack payload `osv_severity_golden.json` and `OSVSeverityGolden()` from
core eshu-hq/eshu#9692, with no schema, stable-key, or wire change).

Prepared next candidate: `factschema/v0.1.0-rc.12` (unpublished). It carries
the optional `status` and `replaced_by` fields on `vulnerability.cve` (core
eshu-hq/eshu#9695) and the four optional NVD operator and negate fields on
`vulnerability.affected_product` (core eshu-hq/eshu#9696), both additive with no
stable-key or wire change. Merge the reviewed source, then obtain a separate
release ruling against the concrete merged commit, and name rc.12 in the
module's `CHANGELOG.md` before that commit is tagged (the section is already
there). Then run the release workflow (see Releasing a module tag) with
`module=factschema`, `version=v0.1.0-rc.12`, that commit, and
`factschema/schema/vulnerability.cve.v1.schema.json` with SHA-256
`71fd533a311e6d029853c35b08e03f9007b3e5623c238e23f95e90523c7f2c03` as `expect_file` and `expect_sha256`; its proof job is the
fresh external `GOWORK=off` public-proxy check Eshu waits for before it pins it.
Core can only build the follow-up change sets for eshu-hq/eshu#9695 and
eshu-hq/eshu#9696 after that tag is on the proxy.

## Historical RC3 preparation

The checklist below records the earlier RC3 preparation and cutover plan. Its
candidate-status statements predate the published RC3 through RC6 Git tags;
use the current section above for current release actions and recheck exact tag and proxy
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
   versioned modules, by running the release workflow once per module (see
   Releasing a module tag); it takes the commit as an input and checks it is on
   `main` with green CI. Do not substitute this preparation branch's SHA.
   Preserve all published tags, including both first RC1s, collector
   `v0.1.0-rc.2` and `v0.2.0-rc.1`, and factschema `v0.1.0-rc.2`. Never
   retarget a published tag.
5. From a fresh external consumer, resolve each exact candidate version with
   `GOWORK=off`, no `replace`, and no `@latest`. Record proxy checksums, the
   tag origin SHA, and a successful build; the workflow's prove job records
   these in its log. Candidate availability alone does
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
