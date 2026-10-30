# Wirecontract SDK Agent Rules

This directory is a standalone public Go module for Eshu API wire
contracts. It must remain independent from Eshu internals, mirroring
`collector` and `factschema` `AGENTS.md`.

## Required Checks

- Read the repository root `AGENTS.md` and `README.md` before edits.
- Keep `go.mod` as a standalone module with no require or replace directives.
- Do not import `github.com/eshu-hq/eshu/go/internal/...`.
- Use only the Go standard library outside tests.
- When changing a shape or entry, update `README.md`, `doc.go`,
  `testdata/golden/*.json`, and the core drift tests in the same change.
- Run `go test ./... -count=1` from this directory.
- Run `gofmt` for changed Go files and `git diff --check` from the repo root.

## Contract Rules

- Every JSON field name, tag, and enum value must stay byte-identical to the
  core source type. The golden files are the machine-checked proof.
- Shapes only: no rendering, composition, storage, reducer, graph, queue, or
  supervisor logic. Pure validation and parsing helpers are allowed.
- A missing registry name in `ClientEntries` fails the core drift test; add
  the name to the test list, the table, and the golden together.
- New files need the MIT SPDX header; copy the two header lines from a
  sibling file before the first commit that holds the file.
