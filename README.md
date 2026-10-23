# Eshu Go SDK preparation

This repository is a local preparation of two independent public Go modules copied from Eshu at commit `eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`. Neither proposed module path has been tagged, published as a Go module, or accepted by a production Eshu consumer.

| Directory | Proposed module path | Tag prefix |
| --- | --- | --- |
| `collector/` | `github.com/eshu-hq/eshu-sdk/collector` | `collector/vX.Y.Z` |
| `factschema/` | `github.com/eshu-hq/eshu-sdk/factschema` | `factschema/vX.Y.Z` |

There is no root Go module. Each module is tested and versioned independently. No release version has been selected. The original in-tree modules and existing immutable tags remain in place until consumer tests and production cutover pass.

## Source and scope

The modules came from `sdk/go/collector` and `sdk/go/factschema` in Eshu. Go import paths and module declarations now use the proposed SDK paths. Schema and fixture bytes are unchanged. The inherited factschema tests were split below 500 lines; their public-decoder coverage remains explicit. See [provenance](docs/provenance.md) for the mapping and checks.

The reusable collector HTTP, TLS, and retry helpers remain a separate future slice in `collector/transport`. Core conformance orchestration, AWS parity, and entrypoint templates have not moved. This preparation does not grant support or release status to any new module path.

## Local checks

Run `make pre-push` for both module suites, vet, generation drift, and the file cap. `GOWORK=off` ensures each module works independently. CI runs the same checks. Contract changes require updated tests, schemas, fixtures, and a consumer compatibility review before release.

See [release and cutover](docs/release.md) for the unresolved acceptance steps. The upstream Eshu [Contract System v1](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/internal/design/contract-system-v1.md) defines the original payload contract.
