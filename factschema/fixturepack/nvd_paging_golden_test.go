// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package fixturepack

import (
	"encoding/json"
	"testing"
)

func TestNVDPagingGoldenContainsInputAndExpectedPayloads(t *testing.T) {
	raw := NVDPagingGolden()
	var golden struct {
		Responses        []json.RawMessage `json:"responses"`
		ExpectedPayloads []json.RawMessage `json:"expected_payloads"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("decode NVD paging golden: %v", err)
	}
	if len(golden.Responses) != 2 || len(golden.ExpectedPayloads) != 3 {
		t.Fatalf("NVD paging golden has %d responses and %d payloads, want 2 and 3", len(golden.Responses), len(golden.ExpectedPayloads))
	}
}
