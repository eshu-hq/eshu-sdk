// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package wirecontract

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// validateVectors mirrors the generator inputs so the SDK Validate behavior
// is checked against the client-scoped goldens the live registry produced.
func validateVectors() map[string]struct {
	env    map[string]string
	strict bool
	golden string
} {
	return map[string]struct {
		env    map[string]string
		strict bool
		golden string
	}{
		"valid": {
			env: map[string]string{
				"ESHU_API_KEY":          "redacted",
				"ESHU_QUERY_PROFILE":    "production",
				"ESHU_MCP_TRANSPORT":    "stdio",
				"ESHU_GRAPH_BACKEND":    "nornicdb",
				"ESHU_HOME":             "/tmp/x",
				"ESHU_UNRELATED_CUSTOM": "kept-quiet",
			},
			golden: "validate_client_valid_client_env.json",
		},
		"typo": {
			env: map[string]string{
				"ESHU_QUERY_PROFILE": "nonsense",
				"ESHU_MCP_TRANSPOR":  "stdio",
				"ESHU_API_KEY":       "redacted",
			},
			golden: "validate_client_invalid_and_typo.json",
		},
		"strict": {
			env: map[string]string{
				"ESHU_API_KEY":        "redacted",
				"ESHU_CUSTOM_THING_X": "1",
			},
			strict: true,
			golden: "validate_client_strict_unknown.json",
		},
	}
}

// TestValidateVectors pins the ported Validate behavior: the same inputs
// through the SDK client registry must yield the findings the live core
// registry produced on the same scope. A divergence in typo suggestions,
// strict handling, or per-type value checks fails here.
func TestValidateVectors(t *testing.T) {
	t.Parallel()
	reg, err := New(ClientEntries())
	if err != nil {
		t.Fatalf("New(ClientEntries()): %v", err)
	}
	for name, vec := range validateVectors() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var want []Finding
			if err := json.Unmarshal(readGolden(t, vec.golden), &want); err != nil {
				t.Fatalf("unmarshal golden %s: %v", vec.golden, err)
			}
			got := reg.Validate(vec.env, vec.strict)
			if got == nil {
				got = []Finding{}
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("Validate(%v, strict=%v) mismatch:\n got: %+v\nwant: %+v", vec.env, vec.strict, got, want)
			}
		})
	}
}

// TestClientRegistryVocabularySplitDocumentsExclusion pins the inherited
// core split: ParseQueryProfile accepts local_full_stack (trimming spaces),
// while the client registry admits only three ESHU_QUERY_PROFILE values.
// Enum comparison is untrimmed, exactly like core: padded input is invalid.
// Both halves are faithful to core sdk/go/apicontract; do not "fix" one
// side without the core comparison.
func TestClientRegistryVocabularySplitDocumentsExclusion(t *testing.T) {
	t.Parallel()

	if _, err := ParseQueryProfile("local_full_stack"); err != nil {
		t.Fatalf("ParseQueryProfile(local_full_stack) = %v, want nil", err)
	}
	reg, err := New(ClientEntries())
	if err != nil {
		t.Fatalf("New(ClientEntries()): %v", err)
	}
	for _, value := range []string{"local_full_stack", "  production  "} {
		findings := reg.Validate(map[string]string{"ESHU_QUERY_PROFILE": value}, false)
		found := false
		for _, f := range findings {
			if f.Name == "ESHU_QUERY_PROFILE" {
				found = true
			}
		}
		if !found {
			t.Fatalf("Validate(ESHU_QUERY_PROFILE=%q) has no finding, want invalid_value", value)
		}
	}
	// Bool validation trims spaces, like core: padded input stays valid.
	if findings := reg.Validate(map[string]string{"ESHU_MCP_ALLOW_UNAUTHENTICATED": " true "}, false); len(findings) != 0 {
		t.Fatalf("padded bool yields findings, want 0: %+v", findings)
	}
}

// TestNewRejectsCanonicalShadowingAlias proves a canonical name that shadows
// an earlier entry's alias fails fast instead of silently killing the alias:
// Lookup would resolve the name to the new entry and leave the alias dead.
func TestNewRejectsCanonicalShadowingAlias(t *testing.T) {
	t.Parallel()

	entries := []Entry{
		{Name: "ESHU_X_ONE", Type: VarString, Subsystem: "api", Aliases: []string{"ESHU_X_TWO"}},
		{Name: "ESHU_X_TWO", Type: VarString, Subsystem: "api"},
	}
	if _, err := New(entries); err == nil {
		t.Fatal("New() with canonical shadowing an earlier alias = nil, want an error")
	} else if !strings.Contains(err.Error(), "shadows alias") {
		t.Fatalf("New() error = %q, want the shadows-alias refusal", err.Error())
	}
}

// TestValidateEmptyEnvIsClean guards the normal first-run state: an operator
// who never set a client variable gets no findings, not a nil-map surprise.
func TestValidateEmptyEnvIsClean(t *testing.T) {
	t.Parallel()
	reg, err := New(ClientEntries())
	if err != nil {
		t.Fatalf("New(ClientEntries()): %v", err)
	}
	if findings := reg.Validate(map[string]string{}, false); len(findings) != 0 {
		t.Fatalf("empty env yields %d findings, want 0: %+v", len(findings), findings)
	}
	if findings := reg.Validate(map[string]string{}, true); len(findings) != 0 {
		t.Fatalf("empty env in strict mode yields %d findings, want 0: %+v", len(findings), findings)
	}
}
