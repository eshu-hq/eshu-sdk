// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

// Package transport provides reusable HTTP, TLS, retry, and bounded failure
// helpers for collector authors.
//
// The helpers execute requests and classify provider failures. Callers own
// source-specific endpoint traversal, redaction, fact emission, workflow
// decisions, and telemetry. This package does not import Eshu core internals.
package transport
