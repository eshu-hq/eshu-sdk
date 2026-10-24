// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package factschema

import (
	"testing"

	"github.com/eshu-hq/eshu-sdk/factschema/fixturepack"
)

func TestAzureTagFixtureUsesDeterministicMarker(t *testing.T) {
	payload, ok := fixturepack.ValidPayload(FactKindAzureTagObservation)
	if !ok {
		t.Fatal("azure tag observation valid fixture is missing")
	}
	fingerprints, ok := payload["tag_value_fingerprints"].(map[string]any)
	if !ok {
		t.Fatal("tag_value_fingerprints is not an object")
	}
	const want = "redacted:hmac-sha256:c78a3aed0e9e0e9662dc19a40eddcae8440730c99950d67b15453e8affde39a8"
	if got := fingerprints["env"]; got != want {
		t.Fatalf("synthetic env fingerprint = %v, want %s", got, want)
	}
}
