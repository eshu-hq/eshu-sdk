# Contributing to eshu-sdk

There's no separate `CONTRIBUTING.md` or `SECURITY.md` in this repo, so
[AGENTS.md](../../AGENTS.md) is the contribution rulebook. This page is
the friendly version; AGENTS.md wins if they ever disagree.

## Setup

- Go 1.26.10 or newer.
- Work on a named branch in an isolated worktree. Check `pwd` before
  writing and `git symbolic-ref -q HEAD` before committing.
- Read the scoped `AGENTS.md` in each module or family before editing it.

## Before you push

```sh
make pre-push
```

That builds, tests, vets, and lints all three modules, and checks tidy
state, generation drift, and the 500-line file cap — with `GOWORK=off`
so each module proves it stands alone. For a single module, run its
`GOWORK=off go test ./... -count=1` and `go vet ./...`. For factschema
payload changes, also run `go generate ./...` and inspect the drift.

## Ground rules

- TDD: a bug fix starts with a failing regression; a new guard needs a
  seeded RED/GREEN check.
- Keep checked-in schemas and fixtures aligned with the public wire contract.
- Keep `collector/`, `factschema/`, and `wirecontract/` as separate
  modules. Don't add a root module, workspace, replacement directive,
  or a dependency on Eshu core internals.
- Use `rg` for searches. Keep every tracked file under 500 lines. No AI attribution.
- Never claim release or production acceptance without consumer and
  deployment proof. Check the [compatibility policy](../compatibility.md)
  and the [release checklist](../release.md).
