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
		Responses     []json.RawMessage `json:"responses"`
		ExpectedFacts []struct {
			Kind          string          `json:"kind"`
			StableFactKey string          `json:"stable_fact_key"`
			Payload       json.RawMessage `json:"payload"`
		} `json:"expected_facts"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("decode NVD paging golden: %v", err)
	}
	if len(golden.Responses) != 2 || len(golden.ExpectedFacts) != 3 {
		t.Fatalf("NVD paging golden has %d responses and %d facts, want 2 and 3", len(golden.Responses), len(golden.ExpectedFacts))
	}
	for i, fact := range golden.ExpectedFacts {
		if fact.Kind == "" || len(fact.StableFactKey) != 64 || len(fact.Payload) == 0 {
			t.Fatalf("expected fact %d lacks kind, stable key, or payload", i)
		}
	}
	raw[0] = 0
	if NVDPagingGolden()[0] == 0 {
		t.Fatal("accessor returned mutable embedded fixture bytes")
	}
}
