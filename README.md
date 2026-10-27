# Eshu Go SDK preparation

This repository prepares two independent public Go modules initially copied
from Eshu at commit `eb62fa3056f5ed3518bb5aa676bf2e3adc96c1d5`. The
checked Git tags reach `collector/v0.1.0-rc.3` and
`factschema/v0.1.0-rc.5`. This branch prepares an optional EPSS field for a
separately reviewed factschema RC6 publication. Check each candidate's tag and
fresh public-proxy origin before use. Publication alone does not establish
acceptance by a production Eshu consumer.

| Directory | Module path | Tag prefix |
| --- | --- | --- |
| `collector/` | `github.com/eshu-hq/eshu-sdk/collector` | `collector/vX.Y.Z` |
| `factschema/` | `github.com/eshu-hq/eshu-sdk/factschema` | `factschema/vX.Y.Z` |

There is no root Go module. Each module is tested and versioned independently. Their first candidate tags are `collector/v0.1.0-rc.1` and `factschema/v0.1.0-rc.1`. The original in-tree modules and existing immutable tags remain in place until consumer tests and production cutover pass. See the [compatibility policy](docs/compatibility.md) for the old/new paths and acceptance states.

## Source and scope

The modules came from `sdk/go/collector` and `sdk/go/factschema` in Eshu. Go import paths and module declarations now use the proposed SDK paths. Inherited schema and fixture bytes are unchanged; the public request schema and fixture are added separately. The inherited factschema tests were split below 500 lines; their public-decoder coverage remains explicit. See [provenance](docs/provenance.md) for the mapping and checks.

The collector module also provides `transport` helpers, the public `Request`,
a bounded `command.Run`, and a neutral `entrypoints` generator. The fact-schema
module includes stable `identity` derivation and generic `contracttest`
assertions. See [command behavior](docs/command.md) and
[entrypoint generation](docs/entrypoints.md).

Core call sites, full conformance orchestration, AWS scanner checks, and legacy
hosted entrypoint templates remain in Eshu. The neutral generator gives
collectors a stdio command without moving database or workflow ownership.
Published candidates do not grant production support to a new module path.

## Local checks

Run `make pre-push` for both module suites, vet, generation drift, and the file cap. `GOWORK=off` ensures each module works independently. CI runs the same checks. Contract changes require updated tests, schemas, fixtures, and a consumer compatibility review before release.

See [release and cutover](docs/release.md) for the unresolved acceptance steps. The upstream Eshu [Contract System v1](https://github.com/eshu-hq/eshu/blob/eb62fa3056/docs/internal/design/contract-system-v1.md) defines the original payload contract.
