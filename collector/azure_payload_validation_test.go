// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/eshu-hq/eshu-sdk/collector/marker"
)

const azureFixtureMarker = "redacted:hmac-sha256:c78a3aed0e9e0e9662dc19a40eddcae8440730c99950d67b15453e8affde39a8"

type azurePayloadJSON string

func (v azurePayloadJSON) MarshalJSON() ([]byte, error) { return []byte(v), nil }

func azureTagResult(t *testing.T) (Result, Validator) {
	t.Helper()
	result := readFixture(t, "complete")
	result.Facts[0].Kind = "azure_tag_observation"
	key, err := marker.NewKey([]byte("azure-tag-observation-test-key-material"))
	if err != nil {
		t.Fatal(err)
	}
	got := marker.String("prod", "azure_tag_value", "azure_tag:env", key).Marker
	if got != azureFixtureMarker {
		t.Fatalf("marker = %q, want fixed public synthetic vector", got)
	}
	result.Facts[0].Payload = map[string]any{
		"tag_value_fingerprints": map[string]string{
			"api_key":                got,
			"token_policy_join_keys": got,
		},
	}
	contract := Contract{ProtocolVersion: ProtocolVersionV1Alpha1, Facts: []FactDeclaration{
		{Kind: "azure_tag_observation", SchemaVersions: []string{"1.0.0", "2.0.0"}, SourceConfidence: []SourceConfidence{SourceConfidenceReported}},
		{Kind: "azure_cloud_resource", SchemaVersions: []string{"1.0.0"}, SourceConfidence: []SourceConfidence{SourceConfidenceReported}},
	}}
	return result, NewValidator(contract)
}

func wireResult(t *testing.T, result Result) Result {
	t.Helper()
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded Result
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func TestAzureTagFingerprintAdmissionPreservesTaxonomy(t *testing.T) {
	result, validator := azureTagResult(t)
	for _, candidate := range []Result{result, wireResult(t, result)} {
		if _, err := validator.ValidateResult(candidate); err != nil {
			t.Fatalf("valid keyed markers rejected: %v", err)
		}
	}
	result.Facts[0].Payload["tag_value_fingerprints"] = map[string]string{}
	for _, candidate := range []Result{result, wireResult(t, result)} {
		if _, err := validator.ValidateResult(candidate); err != nil {
			t.Fatalf("empty fingerprint object rejected: %v", err)
		}
	}
	if err := ValidateShareSafeKeys(map[string]any{"tag_value_fingerprints": map[string]any{"api_key": azureFixtureMarker}}); err == nil {
		t.Fatal("context-free secret-key rejection weakened")
	}
}

func TestAzureTagFingerprintAdmissionRejectsUnsafeRepresentations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Result)
	}{
		{"nil native string map", func(r *Result) { r.Facts[0].Payload["tag_value_fingerprints"] = map[string]string(nil) }},
		{"nil native object map", func(r *Result) { r.Facts[0].Payload["tag_value_fingerprints"] = map[string]any(nil) }},
		{"nil wire value", func(r *Result) { r.Facts[0].Payload["tag_value_fingerprints"] = nil }},
		{"prefix only", func(r *Result) {
			r.Facts[0].Payload["tag_value_fingerprints"] = map[string]string{"api_key": marker.Prefix}
		}},
		{"malformed digest", func(r *Result) {
			r.Facts[0].Payload["tag_value_fingerprints"] = map[string]string{"api_key": marker.Prefix + strings.Repeat("A", 64)}
		}},
		{"safe name with raw value", func(r *Result) {
			r.Facts[0].Payload["tag_value_fingerprints"] = map[string]string{"token_policy_join_keys": "raw"}
		}},
		{"wrong kind", func(r *Result) { r.Facts[0].Kind = "azure_cloud_resource" }},
		{"wrong schema", func(r *Result) { r.Facts[0].SchemaVersion = "2.0.0" }},
		{"nested typed map", func(r *Result) { r.Facts[0].Payload["nested"] = map[string]string{"api_key": "raw"} }},
		{"custom JSON", func(r *Result) { r.Facts[0].Payload["nested"] = azurePayloadJSON("{\"api_key\":\"raw\"}") }},
		{"custom fingerprint JSON", func(r *Result) {
			r.Facts[0].Payload["tag_value_fingerprints"] = azurePayloadJSON("{\"api_key\":\"raw\"}")
		}},
		{"nested trust path", func(r *Result) {
			r.Facts[0].Payload["nested"] = map[string]any{"tag_value_fingerprints": map[string]string{"api_key": azureFixtureMarker}}
		}},
		{"empty outer key", func(r *Result) {
			r.Facts[0].Payload[""] = map[string]any{"tag_value_fingerprints": map[string]string{"api_key": azureFixtureMarker}}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, validator := azureTagResult(t)
			tt.mutate(&result)
			for _, candidate := range []Result{result, wireResult(t, result)} {
				if _, err := validator.ValidateResult(candidate); err == nil {
					t.Fatal("unsafe native or wire payload accepted")
				}
			}
		})
	}
}

func TestFactPayloadValidationPreservesSafeCanonicalValues(t *testing.T) {
	result, validator := azureTagResult(t)
	result.Facts[0].Payload["nested"] = azurePayloadJSON(`{"label":"safe","large":9007199254740993}`)
	result.Facts[0].Payload["numbers"] = []any{json.Number("9007199254740993")}
	for _, candidate := range []Result{result, wireResult(t, result)} {
		if _, err := validator.ValidateResult(candidate); err != nil {
			t.Fatalf("safe Azure payload rejected: %v", err)
		}
	}

	other := readFixture(t, "complete")
	other.Facts[0].Payload["nested"] = map[string]string{"label": "safe"}
	other.Facts[0].Payload["numbers"] = []any{json.Number("9007199254740993")}
	for _, candidate := range []Result{other, wireResult(t, other)} {
		if _, err := NewValidator(testContract()).ValidateResult(candidate); err != nil {
			t.Fatalf("ordinary SDK fact rejected: %v", err)
		}
	}
}
