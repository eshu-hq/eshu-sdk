# SDK repository agent rules

Read this file and the scoped `AGENTS.md` in each module or family before editing. This repository contains two independently versioned Go modules. Check the [compatibility policy](docs/compatibility.md) for current candidate availability and required qualification and deployment evidence; candidate publication alone does not establish consumer or production acceptance.

- Work on a named branch in an isolated worktree. Check `pwd` before writing and `git symbolic-ref -q HEAD` before committing.
- Use `rg` for text searches and `rg --files` for file discovery. Keep every tracked file under 500 lines. Do not add AI attribution.
- Use TDD: a bug fix starts with a failing regression; a new guard needs a seeded RED/GREEN check. Keep the checked-in schemas and fixtures aligned with the public wire contract.
- Keep `collector/` and `factschema/` as separate Go modules. Do not add a root module, workspace, replacement directive, or dependency on Eshu core internals.
- Run the affected module's `GOWORK=off go test ./... -count=1` and `go vet ./...`. For factschema payload changes, run `go generate ./...` and inspect drift.
- Review source, tests, docs, secrets, and verification before handoff. Do not claim release or production acceptance without consumer and deployment proof.

The [provenance note](docs/provenance.md) records the source commit and directory mapping. The [release note](docs/release.md) lists acceptance work that remains.
