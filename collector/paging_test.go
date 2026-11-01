// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestResultPageValidationV1Alpha3 pins the versioned pagination contract
// (eshu-hq/eshu#9385): page fields are admitted only on
// collector-sdk/v1alpha3, a continuation page must carry a cursor, and a
// final page must not.
func TestResultPageValidationV1Alpha3(t *testing.T) {
	t.Parallel()

	contract := testContract()
	contract.ProtocolVersion = ProtocolVersionV1Alpha3
	validator := NewValidator(contract)

	base := Result{
		ProtocolVersion: ProtocolVersionV1Alpha3,
		State:           ResultComplete,
		Claim:           testClaim(),
		Generation: Generation{
			ID:         "gen-20260609",
			ObservedAt: time.Date(2026, time.July, 8, 12, 0, 0, 0, time.UTC),
		},
		Facts: []Fact{testPageFact()},
		Statuses: []Status{{
			Class:     StatusComplete,
			FactCount: 1,
		}},
	}

	for _, tc := range []struct {
		name    string
		mutate  func(*Result)
		wantErr string
	}{
		{
			name:   "absent page stays a valid single-shot result",
			mutate: func(result *Result) {},
		},
		{
			name: "final page with empty cursor",
			mutate: func(result *Result) {
				result.Page = &ResultPage{PageIndex: 2}
			},
		},
		{
			name: "continuation page with cursor",
			mutate: func(result *Result) {
				result.Page = &ResultPage{Cursor: "cursor-1", More: true, TotalFacts: 3}
			},
		},
		{
			name: "partial continuation page with cursor",
			mutate: func(result *Result) {
				result.State = ResultPartial
				result.Statuses = []Status{{Class: StatusWarning, FactCount: 1, WarningCount: 1}}
				result.Page = &ResultPage{Cursor: "cursor-1", More: true}
			},
		},
		{
			name: "continuation without cursor fails",
			mutate: func(result *Result) {
				result.Page = &ResultPage{More: true}
			},
			wantErr: "cursor",
		},
		{
			name: "continuation with whitespace cursor fails",
			mutate: func(result *Result) {
				result.Page = &ResultPage{Cursor: "  ", More: true}
			},
			wantErr: "cursor",
		},
		{
			name: "final page with cursor fails",
			mutate: func(result *Result) {
				result.Page = &ResultPage{Cursor: "cursor-1"}
			},
			wantErr: "cursor",
		},
		{
			name: "negative page index fails",
			mutate: func(result *Result) {
				result.Page = &ResultPage{PageIndex: -1}
			},
			wantErr: "page_index",
		},
		{
			name: "negative total facts fails",
			mutate: func(result *Result) {
				result.Page = &ResultPage{TotalFacts: -1}
			},
			wantErr: "total_facts",
		},
		{
			name: "retryable continuation fails",
			mutate: func(result *Result) {
				result.State = ResultRetryable
				result.Facts = nil
				result.Statuses = []Status{{
					Class:             StatusFailure,
					FailureClass:      "source_unavailable",
					RetryAfterSeconds: 30,
				}}
				result.Page = &ResultPage{Cursor: "cursor-1", More: true}
			},
			wantErr: "continuation",
		},
		{
			name: "terminal continuation fails",
			mutate: func(result *Result) {
				result.State = ResultTerminal
				result.Facts = nil
				result.Statuses = []Status{{
					Class:        StatusFailure,
					FailureClass: "source_unavailable",
				}}
				result.Page = &ResultPage{Cursor: "cursor-1", More: true}
			},
			wantErr: "continuation",
		},
		{
			name: "unchanged continuation fails",
			mutate: func(result *Result) {
				result.State = ResultUnchanged
				result.Facts = nil
				result.Statuses = nil
				result.Page = &ResultPage{Cursor: "cursor-1", More: true}
			},
			wantErr: "continuation",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := base
			tc.mutate(&result)
			_, err := validator.ValidateResult(result)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateResult() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateResult() error = nil, want %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ValidateResult() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

// TestResultPageVersionGate proves a page on the older wire contracts is a
// version violation, and that a legacy single-shot result stays valid when
// validated as the newest protocol version (backward compatibility).
func TestResultPageVersionGate(t *testing.T) {
	t.Parallel()

	for _, protocol := range []string{ProtocolVersionV1Alpha1, ProtocolVersionV1Alpha2} {
		contract := testContract()
		contract.ProtocolVersion = protocol
		paged := Result{
			ProtocolVersion: protocol,
			State:           ResultComplete,
			Claim:           testClaim(),
			Generation: Generation{
				ID:         "gen-20260609",
				ObservedAt: time.Date(2026, time.July, 8, 12, 0, 0, 0, time.UTC),
			},
			Facts:    []Fact{testPageFact()},
			Statuses: []Status{{Class: StatusComplete, FactCount: 1}},
			Page:     &ResultPage{Cursor: "cursor-1", More: true},
		}
		if _, err := NewValidator(contract).ValidateResult(paged); err == nil {
			t.Fatalf("ValidateResult(%s with page) error = nil, want version error", protocol)
		} else if !strings.Contains(err.Error(), "page") {
			t.Fatalf("ValidateResult(%s with page) error = %v, want page version error", protocol, err)
		}
	}

	for _, name := range []string{"complete", "partial", "unchanged", "retryable", "terminal", "tombstone"} {
		legacy := readFixture(t, name)
		newestContract := testContract()
		newestContract.ProtocolVersion = ProtocolVersionV1Alpha3
		legacy.ProtocolVersion = ProtocolVersionV1Alpha3
		if _, err := NewValidator(newestContract).ValidateResult(legacy); err != nil {
			t.Fatalf("ValidateResult(legacy %s as v1alpha3) error = %v, want nil", name, err)
		}
	}
}

// TestOldestDecoderRejectsPagedResult proves the wire-level version story:
// a strict oldest-shape decoder (no page field) rejects a paged v1alpha3
// result as unknown, so old hosts fail closed instead of silently dropping
// the continuation.
func TestOldestDecoderRejectsPagedResult(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("testdata", "fixtures", "paged_first.json"))
	if err != nil {
		t.Fatalf("os.ReadFile(paged_first) error = %v, want nil", err)
	}
	// legacyResult is the Result shape before collector-sdk/v1alpha3.
	type legacyResult struct {
		ProtocolVersion string      `json:"protocol_version"`
		State           ResultState `json:"state"`
		Claim           Claim       `json:"claim"`
		Generation      Generation  `json:"generation"`
		Facts           []Fact      `json:"facts,omitempty"`
		Statuses        []Status    `json:"statuses,omitempty"`
	}
	var legacy legacyResult
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&legacy); err == nil {
		t.Fatal("oldest-shape decode error = nil, want unknown field rejection")
	} else if !strings.Contains(err.Error(), "page") {
		t.Fatalf("oldest-shape decode error = %v, want unknown page field", err)
	}
}

// TestPagedFixturesConform proves the checked-in v1alpha3 page fixtures are
// valid paged results: the continuation page and the final page each pass
// validation, and together they name the same generation and total.
func TestPagedFixturesConform(t *testing.T) {
	t.Parallel()

	contract := testContract()
	contract.ProtocolVersion = ProtocolVersionV1Alpha3
	validator := NewValidator(contract)
	first := readStrictFixture(t, "paged_first")
	final := readStrictFixture(t, "paged_final")
	if _, err := validator.ValidateResult(first); err != nil {
		t.Fatalf("ValidateResult(paged_first) error = %v, want nil", err)
	}
	if _, err := validator.ValidateResult(final); err != nil {
		t.Fatalf("ValidateResult(paged_final) error = %v, want nil", err)
	}
	if first.Page == nil || !first.Page.More || first.Page.Cursor == "" {
		t.Fatal("paged_first must request continuation with a cursor")
	}
	if final.Page == nil || final.Page.More || final.Page.Cursor != "" {
		t.Fatal("paged_final must close the generation without a cursor")
	}
	if first.Generation.ID != final.Generation.ID {
		t.Fatalf("generation drift %q vs %q", first.Generation.ID, final.Generation.ID)
	}
	if first.Page.TotalFacts != final.Page.TotalFacts {
		t.Fatalf("total_facts drift %d vs %d", first.Page.TotalFacts, final.Page.TotalFacts)
	}
	if got := len(first.Facts) + len(final.Facts); got != final.Page.TotalFacts {
		t.Fatalf("reassembled facts = %d, want declared total %d", got, final.Page.TotalFacts)
	}
	if final.Page.PageIndex != first.Page.PageIndex+1 {
		t.Fatalf("page_index gap %d vs %d", first.Page.PageIndex, final.Page.PageIndex)
	}
}

// TestResultPageWireShape pins the exact JSON member names and omitempty
// behavior of the page envelope. Hosts decode it field for field, so a
// renamed tag is a wire break, not a refactor.
func TestResultPageWireShape(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		page ResultPage
		want string
	}{
		{ResultPage{Cursor: "c-1", More: true, PageIndex: 3, TotalFacts: 9}, `{"cursor":"c-1","more":true,"page_index":3,"total_facts":9}`},
		{ResultPage{}, `{}`},
	} {
		raw, err := json.Marshal(tc.page)
		if err != nil {
			t.Fatalf("json.Marshal(%+v) error = %v, want nil", tc.page, err)
		}
		if string(raw) != tc.want {
			t.Fatalf("json.Marshal(%+v) = %s, want %s", tc.page, raw, tc.want)
		}
	}

	legacy := readFixture(t, "complete")
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("json.Marshal(legacy) error = %v, want nil", err)
	}
	if strings.Contains(string(raw), `"page"`) {
		t.Fatalf("pageless result marshals a page member: %s", raw)
	}
}

