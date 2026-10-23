// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

// Package contracttest provides opt-in collector fact-shape assertions for
// tests. Helpers take *testing.T and []factschema.Envelope; they check declared
// fact-kind membership and required payload-key presence. Extra payload keys
// and present nil values are accepted. Collector-kind identity is a separate
// assertion and is not included in AssertFactShape.
//
// This package has no runtime, scanner, AWS boundary, or generated profile API.
package contracttest
