// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestRequestContinuationValidate pins the collector-sdk/v1alpha3 request
// side of pagination (eshu-hq/eshu#9385): the host re-invokes a satellite
// with the cursor the previous page minted, only on v1alpha3, and never with
// a blank cursor.
func TestRequestContinuationValidate(t *testing.T) {
	t.Parallel()

	claim := readFixture(t, "complete").Claim
	alpha3 := Request{
		ProtocolVersion: ProtocolVersionV1Alpha3,
		Claim:           claim,
		Contract:        Contract{ProtocolVersion: ProtocolVersionV1Alpha3},
	}
	if err := alpha3.Validate(); err != nil {
		t.Fatalf("v1alpha3 first call: %v", err)
	}
	resumed := alpha3
	resumed.Continuation = &PageContinuation{Cursor: "page-cursor-1"}
	if err := resumed.Validate(); err != nil {
		t.Fatalf("v1alpha3 continuation: %v", err)
	}

	cases := map[string]func(*Request){
		"blank cursor":      func(r *Request) { r.Continuation = &PageContinuation{} },
		"whitespace cursor": func(r *Request) { r.Continuation = &PageContinuation{Cursor: " \t"} },
		"v1alpha1 continuation": func(r *Request) {
			r.ProtocolVersion = ProtocolVersionV1Alpha1
			r.Contract.ProtocolVersion = ProtocolVersionV1Alpha1
		},
		"v1alpha2 continuation": func(r *Request) {
			r.ProtocolVersion = ProtocolVersionV1Alpha2
			r.Contract.ProtocolVersion = ProtocolVersionV1Alpha2
		},
		"contract mismatch": func(r *Request) { r.Contract.ProtocolVersion = ProtocolVersionV1Alpha1 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			candidate := resumed
			mutate(&candidate)
			if err := candidate.Validate(); err == nil {
				t.Fatal("Validate() accepted invalid continuation request")
			}
		})
	}
}

// TestRequestContinuationWireShape pins the host wire shape
// {"continuation":{"cursor":"..."}} and proves a request without a
// continuation marshals byte-identical to the legacy request.
func TestRequestContinuationWireShape(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/fixtures/request_continuation.json")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var request Request
	if err := decoder.Decode(&request); err != nil {
		t.Fatalf("strict decode continuation fixture: %v", err)
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("continuation fixture invalid: %v", err)
	}
	if request.Continuation == nil || request.Continuation.Cursor != "page-cursor-1" {
		t.Fatalf("continuation = %+v, want cursor page-cursor-1", request.Continuation)
	}

	encoded, err := json.Marshal(PageContinuation{Cursor: "c-1"})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"cursor":"c-1"}` {
		t.Fatalf("PageContinuation wire = %s", encoded)
	}

	request.Continuation = nil
	firstCall, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(firstCall), "continuation") {
		t.Fatalf("first-call request marshals a continuation member: %s", firstCall)
	}
}

// TestOldestDecoderRejectsContinuationRequest proves a satellite built on a
// pre-v1alpha3 SDK (strict Request decode without the continuation member)
// fails closed on a resumed call instead of silently restarting at page 0.
func TestOldestDecoderRejectsContinuationRequest(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/fixtures/request_continuation.json")
	if err != nil {
		t.Fatal(err)
	}
	type legacyRequest struct {
		ProtocolVersion string         `json:"protocol_version"`
		Claim           Claim          `json:"claim"`
		Contract        Contract       `json:"contract"`
		Config          map[string]any `json:"config,omitempty"`
	}
	var legacy legacyRequest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&legacy); err == nil {
		t.Fatal("legacy decode accepted a continuation request")
	} else if !strings.Contains(err.Error(), "continuation") {
		t.Fatalf("legacy decode error = %v, want unknown continuation field", err)
	}
}

// TestRequestJSONSchemaV1Alpha3MatchesGolden pins the v1alpha3 request
// schema artifact and proves the v1alpha1 request artifact gains no
// v1alpha3 member.
func TestRequestJSONSchemaV1Alpha3MatchesGolden(t *testing.T) {
	t.Parallel()

	got, err := RequestJSONSchemaV1Alpha3()
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("schema/collector-request-v1alpha3.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("v1alpha3 request schema drift:\n%s", got)
	}
	var schema map[string]any
	if err := json.Unmarshal(got, &schema); err != nil {
		t.Fatal(err)
	}
	properties := schema["properties"].(map[string]any)
	if _, ok := properties["continuation"]; !ok {
		t.Fatal("v1alpha3 request schema omits continuation")
	}
	if got := properties["protocol_version"].(map[string]any)["const"]; got != ProtocolVersionV1Alpha3 {
		t.Fatalf("v1alpha3 request protocol const = %v", got)
	}

	legacy, err := RequestJSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(legacy, []byte("continuation")) || bytes.Contains(legacy, []byte(ProtocolVersionV1Alpha3)) {
		t.Fatal("v1alpha1 request schema gained v1alpha3 members")
	}
}
