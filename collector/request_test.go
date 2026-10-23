// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestRequestValidate(t *testing.T) {
	claim := readFixture(t, "complete").Claim
	claim.Deadline = time.Now().Add(-time.Hour) // Validation is structural, not clock dependent.
	valid := Request{ProtocolVersion: ProtocolVersionV1Alpha1, Claim: claim, Contract: testContract()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid request: %v", err)
	}
	cases := map[string]func(*Request){
		"unknown protocol":  func(r *Request) { r.ProtocolVersion = "collector-sdk/v99" },
		"blank protocol":    func(r *Request) { r.ProtocolVersion = "" },
		"contract mismatch": func(r *Request) { r.Contract.ProtocolVersion = ProtocolVersionV1Alpha2 },
		"invalid claim":     func(r *Request) { r.Claim.WorkItemID = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			mutate(&candidate)
			if err := candidate.Validate(); err == nil {
				t.Fatal("Validate() accepted invalid request")
			}
		})
	}
	valid.ProtocolVersion = ProtocolVersionV1Alpha2
	valid.Contract.ProtocolVersion = ProtocolVersionV1Alpha2
	if err := valid.Validate(); err != nil {
		t.Fatalf("explicit v1alpha2: %v", err)
	}
}

func TestRequestJSONSchemaMatchesGolden(t *testing.T) {
	got, err := RequestJSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("schema/collector-request-v1alpha1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(got, want) {
		t.Fatalf("request schema drift:\n%s", got)
	}
	var schema map[string]any
	if err := json.Unmarshal(got, &schema); err != nil {
		t.Fatal(err)
	}
	properties := schema["properties"].(map[string]any)
	if _, ok := properties["config"]; !ok {
		t.Fatal("request schema omits config")
	}
}

func TestRequestFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/fixtures/request.json")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var request Request
	if err := decoder.Decode(&request); err != nil {
		t.Fatal(err)
	}
	if err := request.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := request.Config["large_integer"].(json.Number).String(); got != "9007199254740993" {
		t.Fatalf("large_integer=%s", got)
	}
}
