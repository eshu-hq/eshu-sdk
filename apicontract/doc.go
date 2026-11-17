// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

// Package apicontract publishes the Eshu API wire constants a thin client or
// gate needs, without importing Eshu internals.
//
// It carries the canonical response-envelope media type from
// go/internal/query/querycontract (http.go) in Eshu core. The value is
// frozen: changing the string breaks envelope negotiation for every client
// that sends it. The frozen contract test (contract_test.go) pins the exact
// bytes.
//
// The module is intentionally independent from Eshu internal Go packages,
// the same constraint github.com/eshu-hq/eshu-sdk/collector and
// github.com/eshu-hq/eshu-sdk/factschema already satisfy: it imports nothing
// under github.com/eshu-hq/eshu/go/internal/..., and it uses only the Go
// standard library.
package apicontract
