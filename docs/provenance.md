# Source provenance

The source is Eshu commit `eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`.

| Source directory | Destination | Permitted source transformation |
| --- | --- | --- |
| `sdk/go/collector/` | `collector/` | Module and import prefix substitution |
| `sdk/go/factschema/` | `factschema/` | Module/import and contributor-path metadata substitution; test-only splitting and dispatch consolidation |
| `go/internal/collector/sdk/` | `collector/transport/` | `package sdk` to `package transport`; package documentation and destination guidance updated |

The relocation proof counted 1,076 source files. All 730 non-Go schema and fixture data files were byte identical in the relocation specimen. The original and relocated suites discovered the same 46 collector and 144 factschema Test/Benchmark names. The two original factschema decoder switches bound the same 179 fact-kind and public-decoder pairs; the relocated test map preserves that set. These are pre-publication local proofs, not consumer or deployment evidence.

The Go production implementation remains owned by Eshu until consumer validation and cutover. No version or tag is assigned here. Module-specific README files describe the wire contract inherited at the source commit.

The transport extraction copies five production Go files and six test files.
The four non-documentation production files and all six tests differ from the
source only in the package clause; `doc.go` describes the new public helper
boundary. The inherited `Version = "v0.1.0"` remains an internal marker, not a
selected module version. No Eshu core import has been redirected.
