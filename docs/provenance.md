# Source provenance

The source is Eshu commit `eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`.

| Source directory | Destination | Permitted source transformation |
| --- | --- | --- |
| `sdk/go/collector/` | `collector/` | Module and import prefix substitution |
| `sdk/go/factschema/` | `factschema/` | Module/import and contributor-path metadata substitution; test-only splitting and dispatch consolidation |
| `go/internal/collector/sdk/` | `collector/transport/` | Package declaration and documentation changes; three inherited public comments corrected |

The initial relocation proof counted 1,076 source files. All 730 non-Go schema and fixture data files were byte identical in the relocation specimen. The original and relocated suites discovered the same 46 collector and 144 factschema Test/Benchmark names. The two original factschema decoder switches bound the same 179 fact-kind and public-decoder pairs; the relocated test map preserves that set. These are pre-publication local proofs, not consumer or deployment evidence.

The existing in-tree implementations remain active until consumer validation and cutover. No version or tag is assigned here. Module-specific README files describe inherited contracts and prepared additions. Canonical admission and graph ownership remain in Eshu.

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
