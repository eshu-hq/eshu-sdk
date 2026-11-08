# eshu-sdk docs

Welcome. These docs cover the three shared Go modules in this repo —
`collector`, `factschema`, and `wirecontract`. Start with
[Getting started](getting-started/README.md) if you just want to add a
module and make a first import work.

## Start here

| Section | What you'll find |
| --- | --- |
| [Getting started](getting-started/README.md) | Pick an exact version, `go get` it, and import it |
| [Guides](guides/README.md) | Task walkthroughs: wiring a collector, decoding facts, decoding API answers |
| [Concepts](concepts/README.md) | The three contracts and where each one stops |
| [Reference](reference/README.md) | Module table, imports, dependencies, and versioning rules |
| [Architecture](architecture/README.md) | Who imports what, and why nothing here imports Eshu core |
| [Diagrams](diagrams/README.md) | Every diagram in one place, each with a plain-text version |
| [Decisions](decisions/README.md) | Architecture decision records and how to add one |
| [Contributing](contributing/README.md) | Local setup, `make pre-push`, and what a good PR looks like |

## Detailed contract docs (kept where they are)

These pages predate this docs tree and stay at their existing paths.
They are the deep detail behind the summaries above.

| Doc | What's in it |
| --- | --- |
| [Compatibility policy](compatibility.md) | Candidate tags, checked versions, and acceptance state per module |
| [Command behaviour](command.md) | How a collector command swaps one JSON request and result with a core host |
| [Entrypoint generation](entrypoints.md) | How the SDK renders a stdio collector command from a provider handler |
| [Conformance grants](conformance-grants.md) | How trusted core fact context lets conformance checks cover reserved fact kinds |
| [Identity migration](identity-migration.md) | Where `StableID` came from in core and how it landed in `factschema/identity` |
| [Contracttest migration](contracttest-migration.md) | How the generic contract helpers moved into `factschema/contracttest` |
| [Migration evidence](sdk-migration-evidence.md) and [evidence notes](evidence) | Recorded comparisons and consumer proofs for specific candidates |
| [Provenance](provenance.md) | Which Eshu commit this code came from, file by file |
| [Release checklist](release.md) | How a tag gets cut and proved in CI |

## Per-module READMEs

- [collector README](../collector/README.md)
- [factschema README](../factschema/README.md)
- [wirecontract README](../wirecontract/README.md)
- [Entrypoint generator README](../collector/cmd/collector-entrypoints-gen/README.md)

Back to the [repo README](../README.md).
