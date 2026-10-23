# Source provenance

The source is Eshu commit `eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`.

| Source directory | Destination | Permitted source transformation |
| --- | --- | --- |
| `sdk/go/collector/` | `collector/` | Module and import prefix substitution |
| `sdk/go/factschema/` | `factschema/` | Module and import prefix substitution; test-only splitting and dispatch consolidation |

The relocation proof counted 1,076 source files. All 730 non-Go schema and fixture data files were byte identical in the relocation specimen. The original and relocated suites discovered the same 46 collector and 144 factschema Test/Benchmark names. The two original factschema decoder switches bound the same 179 fact-kind and public-decoder pairs; the relocated test map preserves that set. These are pre-publication local proofs, not consumer or deployment evidence.

The Go production implementation remains owned by Eshu until consumer validation and cutover. No version or tag is assigned here. Module-specific README files describe the wire contract inherited at the source commit.
