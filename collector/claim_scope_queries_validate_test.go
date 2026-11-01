// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"strings"
	"testing"
)

// TestValidateClaimScopeQueries pins the Go validator to the published
// schemas: a claim may carry at most MaxScopeQueries entries, and every
// entry needs a non-blank package_name and version. Without it a slot-scope
// claim the schema rejects would pass the SDK and fail only downstream.
func TestValidateClaimScopeQueries(t *testing.T) {
	t.Parallel()

	base := readFixture(t, "complete").Claim
	valid := func(n int) []ScopeQuery {
		queries := make([]ScopeQuery, n)
		for i := range queries {
			queries[i] = ScopeQuery{Ecosystem: "go", Package: "example.com/pkg", Version: "v1.0.0"}
		}
		return queries
	}

	accepted := map[string][]ScopeQuery{
		"none":         nil,
		"one":          valid(1),
		"at the bound": valid(MaxScopeQueries),
		"no ecosystem": {{Package: "p", Version: "1"}},
	}
	for name, queries := range accepted {
		t.Run("accepts "+name, func(t *testing.T) {
			t.Parallel()

			claim := base
			claim.ScopeQueries = queries
			if err := validateClaim(claim); err != nil {
				t.Fatalf("validateClaim() = %v, want nil", err)
			}
		})
	}

	rejected := map[string]struct {
		queries []ScopeQuery
		want    string
	}{
		"over the bound":  {valid(MaxScopeQueries + 1), "scope_queries"},
		"blank package":   {[]ScopeQuery{{Package: " ", Version: "1"}}, "scope_queries[0].package_name"},
		"blank version":   {[]ScopeQuery{{Package: "p", Version: "\t"}}, "scope_queries[0].version"},
		"second is blank": {[]ScopeQuery{{Package: "p", Version: "1"}, {}}, "scope_queries[1]"},
		"many blank":      {make([]ScopeQuery, MaxScopeQueries+51), "scope_queries"},
	}
	for name, tc := range rejected {
		t.Run("rejects "+name, func(t *testing.T) {
			t.Parallel()

			claim := base
			claim.ScopeQueries = tc.queries
			err := validateClaim(claim)
			if err == nil {
				t.Fatal("validateClaim() accepted an invalid scope query list")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("validateClaim() = %q, want it to name %q", err, tc.want)
			}
		})
	}
}
