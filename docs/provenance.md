# Source provenance

The initial relocation source is Eshu commit
`eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`. Later SDK additions are
recorded in the module changelogs; they are not part of that source snapshot.

| Source directory | Destination | Permitted source transformation |
| --- | --- | --- |
| `sdk/go/collector/` | `collector/` | Module and import prefix substitution |
| `sdk/go/factschema/` | `factschema/` | Module/import and contributor-path metadata substitution; test-only splitting and dispatch consolidation |
| `go/internal/collector/sdk/` | `collector/transport/` | Package declaration and documentation changes; three inherited public comments corrected |

The initial relocation proof counted 1,076 source files. All 730 non-Go schema and fixture data files were byte identical in the relocation specimen. The original and relocated suites discovered the same 46 collector and 144 factschema Test/Benchmark names. The two original factschema decoder switches bound the same 179 fact-kind and public-decoder pairs; the relocated test map preserves that set. These are pre-publication local proofs, not consumer or deployment evidence.

The existing in-tree implementations remain active until consumer validation
and cutover. The new paths have source prepared for `v0.1.0-rc.1`; at this
preparation, no candidate tag was cut. Check Git tags and the Go proxy for
current availability. They are new module identities, distinct from the
supported original paths at `v0.2.0`; see [compatibility](compatibility.md). Module
READMEs describe inherited contracts and prepared additions. Canonical
admission and graph ownership remain in Eshu.

The transport extraction copies five production Go files and six test files.
The four non-documentation production files and all six tests preserve the
implementation after the package rename. Three public comments now accurately
describe URL credential rejection, the inherited version marker, and caller-owned
TLS reporting; `doc.go` describes the new public helper boundary. The inherited `Version = "v0.1.0"` remains an internal marker, not a
selected module version. No Eshu core import has been redirected.

The public `collector.Request` promotes the existing host request fields without
adding wire fields. Its validator reuses the SDK claim rules. `command.Run` and
the neutral generator are new SDK tooling around that wire boundary; they do
not copy the legacy templates that open core databases or manage leases.
The request schema is a separate artifact; the existing result schema is
unchanged. Provider handlers own configuration decoding and telemetry.

The identity and generic fact-contract helpers record their pinned sources and
later core-adoption requirements in [identity migration](identity-migration.md)
and [contract-test migration](contracttest-migration.md). Temporary overlap does
not permit permanently independent canonical implementations.

The supported old modules resolve through the Go proxy at exact `v0.2.0`
versions. Against that old factschema version, the source baseline has 176 to
179 schemas: three additions, no removals, and four optional nullable field
additions. The [migration evidence](sdk-migration-evidence.md) records the tag
origins, checksums, full schema diff, and reproduction commands. This is a
source comparison, not a qualified core/SDK pair or a deployed result. The
release decision packet will bind any candidate tag to its final reviewed,
merged source commit and green CI; this provenance note does not preselect
that changing commit SHA.
