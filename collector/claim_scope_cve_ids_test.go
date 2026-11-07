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

// TestClaimScopeCVEIDsRoundTrip proves a findings claim carries its CVE id
// list across the wire: ids survive a JSON round trip. The satellite
// resolves _findings scope ids through this field (eshu-hq/eshu#9876); the
// scope id itself carries no CVE ids.
func TestClaimScopeCVEIDsRoundTrip(t *testing.T) {
	claim := Claim{
		ComponentID:   "vuln-intel",
		InstanceID:    "instance-1",
		CollectorKind: "vulnerability_intelligence",
		SourceSystem:  "first_epss",
		Scope:         Scope{ID: "vuln-intel://first_epss/_findings/0", Kind: "derived-epss"},
		SourceRunID:   "run-1",
		GenerationID:  "gen-1",
		WorkItemID:    "work-1",
		FencingToken:  "7",
		Attempt:       1,
		Deadline:      time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC),
		ConfigHandle:  "handle-1",
		ScopeCVEIDs: []string{
			"CVE-2024-0001",
			"CVE-2024-0002",
		},
	}
	raw, err := json.Marshal(claim)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"scope_cve_ids"`) {
		t.Fatalf("marshaled claim omits scope_cve_ids: %s", raw)
	}
	var decoded Claim
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.ScopeCVEIDs, claim.ScopeCVEIDs) {
		t.Fatalf("scope_cve_ids round trip mismatch: got %+v want %+v", decoded.ScopeCVEIDs, claim.ScopeCVEIDs)
	}
}

// TestClaimScopeCVEIDsEmptySerializesIdentical proves old-satellite
// compatibility: a claim without CVE ids serializes byte-identical to the
// legacy shape, so a satellite that never heard of findings transport reads
// exactly the bytes it always has. Both nil and an empty slice omit.
func TestClaimScopeCVEIDsEmptySerializesIdentical(t *testing.T) {
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
	claim.ScopeCVEIDs = []string{}
	raw, err = json.Marshal(claim)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != want {
		t.Fatalf("empty scope_cve_ids leaks onto the wire:\n got: %s\nwant: %s", raw, want)
	}
}

// TestValidateScopeCVEIDs proves the Go validator holds CVE transport to the
// published shape: at most MaxScopeCVEIDs entries, each non-blank. Shape and
// join validation against the FIRST client boundary lives satellite-side
// (epssCVEParam), mirroring how the host never validates query contents.
func TestValidateScopeCVEIDs(t *testing.T) {
	valid := Claim{
		ComponentID: "c", InstanceID: "i", CollectorKind: "k", SourceSystem: "s",
		Scope: Scope{ID: "scope", Kind: "kind"}, SourceRunID: "r", GenerationID: "g",
		WorkItemID: "w", FencingToken: "f", Attempt: 1,
		Deadline:     time.Now().Add(time.Hour),
		ConfigHandle: "h",
		ScopeCVEIDs:  []string{"CVE-2024-0001"},
	}
	if err := validateClaim(valid); err != nil {
		t.Fatalf("valid scope_cve_ids rejected: %v", err)
	}
	over := make([]string, MaxScopeCVEIDs+1)
	for i := range over {
		over[i] = "CVE-2024-0001"
	}
	valid.ScopeCVEIDs = over
	if err := validateClaim(valid); err == nil {
		t.Fatal("validator accepted scope_cve_ids over the cap")
	}
	valid.ScopeCVEIDs = []string{"CVE-2024-0001", "  "}
	if err := validateClaim(valid); err == nil {
		t.Fatal("validator accepted blank scope_cve_ids entry")
	}
}
