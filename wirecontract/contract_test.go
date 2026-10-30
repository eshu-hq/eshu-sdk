// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package wirecontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestGoldenEnvelopeBytes pins the wire contract: the SDK shapes marshal to
// the exact bytes the core querycontract types produced. A renamed JSON tag,
// a retyped field, or a dropped omitempty fails here before any consumer can
// pin the module.
func TestGoldenEnvelopeBytes(t *testing.T) {
	t.Parallel()

	canonical := ResponseEnvelope{
		Data: map[string]any{"truncated": true, "rows": []any{"a", "b"}},
		Truth: &TruthEnvelope{
			Level:      TruthLevelDerived,
			Capability: "search_code",
			Profile:    ProfileProduction,
			Basis:      TruthBasisHybrid,
			Backend:    GraphBackendNeo4j,
			Freshness: TruthFreshness{
				State:      FreshnessStale,
				ObservedAt: "2026-10-29T00:00:00Z",
				Detail:     "reducer backlog",
				Cause:      FreshnessCauseReducerBacklog,
				NextCheck: &FreshnessNextCheck{
					Tool:   "get_reducer_status",
					Route:  "/api/v0/admin/reducer/status",
					Reason: "drain state",
					Params: map[string]string{"scope": "repo"},
				},
			},
			Reason:    "hybrid evidence",
			Truncated: true,
			Omissions: []TruthOmission{
				{Section: "citations", Detail: "handles", Total: 7},
			},
		},
	}

	want, err := os.ReadFile(filepath.Join("testdata", "golden", "envelope_truth.json"))
	if err != nil {
		t.Fatalf("read golden envelope_truth.json: %v", err)
	}

	raw, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		t.Fatalf("marshal canonical: %v", err)
	}

	got := append(raw, '\n')
	if string(got) != string(want) {
		t.Fatalf("SDK marshal differs from golden envelope_truth.json:\n got: %s\nwant: %s", got, want)
	}

	var decoded ResponseEnvelope
	if err := json.Unmarshal(want, &decoded); err != nil {
		t.Fatalf("unmarshal golden into SDK types: %v", err)
	}

	raw2, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		t.Fatalf("marshal decoded: %v", err)
	}

	if string(append(raw2, '\n')) != string(want) {
		t.Fatalf("golden does not round-trip:\n got: %s\nwant: %s", raw2, want)
	}
}

// TestGoldenErrorEnvelope pins the error wire shape against the core bytes.
func TestGoldenErrorEnvelope(t *testing.T) {
	t.Parallel()

	canonical := ResponseEnvelope{
		Error: &ErrorEnvelope{
			Code:       ErrorCodeNotFound,
			Message:    "no such service",
			Capability: "service_story",
			Profiles: &ErrorProfiles{
				Current:  ProfileProduction,
				Required: ProfileLocalAuthoritative,
			},
			Details: map[string]any{"service": "api"},
		},
	}

	want, err := os.ReadFile(filepath.Join("testdata", "golden", "envelope_error.json"))
	if err != nil {
		t.Fatalf("read golden envelope_error.json: %v", err)
	}

	raw, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		t.Fatalf("marshal canonical: %v", err)
	}

	if got := append(raw, '\n'); string(got) != string(want) {
		t.Fatalf("SDK marshal differs from golden envelope_error.json:\n got: %s\nwant: %s", got, want)
	}
}

// TestGoldenToolDefinition pins the MCP tool-definition wire shape against
// the bytes the core toolcontract type produced.
func TestGoldenToolDefinition(t *testing.T) {
	t.Parallel()

	canonical := ToolDefinition{
		Name:        "search_code",
		Description: "Search code across repositories",
		InputSchema: map[string]any{"type": "object"},
	}

	want, err := os.ReadFile(filepath.Join("testdata", "golden", "tool_definition.json"))
	if err != nil {
		t.Fatalf("read golden tool_definition.json: %v", err)
	}

	raw, err := json.MarshalIndent(canonical, "", "  ")
	if err != nil {
		t.Fatalf("marshal canonical: %v", err)
	}

	if got := append(raw, '\n'); string(got) != string(want) {
		t.Fatalf("SDK ToolDefinition marshal differs:\n got: %s\nwant: %s", got, want)
	}

	var decoded ToolDefinition
	if err := json.Unmarshal(want, &decoded); err != nil {
		t.Fatalf("unmarshal tool golden into SDK type: %v", err)
	}

	if !reflect.DeepEqual(decoded, canonical) {
		t.Fatalf("tool golden round-trip mismatch:\n got: %+v\nwant: %+v", decoded, canonical)
	}
}

// TestClientEntriesMatchGolden pins the published client variable table to
// the envregistry subset the core generator snapshotted.
func TestClientEntriesMatchGolden(t *testing.T) {
	t.Parallel()

	want, err := os.ReadFile(filepath.Join("testdata", "golden", "env_client_entries.json"))
	if err != nil {
		t.Fatalf("read golden env_client_entries.json: %v", err)
	}

	var golden []Entry
	if err := json.Unmarshal(want, &golden); err != nil {
		t.Fatalf("unmarshal entries golden: %v", err)
	}

	if got := ClientEntries(); !reflect.DeepEqual(got, golden) {
		t.Fatalf("ClientEntries differs from golden:\n got: %+v\nwant: %+v", got, golden)
	}
}

// TestNoInternalImports enforces the module independence contract: no
// non-test Go source in this module may import Eshu core internals. Test
// files may name the forbidden prefix in their assertions.
func TestNoInternalImports(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read module dir: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || len(name) < 4 || name[len(name)-3:] != ".go" {
			continue
		}

		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}

		if strings.Contains(string(raw), `"github.com/eshu-hq/eshu/go/internal`) {
			t.Fatalf("%s imports go/internal: module must not depend on Eshu core internals", name)
		}
	}
}

func readGolden(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "golden", name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}

	return raw
}

// TestGoModStaysStandalone is the machine-checked half of the module's
// independence contract: no require or replace directives.
func TestGoModStaysStandalone(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "require ") || trimmed == "require" || strings.HasPrefix(trimmed, "require(") {
			t.Fatalf("go.mod must carry no require directives, found: %q", line)
		}

		if strings.Contains(line, "replace") {
			t.Fatalf("go.mod must carry no replace directives, found: %q", line)
		}
	}

	if strings.Contains(string(raw), "github.com/eshu-hq/eshu/go/internal") {
		t.Fatalf("go.mod must not reference go/internal")
	}
}