// TestV1Alpha3SchemaMatchesGolden pins the v1alpha3 result JSON Schema
// artifact to its generator and proves it differs from v1alpha1 only by the
// protocol constant, the schema id, and the optional page member.
func TestV1Alpha3SchemaMatchesGolden(t *testing.T) {
	t.Parallel()

	got, err := JSONSchemaV1Alpha3()
	if err != nil {
		t.Fatalf("JSONSchemaV1Alpha3() error = %v, want nil", err)
	}
	want, err := os.ReadFile(filepath.Join("schema", "collector-sdk-v1alpha3.schema.json"))
	if err != nil {
		t.Fatalf("os.ReadFile(schema) error = %v, want nil", err)
	}
	if !jsonEqual(got, want) {
		t.Fatalf("JSONSchemaV1Alpha3() did not match golden schema\nwant:\n%s\n\ngot:\n%s", want, got)
	}

	var schema map[string]any
	if err := json.Unmarshal(got, &schema); err != nil {
		t.Fatalf("json.Unmarshal(schema) error = %v, want nil", err)
	}
	properties := schema["properties"].(map[string]any)
	if _, ok := properties["page"]; !ok {
		t.Fatal("v1alpha3 schema omits page")
	}
	version := properties["protocol_version"].(map[string]any)["const"]
	if version != ProtocolVersionV1Alpha3 {
		t.Fatalf("v1alpha3 schema protocol const = %v, want %q", version, ProtocolVersionV1Alpha3)
	}

	legacy, err := JSONSchema()
	if err != nil {
		t.Fatalf("JSONSchema() error = %v, want nil", err)
	}
	if strings.Contains(string(legacy), `"page"`) || strings.Contains(string(legacy), ProtocolVersionV1Alpha3) {
		t.Fatal("v1alpha1 schema gained v1alpha3 members")
	}
}

func readStrictFixture(t *testing.T, name string) Result {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "fixtures", name+".json"))
	if err != nil {
		t.Fatalf("os.ReadFile(%s) error = %v, want nil", name, err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	var result Result
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("decode(%s) error = %v, want nil", name, err)
	}
	return result
}
