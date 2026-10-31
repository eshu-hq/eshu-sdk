// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestClaimScopeQueriesRoundTrip proves a slot claim carries its per-scope
// query list across the wire: name and version survive a JSON round trip.
// The satellite resolves _slot scope ids through this field (eshu-hq/eshu#9432);
// the scope id itself carries no queries.
func TestClaimScopeQueriesRoundTrip(t *testing.T) {
	claim := Claim{
		ComponentID:   "vuln-intel",
		InstanceID:    "instance-1",
		CollectorKind: "vulnerability_intelligence",
		SourceSystem:  "osv",
		Scope:         Scope{ID: "vuln-intel://osv/npm/_slot/3/0", Kind: "derived-osv"},
		SourceRunID:   "run-1",
		GenerationID:  "gen-1",
		WorkItemID:    "work-1",
		FencingToken:  "7",
		Attempt:       1,
		Deadline:      time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
		ConfigHandle:  "handle-1",
		ScopeQueries: []ScopeQuery{
			{Ecosystem: "npm", Package: "left-pad", Version: "1.3.0"},
			{Ecosystem: "npm", Package: "is-odd", Version: "3.0.1"},
		},
	}
	raw, err := json.Marshal(claim)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"scope_queries"`) {
		t.Fatalf("marshaled claim omits scope_queries: %s", raw)
	}
	var decoded Claim
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.ScopeQueries, claim.ScopeQueries) {
		t.Fatalf("scope_queries round trip mismatch: got %+v want %+v", decoded.ScopeQueries, claim.ScopeQueries)
	}
}

// TestClaimScopeQueriesEmptySerializesIdentical proves old-satellite
// compatibility: a claim without scope queries serializes byte-identical to
// the legacy shape, so a satellite that never heard of slot transport reads
// exactly the bytes it always has.
func TestClaimScopeQueriesEmptySerializesIdentical(t *testing.T) {
	claim := Claim{
		ComponentID:   "c",
		InstanceID:    "i",
		CollectorKind: "k",
		SourceSystem:  "s",
		Scope:         Scope{ID: "scope", Kind: "kind"},
		SourceRunID:   "r",
		GenerationID:  "g",
		WorkItemID:    "w",
		FencingToken:  "f",
		Attempt:       1,
		Deadline:      time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		ConfigHandle:  "h",
	}
	raw, err := json.Marshal(claim)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"component_id":"c","instance_id":"i","collector_kind":"k","source_system":"s",` +
		`"scope":{"id":"scope","kind":"kind"},"source_run_id":"r","generation_id":"g",` +
		`"work_item_id":"w","fencing_token":"f","attempt":1,` +
		`"deadline":"2026-01-02T03:04:05Z","config_handle":"h"}`
	if string(raw) != want {
		t.Fatalf("legacy claim shape drift:\n got: %s\nwant: %s", raw, want)
	}
}
