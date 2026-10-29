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
