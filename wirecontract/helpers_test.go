// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package wirecontract

import (
	"testing"
)

// TestParseQueryProfileVectors pins the pure profile parser the thin CLI
// uses to validate ESHU_QUERY_PROFILE without importing core.
func TestParseQueryProfileVectors(t *testing.T) {
	t.Parallel()

	cases := map[string]QueryProfile{
		"production":          ProfileProduction,
		"local_authoritative": ProfileLocalAuthoritative,
		"local_lightweight":   ProfileLocalLightweight,
		"local_full_stack":    ProfileLocalFullStack,
		"  production  ":      ProfileProduction,
		"":                    "",
		"nonsense":            "",
		"PRODUCTION":          "",
	}

	for raw, want := range cases {
		got, err := ParseQueryProfile(raw)
		if want == "" && raw != "" && raw != "nonsense" && raw != "PRODUCTION" {
			continue
		}

		if want == "" {
			// Empty input is valid and yields the empty profile; unknown
			// input must fail.
			if raw == "" {
				if err != nil || got != "" {
					t.Fatalf("ParseQueryProfile(%q) = %q, %v; want empty, nil", raw, got, err)
				}
			} else if _, err := ParseQueryProfile(raw); err == nil {
				t.Fatalf("ParseQueryProfile(%q) succeeded, want error", raw)
			}

			continue
		}

		if err != nil || got != want {
			t.Fatalf("ParseQueryProfile(%q) = %q, %v; want %q, nil", raw, got, err, want)
		}
	}

	if got := NormalizeQueryProfile("production"); got != ProfileProduction {
		t.Fatalf("NormalizeQueryProfile(production) = %q, want production", got)
	}

	if got := NormalizeQueryProfile("nonsense"); got != "" {
		t.Fatalf("NormalizeQueryProfile(nonsense) = %q, want empty", got)
	}
}

// TestParseGraphBackendVectors pins the pure backend parser.
func TestParseGraphBackendVectors(t *testing.T) {
	t.Parallel()

	if got, err := ParseGraphBackend("neo4j"); err != nil || got != GraphBackendNeo4j {
		t.Fatalf("ParseGraphBackend(neo4j) = %q, %v; want neo4j, nil", got, err)
	}

	if got, err := ParseGraphBackend(""); err != nil || got != GraphBackendNornicDB {
		t.Fatalf("ParseGraphBackend(empty) = %q, %v; want nornicdb, nil", got, err)
	}

	if _, err := ParseGraphBackend("nonsense"); err == nil {
		t.Fatalf("ParseGraphBackend(nonsense) succeeded, want error")
	}
}

// TestValidFreshnessCausePinsClosedEnum guards the closed cause vocabulary.
func TestValidFreshnessCausePinsClosedEnum(t *testing.T) {
	t.Parallel()

	for _, cause := range []FreshnessCause{
		FreshnessCausePendingRepoGeneration,
		FreshnessCauseReducerBacklog,
		FreshnessCauseDeadLetteredDomain,
		FreshnessCauseMissingCollectorCompletion,
		FreshnessCauseContentCoverageUnavailable,
		FreshnessCauseUnsupportedProfile,
		FreshnessCauseRetentionExpired,
		FreshnessCausePendingSearchVector,
	} {
		if !ValidFreshnessCause(cause) {
			t.Fatalf("ValidFreshnessCause(%q) = false, want true", cause)
		}
	}

	if ValidFreshnessCause("invented_cause") {
		t.Fatalf("ValidFreshnessCause(invented) = true, want false")
	}
}

// TestRegistryEntriesCoversLookup pins the sorted Entries view and the
// Covers/Lookup helpers the thin CLI uses for config validate.
func TestRegistryEntriesCoversLookup(t *testing.T) {
	t.Parallel()

	reg, err := New(ClientEntries())
	if err != nil {
		t.Fatalf("New(ClientEntries()): %v", err)
	}

	entries := reg.Entries()
	if len(entries) != len(ClientEntries()) {
		t.Fatalf("Entries() = %d rows, want %d", len(entries), len(ClientEntries()))
	}

	for i := 1; i < len(entries); i++ {
		prev, curr := entries[i-1], entries[i]
		if prev.Subsystem > curr.Subsystem {
			t.Fatalf("Entries not sorted by subsystem at %d: %q then %q", i, prev.Subsystem, curr.Subsystem)
		}

		if prev.Subsystem == curr.Subsystem && prev.Name > curr.Name {
			t.Fatalf("Entries not sorted by name at %d: %q then %q", i, prev.Name, curr.Name)
		}
	}

	if !reg.Covers("ESHU_API_KEY") {
		t.Fatalf("Covers(ESHU_API_KEY) = false, want true")
	}

	if !reg.Covers("NEO4J_DATABASE") {
		t.Fatalf("Covers(NEO4J_DATABASE alias) = false, want true")
	}

	if reg.Covers("ESHU_DEFINITELY_NOT_REGISTERED") {
		t.Fatalf("Covers(unknown) = true, want false")
	}

	if _, ok := reg.Lookup("NEO4J_DATABASE"); !ok {
		t.Fatalf("Lookup(alias) failed, want entry")
	}
}
