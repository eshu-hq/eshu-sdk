# API Wire-Contract Module

This module publishes the Eshu API wire shapes a thin Tier-2 CLI needs:
the `{data, truth, error}` response envelope, truth and profile contracts,
the MCP tool-definition shape, and the client-scoped environment-variable
registry with its validator. It intentionally does not import
`github.com/eshu-hq/eshu/go/internal` packages.

Ported from Eshu core `sdk/go/apicontract` (Refs eshu-hq/eshu#9509, part of
#6707), which itself ports `go/internal/query/querycontract` (envelope,
truth, profile, freshness), `go/internal/mcp/contract/tool` (tool
definition), and `go/internal/envregistry` (client entries, validator).

## Compatibility

- Go module path: `github.com/eshu-hq/eshu-sdk/wirecontract`
- Tag prefix: `wirecontract/vX.Y.Z`
- Wire sources: core `querycontract`, `toolcontract`, `envregistry` via
  core `sdk/go/apicontract` at the port commit
- Golden pins: `testdata/golden/*.json`, byte-identical to the core goldens;
  the in-module parity tests compare them against these shapes

The module version moves only when a ported shape or entry moves. A core
change to any source shape must update the goldens and this module in the
same change; both directions fail closed without it.

There is no root Go module. Each module is tested and versioned
independently.

## Contracts

The published surface includes:

- `ResponseEnvelope`, `TruthEnvelope`, `ErrorEnvelope`, `ErrorCode`, and the
  truth-level/basis/freshness/profile/backend vocabularies — byte-identical
  JSON to the core query contract.
- `QueryProfile` and `GraphBackend` with their pure parse/normalize helpers.
- `FreshnessCause`, `FreshnessNextCheck`, and `ValidFreshnessCause` for
  decoding freshness metadata.
- `ToolDefinition` — the advertised MCP tool shape, decoded from `tools/list`.
- `Entry`, `Finding`, `Registry`, and `Validate` — the client-scoped
  environment registry (`ClientEntries`) with the same pure validation the
  core runs, plus `New`, `Lookup`, `Covers`, and `Entries`.

## What stays out

- Envelope constructors, the capability matrix, and profile ceilings: the
  core query owner admits capabilities, not the client.
- Freshness cause plumbing (cause to next-check table, timestamp formatters):
  the server attaches causes; a client only decodes them.
- Retry markers, packet renderers, report composers, bundle capture/validate,
  and markdown reference pages: those run behind core endpoints, never in a
  client library (`config validate --reference` keeps rendering in core; the
  extracted CLI formats its own reference output at population time).
- Answer classification prompts, server/storage/reducer/graph/queue handles:
  never a client contract.

## Client table scope

`ClientEntries` covers the `ESHU_*` names a thin CLI reads or writes outside
local-owner paths (API auth, MCP setup, query profile, config home and
backend selection). Server, collector, reducer, and local-supervision
variables stay core-owned. Two names the thin CLI reads are not covered
because the core registry does not own them yet — `ESHU_SERVICE_URL`
(remote base URL) and `ESHU_REMOTE_TIMEOUT_SECONDS` (remote timeout).
Registering them is registry-owner follow-up; this module syncs them once
the registry owns them.

## Snippet scope decision

`mcp setup` snippet generation needs no published shape and no remote call:
it renders the config file from the setup request and platform alone
(core `go/internal/cli/mcpsetup/snippet.go` takes no tool definitions). The
`--verify` tools-visible stage needs only the tool count, which a thin client
serves from a remote `tools/list` response decoded into `ToolDefinition`
instead of the in-process catalog. The shape is published for that decode;
no local catalog import travels with it.

Evidence: core `runMCPSetup` builds snippets offline from the static
platform table plus `SetupRequest` (mode, posture, service URL) with no
network and no `tools/list`; `RunVerification` reuses the embedded
`mcp.ReadOnlyTools` catalog in-process for the count, while the hosted
first-query probe uses the API client (`GET /health`, `GET
/api/v0/index-status`) only when the snippet posture holds a credential.
Remote `tools/list` requires a running MCP server with the caller's
credential chain (API key, scoped tokens, or OIDC resource URI; headerless
requests to non-public routes are rejected 401), and its advertisement
varies per deployment via `ESHU_MCP_TOOL_SURFACE` and `ESHU_QUERY_PROFILE`.
Recommendation: keep offline snippet rendering on the published
`ToolDefinition` shape for decoding; do not replace it with a live
`tools/list` call. Use remote `tools/list` for verification counts only.
See Box-4 note in `docs/compatibility.md`.
