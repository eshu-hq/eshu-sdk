// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

// Package wirecontract publishes the Eshu API wire shapes a thin Tier-2 CLI
// needs, without importing Eshu internals.
//
// It carries the response envelope, truth, and profile shapes from
// go/internal/query/querycontract (envelope, profile, freshness) via the
// in-tree sdk/go/apicontract module (Refs eshu-hq/eshu#9509), the MCP
// tool-definition shape from go/internal/mcp/contract/tool, and the
// client-scoped environment-variable registry data plus its pure validator
// from go/internal/envregistry. Every JSON field name, tag, and enum value is
// pinned byte-for-byte to the core types at pin c606960c97 by the committed
// goldens in testdata/golden: scripts/ci/test-wirecontract-core-pin.sh
// rebuilds those bytes from the pinned core live types and from these
// shapes, so the SDK side drifts only by failing that gate (see README.md
// and AGENTS.md; core-side regeneration is core's own suite, and core drift
// arrives here as a linked follow-up, never in the same change).
//
// What it deliberately does not publish, and why:
//
//   - Envelope constructors and the capability matrix (BuildTruthEnvelope,
//     capability ceilings, profile support tables): capability truth belongs
//     to the core query owner.
//   - Freshness cause plumbing (WithFreshnessCause, the cause to next-check
//     table, timestamp formatters): the server attaches causes; a client only
//     decodes them.
//   - Retry behavior markers, rendering and composition (packet renderers,
//     report composers, bundle capture/validate, markdown reference pages):
//     those run behind core endpoints, never in a client library.
//   - Answer classification prompts (AnswerTruthClass): a prompt-surface
//     derivation, not wire truth.
//   - Server, storage, reducer, graph, queue, and supervisor surfaces: never
//     a client contract.
//
// Sync rule: a core change to any ported shape or entry must update the
// goldens and this package in a linked follow-up in eshu-hq/eshu; only the
// SDK side fails closed in this repo (see README.md and AGENTS.md).
// Registry gaps the client reads but the core registry does
// not cover yet (ESHU_SERVICE_URL, ESHU_REMOTE_TIMEOUT_SECONDS) stay out of
// ClientEntries until the registry owns them; see README.md.
package wirecontract
