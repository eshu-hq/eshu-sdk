// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package apicontract

import (
	"strings"
	"testing"
)

// TestEnvelopeMIMETypeIsFrozen pins the canonical envelope media type
// byte-for-byte. The value is copied from Eshu core
// (go/internal/query/querycontract http.go); any change here is a wire break
// that must move core and every consumer in the same reviewed change, never
// in this module alone.
func TestEnvelopeMIMETypeIsFrozen(t *testing.T) {
	t.Parallel()
	const want = "application/eshu.envelope+json"
	if EnvelopeMIMEType != want {
		t.Fatalf("EnvelopeMIMEType = %q, want frozen %q", EnvelopeMIMEType, want)
	}
	if !strings.Contains(EnvelopeMIMEType, "/") || strings.ContainsAny(EnvelopeMIMEType, " \t\n") {
		t.Fatalf("EnvelopeMIMEType = %q, want a bare type/subtype media type", EnvelopeMIMEType)
	}
}
