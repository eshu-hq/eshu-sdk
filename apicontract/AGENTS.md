# API Contract SDK Agent Rules

This directory is a standalone public Go module for the Eshu API wire
media type. It must remain independent from Eshu internals, mirroring
`collector`'s `AGENTS.md`.

## Required Checks

- Read the repository root `AGENTS.md` and `README.md` before edits.
- Keep `go.mod` as a standalone module.
- Do not import `github.com/eshu-hq/eshu/go/internal/...`.
- Use only the Go standard library.
- When changing the frozen constant, update `README.md`, `doc.go`,
  `mime.go`, `contract_test.go`, and `CHANGELOG.md` together; a value
  change is a wire break and must move core and every consumer in the
  same reviewed change.
- Run `go test ./... -count=1` from this directory.
- Run `gofmt` for changed Go files and `git diff --check` from the repo root.

## Contract Rules

- `EnvelopeMIMEType` is frozen at `application/eshu.envelope+json`,
  copied from core `go/internal/query/querycontract` (`http.go`), which
  is the provenance of record. Never invent a variant spelling.
- This module publishes constants only: no rendering, storage, graph,
  queue, or supervisor logic.
