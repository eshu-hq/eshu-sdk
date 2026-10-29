// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package conformance_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/eshu-hq/eshu-sdk/collector/conformance"
)

type leakingPayloadMarshaler struct{}

func (leakingPayloadMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("synthetic-private-value-canary")
}

func TestValidatePayloadSchemasDoesNotExposeMarshalErrorText(t *testing.T) {
	result := awsResourceResult(map[string]any{"body": leakingPayloadMarshaler{}})
	err := conformance.ValidatePayloadSchemas(map[string]json.RawMessage{
		awsResourceKind: json.RawMessage(awsResourceSchema),
	}, result)
	if err == nil {
		t.Fatal("failing custom marshaler accepted")
	}
	if strings.Contains(err.Error(), "synthetic-private-value-canary") {
		t.Fatalf("producer error text escaped schema validation: %v", err)
	}
	if !strings.Contains(err.Error(), "JSON serializable") {
		t.Fatalf("marshal failure stage lost: %v", err)
	}
}

func TestValidatePayloadSchemasDoesNotEchoRejectedValues(t *testing.T) {
	const dateTimeSchema = `{"type":"object","properties":{"observed_at":{"type":"string","format":"date-time"}}}`
	tests := []struct {
		name    string
		schema  string
		field   string
		value   any
		private string
		stage   string
	}{
		{"number overflow", awsResourceSchema, "body", json.Number("1e1000"), "1e1000", "re-decoded as JSON"},
		{"enum mismatch", enumPayloadSchema, "state", "synthetic-private-value-canary", "synthetic-private-value-canary", "state"},
		{"date-time mismatch", dateTimeSchema, "observed_at", "synthetic-private-value-canary", "synthetic-private-value-canary", "observed_at"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := validAWSResourcePayload()
			payload[tc.field] = tc.value
			err := conformance.ValidatePayloadSchemas(map[string]json.RawMessage{
				awsResourceKind: json.RawMessage(tc.schema),
			}, awsResourceResult(payload))
			if err == nil {
				t.Fatal("unsafe payload accepted")
			}
			if strings.Contains(err.Error(), tc.private) {
				t.Fatalf("producer value escaped schema validation: %v", err)
			}
			if !strings.Contains(err.Error(), tc.stage) {
				t.Fatalf("failure reason or field lost: %v", err)
			}
		})
	}
}
