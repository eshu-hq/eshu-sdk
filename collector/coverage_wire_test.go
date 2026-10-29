// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"encoding/json"
	"os"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

func TestCoverageSchemaRequiresReportForPublishingStates(t *testing.T) {
	schemaRaw, err := os.ReadFile("schema/collector-sdk-v1alpha3.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schemaDoc any
	if err := json.Unmarshal(schemaRaw, &schemaDoc); err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("coverage.json", schemaDoc); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("coverage.json")
	if err != nil {
		t.Fatal(err)
	}

	fixtureRaw, err := os.ReadFile("testdata/fixtures/coverage-v1alpha3.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name            string
		state           ResultState
		removeCoverage  bool
		valid           bool
		trustedDeletion bool
	}{
		{"partial with coverage", ResultPartial, false, true, false},
		{"trusted deletion", ResultPartial, false, false, true},
		{"complete with coverage", ResultComplete, false, true, false},
		{"partial missing coverage", ResultPartial, true, false, false},
		{"complete missing coverage", ResultComplete, true, false, false},
		{"retryable without coverage", ResultRetryable, true, true, false},
		{"retryable with coverage", ResultRetryable, false, false, false},
		{"terminal without coverage", ResultTerminal, true, true, false},
		{"terminal with coverage", ResultTerminal, false, false, false},
		{"unchanged without coverage", ResultUnchanged, true, true, false},
		{"unchanged with coverage", ResultUnchanged, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var result map[string]any
			if err := json.Unmarshal(fixtureRaw, &result); err != nil {
				t.Fatal(err)
			}
			result["state"] = string(tc.state)
			if tc.removeCoverage {
				delete(result, "coverage")
			}
			if tc.trustedDeletion {
				result["generation"].(map[string]any)["trusted_deletion"] = true
			}
			err := schema.Validate(result)
			if (err == nil) != tc.valid {
				t.Fatalf("schema validation error=%v, want valid=%t", err, tc.valid)
			}
		})
	}
}
