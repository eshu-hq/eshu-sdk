# ADR-0001: Three independent Go modules, no core import

## Status

Accepted. Recorded in [AGENTS.md](../../AGENTS.md),
[provenance](../provenance.md), and the
[compatibility policy](../compatibility.md).

## Context

Collectors, the core reducer, and thin API clients must agree
byte-for-byte on fact and API shapes. The original modules lived inside
Eshu core (`github.com/eshu-hq/eshu/sdk/go/...`), which forced outside
code to depend on core to get the shapes. Core also keeps the
authority — admission, grants, graph writes — and that must not leak
into a shared library.

## Decision

Ship the agreement as three independent Go modules in this repo —
`collector`, `factschema`, `wirecontract` — each with its own `go.mod`,
changelog, and `<module>/vX.Y.Z` tags. No root module, no workspace,
no `replace` directives, and no import of Eshu core internals. The old
in-core paths stay supported until consumer and deployed acceptance
authorize a cutover; the new types are not drop-in replacements.

## Consequences

- Each module versions and releases on its own; consumers pin exact
  tags per the [compatibility policy](../compatibility.md).
- Every module must prove it stands alone (`GOWORK=off` in `make pre-push`).
- Two type identities exist during migration, so docs must keep saying
  which path a snippet uses. See [Reference](../reference/README.md).
