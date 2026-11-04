# Changelog

This file preserves the release history of the original in-tree
`github.com/eshu-hq/eshu/sdk/go/factschema` module in Eshu. The `[0.1.0]`
and `[0.2.0]` entries and tags below belong to that original path and
repository. The new `github.com/eshu-hq/eshu-sdk/factschema` path starts an
independent version line. Candidate entries describe source changes; tag
publication and production consumer support are separate states. Check Git tags and
the Go proxy for current availability. See the
[compatibility policy](../docs/compatibility.md).

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Per-fact-kind payload schemas follow the breaking-change policy in
[Contract System v1 §5](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/internal/design/contract-system-v1.md#5-versioning-and-compatibility-policy):
removing or renaming a field, narrowing a field's type, or changing a stable
key's derivation is a major bump with a decode shim in the same change; an
additive optional field is a minor bump; docs-only is a patch. The **module**
version tracks the highest bump any fact kind needed in that release. See
[SDK Compatibility](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/public/extend/sdk-compatibility.md) for how
this module version lines up with core Eshu releases and fixture-pack pins.

Tags for the new module use the Go subdirectory-module format after separate release approval:
`factschema/vX.Y.Z`. The fixture pack
(`factschema/fixturepack`) ships inside this module and has no separate
version — pinning this module pins the fixture pack too.

## [Unreleased]

## [0.1.0-rc.17] - Prepared (unpublished)

This candidate source is not tagged. Publishing it needs a separate release
ruling against the merged commit, and no proxy verification is claimed for it.
It adds five optional fields to one existing kind, so the module bump is a
minor step within the 0.1 line (the counter moves from rc.16 to rc.17). No
required field, existing field, stable key, schema version, or wire format
changes; the generated schema only gains optional properties. File digests
(SHA-256) at this source:
`schema/vulnerability.go_call_reachability.v1.schema.json`
`9470f63b617a4650dcfe9dd73f65bf8cc794d204fd62e52c3b736397e55fba7d`,
`fixturepack/payloads/vulnerability.go_call_reachability.valid.json`
`cbcc32e16ab71db090acb25cd720f6c7d71cc52297a6513a6069b9ddadbdf48e`.

### Added

- `vulnerability.go_call_reachability` gains five optional provenance fields,
  `scan_level`, `scan_mode`, `scanner_version`, `source_commit_sha`, and
  `trace_count` (core eshu-hq/eshu#10026, part of eshu-hq/eshu#10002; additive
  minor, no decode shim). `scan_level` (`module`, `package`, or `symbol`) says
  how deep the scan went, so a reader can tell a `symbol` scan, the only kind
  that can show no call path exists, from a shallower one. `scan_mode` is
  `source` or `binary`. `scanner_version` names the scanner release,
  `source_commit_sha` the commit that was scanned, and `trace_count` how many
  traces the scanner found; `trace` carries one of them. `scan_level` and
  `scan_mode` are open strings: the schema does not restrict them and the
  decode does not check them, so a reader keeps a value it does not know. A
  payload without the keys decodes as before, with nil fields, and a value of
  the wrong JSON type now fails the decode as `input_invalid`.
- The struct documentation now states Eshu's trace order: `trace` runs entry
  point first and vulnerable symbol last, and `deepest_*` describes the last
  frame. This is the opposite of govulncheck's own order, which lists the
  vulnerable symbol first, so a producer reading govulncheck output reverses it.
  `position` stays one string rendered `filename:line:column`; govulncheck
  reports an object, so a producer renders it and drops the offset. No bytes on
  the wire change meaning: the order was already entry first in every recorded
  fact.
- The type documentation states what a producer writes for each outcome
  (`symbol`, `import`, `module`, `not_called`, and a partial scan) and ties
  `reachability_level` to `scan_level`: `import` and `module` are for package
  and module scans, and a symbol scan's package-only or module-only finding is
  `not_called`, because the reducer reads `import` as reachable. Every finding that
  names a module sets `deepest_module`, including a `not_called` finding with
  an empty `trace`: the core reducer joins on it, and a fact without it reads
  as "no govulncheck evidence loaded". The generated schema is unchanged.
- The fixture pack's `vulnerability.go_call_reachability.valid.json` carries the
  five new fields, a frame `version`, and string positions, so a conformance run
  validates them against the schema. The invalid payload is unchanged.

## [0.1.0-rc.16] - 2026-11-04

This tag is published. The `release` workflow ([run 41860075638](https://github.com/eshu-hq/eshu-sdk/actions/runs/41860075638))
verified commit `95bae1a95dd5080cb90fc3a120d99c18af872607`, created the tag
`factschema/v0.1.0-rc.16` at it on 2026-11-04, and the run concluded success.
It changes one fixture-pack payload and adds no field, so the module bump is a
minor step within the 0.1 line (the counter moves from rc.15 to rc.16). No
payload schema changes. File digest (SHA-256) at this source:
`fixturepack/payloads/source_terms_golden.json`
`c8a611ffb9afd154c3ecc3377ae73d073d1394086c3fdc569fa4470ede8ec6a4`.

### Added

- The shared `source_terms_golden.json` gains a `ghsa` row (core
  eshu-hq/eshu#9959, Phase 4 item 4.2 of eshu-hq/eshu#9209): license
  `CC-BY-4.0`, with an attribution line that credits the GitHub Advisory
  Database, links `https://github.com/advisories` (the link GitHub's terms say
  fulfils the attribution term), names the license and its URL, and says Eshu
  normalizes the records. A consumer that replays the golden through its
  source table must add the `ghsa` row to pass, so a consumer pinned to rc.15
  keeps the old bytes and the old expectation.

## [0.1.0-rc.15] - 2026-11-04

This tag is published. The `release` workflow ([run 41852031597](https://github.com/eshu-hq/eshu-sdk/actions/runs/41852031597))
verified commit `10f6c1772aa1f6f15b3ced4d26f2a7a59ebc4818`, created the tag
`factschema/v0.1.0-rc.15` at it on 2026-11-04, and ran its public-proxy proof,
which succeeded. The public proxy lists the version and reports that commit as
its origin; module zip checksum `h1:lJ99lFqDQtZd5YfCBLRJC2up31A9SDBzVaGr1qvKes4=`.
It carries one additive change, four optional fields on an existing kind plus
one fixture-pack payload, so the module bump is a minor step within the 0.1
line (the release candidate counter moves from rc.14 to rc.15). No required
field, existing field, stable key, schema version, or wire format changes; the
generated schema only gains optional properties. Schema file digest (SHA-256)
at this source: `schema/vulnerability.source_snapshot.v1.schema.json`
`17905bba18014b479b61b3709747130877842a8d3f41ca5336b2c65ad8eeeead`.

### Added

- `vulnerability.source_snapshot` gains four optional fields, `license`,
  `attribution`, `cache_expires_at`, and `cache_mode` (core eshu-hq/eshu#9958,
  Phase 4 item 4.0 of eshu-hq/eshu#9209, additive minor, no decode shim).
  `license` names the terms the source's data was collected under: an SPDX
  identifier such as `CC0-1.0`, a `LicenseRef-` identifier where the source
  publishes terms with no SPDX identifier, or `NOASSERTION` where no stated
  license was found. It describes the source, not each record, and it is a
  disclosure carried with the snapshot, not an access decision; no reader
  enforces it. `attribution` is one plain-text credit line. `cache_expires_at`
  and `cache_mode` were already written by the cache-wrapping layer but absent
  from the typed struct. A snapshot without the keys decodes as before, with nil
  fields, and a value of the wrong JSON type now fails the decode as
  `input_invalid`.
- The fixture pack gains `payloads/source_terms_golden.json` and
  `SourceTermsGolden()`: the per-source `license` and `attribution` that the
  `osv`, `nvd`, `cisa_kev`, and `first_epss` snapshots carry (SHA-256
  `2c60495311714fcac206c7df467657c8d7e0430370a4cd21554374046ee5b335`). Core and
  the extracted collector each replay it against their own snapshot builder.
  The NVD paging golden's expected `source_snapshot` payload gains the same two
  keys, so its digest moves to
  `2c7c83dcfd8905ae7eddb753268fcbb0485e2cbda37839659e9ceacb3233c76a`; a
  consumer pinned to rc.14 keeps the old bytes and the old expectation.

## [0.1.0-rc.14] - 2026-11-04

This tag is published. The `release` workflow ([run 41750345061](https://github.com/eshu-hq/eshu-sdk/actions/runs/41750345061))
verified commit `dd82190ef9192564c410c32892e3b2caf0751d27`, created the tag
`factschema/v0.1.0-rc.14` at it on 2026-11-04, and ran its public-proxy proof.
The first two proof attempts ran before `proxy.golang.org` had indexed the tag
and failed with a 404; a re-run of the proof job after the proxy served the
version succeeded, and the run concluded success. The public proxy lists the
version and reports that commit as its origin.
It carries two additive changes: three optional fields on an existing kind and
one fixture-pack payload, so the module bump is a minor step within the 0.1
line (the release candidate counter moves from rc.13 to rc.14). No required
field, existing field, stable key, schema version, or wire format changes; the
generated schema only gains optional properties. File digests (SHA-256) at this
source:
`schema/vulnerability.cve.v1.schema.json` `5bc338252fd10e126ad4aef9e0de7305fd50c6f2984565c76e2dd3edd8ef8941`,
`fixturepack/payloads/osv_identity_edges_golden.json` `3df283f40ecbe08dcd559bd3e3500c865411d247d8e644383c52fda8b3266010`.

### Added

- `vulnerability.cve` gains three optional raw identity-edge fields, `aliases`,
  `related`, and `upstream` (core eshu-hq/eshu#9954, Phase 3 item 3.1 of
  eshu-hq/eshu#9209, additive minor, no decode shim). Each is a string array
  copied from the OSV record field of the same name; the collector trims each
  value, drops blanks and duplicates, and sorts the list, and writes the edges
  without interpreting them. An absent key, a null, and an empty list read
  alike. `aliases` is a symmetric identity edge: the ids name the same
  advisory, and it is the only edge a reducer may group advisories through.
  `related` is display-only and never groups, keys, or enriches. `upstream` is
  directed: this distro or downstream record names the upstream CVEs it
  repackages, and a reducer may pass CVE enrichment down to this record, never
  up. `upstream` is never an alias or a correlation anchor; folding it into the
  symmetric set over-merged advisory groups in osv.dev (core
  `docs/internal/design/advisory-intelligence-correlation-plan.md` section 4.2).
  A fact without the keys decodes as before, with nil fields, and an unknown
  extra key is still ignored. A value of the wrong JSON type (a string, a
  number, an object, or a non-string element) now fails the decode and
  quarantines the fact as `input_invalid`, where the extra key used to be
  ignored; a conforming collector emits a string array or null. An rc.13 reader
  ignores the keys, so a producer may emit them before every reader reads them.

- Fixture-pack payload `payloads/osv_identity_edges_golden.json` and the
  accessor `OSVIdentityEdgesGolden()` (core eshu-hq/eshu#9954), byte-identical
  to the core copy. Seven synthetic OSV records and the `aliases`, `related`,
  `upstream`, and `status` a conforming OSV collector writes for each:
  `aliases` and `related` are always written, an empty list as JSON `null`;
  `upstream` is written only when it has an id; `status` is `withdrawn` only for
  a record with a non-blank `withdrawn` timestamp. A DSA bundle of CVEs is
  `upstream`, not `aliases`. The pack test derives every expected field from
  the case record and pins the file digest.

## [0.1.0-rc.13] - 2026-11-03

This tag is published. The `release` workflow ([run 41726236370](https://github.com/eshu-hq/eshu-sdk/actions/runs/41726236370))
verified commit `9e57ec4a90c0a76d86b0cb84fc94ea455178a935`, created the tag
`factschema/v0.1.0-rc.13` at it on 2026-11-03, and ran its public-proxy proof;
the run concluded success. The public proxy lists the version and reports that
commit as its origin.
It carries three additive changes, each an optional field on an existing kind,
so the module bump is a minor step within the 0.1 line (the release candidate
counter moves from rc.12 to rc.13). No required field, existing field, stable
key, or wire format changes; the generated schemas only gain optional
properties. Schema file digests (SHA-256) at this source:
`schema/vulnerability.affected_package.v1.schema.json` `fbe6f43e156e6db8e40e7fdc960584f3023136d2c90e1061244fc62a95da99b6`,
`schema/vulnerability.os_package.v1.schema.json` `0d0dd4eae5d6a24b09ac795be698a31d1455faad601bd47d22afea9348bdf8de`,
`schema/reducer_supply_chain_impact_finding.v1.schema.json` `bd3af73b4ee64eea5feca757caef14c753ad182c633d398ed91e88eb3ecfdfcb`.

### Added

- `vulnerability.affected_package` gains an optional `releases` field on each
  `AffectedRange` (core eshu-hq/eshu#9806, additive minor, no decode shim). It
  names the operating-system release lines a range applies to, so a reducer can
  match an installed OS package against the range for its own release instead of
  every release at once. Each element is `<vendor>:<line>` in lowercase, for
  example `debian:12`, `alpine:3.18`, `ubuntu:22.04`; the list is sorted and has
  no duplicates. The line comes from the source ecosystem string by one closed
  rule: Debian uses the major version, Alpine and Ubuntu use major.minor, a
  leading `Pro` segment, a trailing `LTS` segment, and a leading `v` are
  dropped, and a non-numeric ecosystem suffix gives no tag. The field is omitted
  when the producer did not tag the range (an empty list means the same), and a
  reader treats an omitted value as
  "release unknown", never "every release". A fact is tagged on every ECOSYSTEM
  range or on none. A fact without the key decodes as before, with a nil
  `Releases`, and an unknown extra key on a range is still tolerated. A value of
  the wrong JSON type (a string for `releases`, a non-string element) now fails
  the decode and quarantines the fact as `input_invalid`, where the extra key
  used to be ignored; a conforming producer always emits a string array. An rc.12
  reader ignores the key because the range item keeps `additionalProperties:
  true`, so a producer may emit `releases` before every reader reads it.

- `vulnerability.os_package` gains an optional `source_package` field (core
  eshu-hq/eshu#9859, additive minor, no decode shim). It names the source
  package the installed binary was built from (`openssl` for the dpkg row
  `libssl3`), verbatim from the installed database's Source field. Only
  dpkg-family rows carry one today; a fact without the key decodes as before,
  with a nil `SourcePackage`, and `name` stays the installed binary key, never
  rewritten. A value of the wrong JSON type now fails the decode and quarantines
  the fact as `input_invalid`, where the extra key used to be ignored. There is
  deliberately no `source_version` field: impact is decided by
  vendor-advisory-source matching only, never by version comparison. An rc.12
  reader ignores the key, so a producer may emit `source_package` before every
  reader reads it.

- `reducer_supply_chain_impact_finding` gains three optional evidence fields
  (core eshu-hq/eshu#9859, additive minor, no decode shim):
  `os_package_match_derivation`, `installed_package_name`, and
  `source_package`. They record which installed-row key an OS advisory finding
  joined through: `os_package_match_derivation` is `binary_package` or
  `source_package`, `installed_package_name` is the binary name joined on, and
  `source_package` is the source name joined through (set only on a
  source-joined row). A finding written before them decodes with nil fields, and
  nil fields stay off the wire. A value of the wrong JSON type now fails the
  decode and quarantines the fact as `input_invalid`.

## [0.1.0-rc.12] - 2026-10-31

This tag is published. The `release` workflow ([run 41250267912](https://github.com/eshu-hq/eshu-sdk/actions/runs/41250267912))
verified commit `7bfa7868926d0af0f0aae3a616889199790b126e` and ran its public-proxy proof; the run
concluded success. The public proxy lists the version and reports that commit
as its origin.
It carries two additive changes, both optional fields on existing kinds, so the
module bump is a minor step within the 0.1 line (the release candidate counter
moves from rc.11 to rc.12). No required field, existing field, stable key, or
wire format changes in either; the two generated schemas only gain optional
properties. Schema file digests (SHA-256) at this source:
`schema/vulnerability.cve.v1.schema.json` `71fd533a311e6d029853c35b08e03f9007b3e5623c238e23f95e90523c7f2c03`,
`schema/vulnerability.affected_product.v1.schema.json` `76956466c7de1cb783dd49927efde78066a6e0a3c36f9a363e548da7bc8e4450`.

### Added

- `vulnerability.cve` gains two optional fields, `status` and `replaced_by`
  (issue #9695, additive minor). `status` is the source-reported lifecycle
  status: the NVD collector writes the CVE API 2.0 `vulnStatus` verbatim, and
  the OSV collector writes `withdrawn` on a record whose `withdrawn` timestamp
  is set. A reducer compares it case-insensitively and treats only `rejected`
  and `withdrawn` as the source retracting the advisory; any other value, new
  or unknown, retracts nothing. `replaced_by` lists the CVE ids a source names
  as the replacement, in source order. It is filled only from structured source
  data (the CVE Record Format `cnaRejectedContainer.replacedBy`); the NVD API 2.0 and
  OSV formats have no such field, so their collectors never write it. A fact
  without either key decodes as before. A value of the wrong JSON type (a
  number for `status`, a string for `replaced_by`) now fails the decode and
  quarantines the fact as `input_invalid`, where the extra key used to be
  ignored; the collectors always emit a string and a string array. Ships in the
  same tag as the `vulnerability.affected_product` flags below.

- `vulnerability.affected_product` gains four optional fields the NVD
  collector already emitted: `source_configuration_operator`,
  `source_configuration_negate`, `source_node_operator`, and
  `source_node_negate` (issue #9696, additive minor). A nil negate flag means
  an older fact that did not carry it. Facts carry no node id, so a reducer can
  use the negate flags only to refuse an affected result; it cannot rebuild AND
  groups from the operators. A flag of the wrong JSON type (for example the
  string `"true"`) now fails the decode and quarantines the fact as
  `input_invalid`, where the extra key used to be ignored; the NVD collector
  always emits a bool and a string. Ships in the same tag as the
  `vulnerability.cve` status fields above.

## [0.1.0-rc.11] - 2026-10-31

This tag is published; it does not qualify a core consumer or deployed pair. The
`release` workflow ([run 41231620015](https://github.com/eshu-hq/eshu-sdk/actions/runs/41231620015)) verified commit
`d966966bbbda6b2253ab56c857050cbf64f58b8b`, created the tag
`factschema/v0.1.0-rc.11` at it on 2026-10-31, and ran its public-proxy proof;
the run concluded success. The public proxy lists the version and reports that
commit as its origin.

### Added

- Fixture-pack payload `payloads/osv_severity_golden.json` and the accessor
  `fixturepack.OSVSeverityGolden()`: twelve synthetic OSV records and the
  `cvss_vector` and `severity_label` a conforming OSV collector writes on
  `vulnerability.cve` for each. Core and the extracted collector replay the same
  file, so they read one OSV severity rule instead of two copies. The validator
  derives every expected field from its record. This is additive: no schema
  file, stable key, or wire format changes, and `vulnerability.cve` v1 already
  declares both fields. Ported from core eshu-hq/eshu#9692.

## [0.1.0-rc.10] - 2026-10-28

This tag is published and proxy-verified; it does not qualify a core consumer
or deployed pair. See the
[RC10 public-consumer proof](../docs/evidence/factschema-v0.1.0-rc.10.md).

### Added

- `vulnerability.source_snapshot` gains an optional `coverage_descriptor` for
  NVD: a versioned, normalized record of the explicit CVE set (trimmed, upper
  case, deduplicated, sorted), query mode and window, configured endpoint set,
  and endpoint actually used. Older snapshots remain valid source observations
  but cannot certify zero advisory coverage. Optional `collector_instance_id`
  and `result_count` fields now declare fields the core SQL readiness reader
  already uses. Additive optional fields are a minor change: `required` stays
  `source` and `complete`, and no stable key changes. Ported from core
  eshu-hq/eshu#9435 (core eshu-hq/eshu#9227); the canonical schema and the
  fixture-pack copy were regenerated in lockstep and match the core in-tree
  schema byte for byte.

## Older entries

Entries from `0.1.0-rc.9` back to `0.1.0` are in [CHANGELOG-archive.md](CHANGELOG-archive.md).
