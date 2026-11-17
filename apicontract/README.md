# API Contract SDK

This module publishes the Eshu API wire constants an out-of-tree gate or
thin client needs. It intentionally does not import
`github.com/eshu-hq/eshu/go/internal` packages. Core Eshu owns the
envelope negotiation; this module carries a frozen copy of the media type
so consumers can negotiate it without depending on core internals.

## Compatibility

- Go module path: `github.com/eshu-hq/eshu-sdk/apicontract`
- Release version: not yet selected for this new module
- Tag format after separate release approval: `apicontract/vX.Y.Z`

The module version may add constants without changing the wire contract.
Changing the `EnvelopeMIMEType` string itself is a wire break: it must
move core and every consumer in the same reviewed change, never here
alone.

## Contracts

The public surface is currently one frozen constant:

- `EnvelopeMIMEType` (`application/eshu.envelope+json`) — the media type
  a client sends in `Accept` (and the server honors) to select the stable
  Eshu response envelope. Provenance: Eshu core
  `go/internal/query/querycontract` (`http.go`).

The value is pinned byte-for-byte by `contract_test.go`.
