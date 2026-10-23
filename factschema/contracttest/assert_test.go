// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package contracttest

import (
	"testing"

	"github.com/eshu-hq/eshu-sdk/factschema"
)

func TestAssertFactKinds(t *testing.T) {
	contract := CollectorContract{CollectorKind: "aws", FactKinds: []FactKindShape{{Kind: "aws_resource"}, {Kind: "aws_relationship"}}}
	envelopes := []factschema.Envelope{{FactKind: "aws_resource"}, {FactKind: "aws_relationship"}}
	AssertFactShape(t, contract, envelopes)
}

func TestAssertRequiredPayloadKeys(t *testing.T) {
	contract := CollectorContract{CollectorKind: "aws", FactKinds: []FactKindShape{{Kind: "aws_resource", RequiredPayloadKeys: []string{"arn", "resource_type"}}}}
	envelopes := []factschema.Envelope{{FactKind: "aws_resource", Payload: map[string]any{"arn": "a", "resource_type": "t"}}}
	AssertFactShape(t, contract, envelopes)
}

func TestValidateCollectorKind(t *testing.T) {
	contract := CollectorContract{CollectorKind: "aws", FactKinds: []FactKindShape{{Kind: "aws_resource"}}}
	envelopes := []factschema.Envelope{{FactKind: "aws_resource", CollectorKind: "aws"}}
	ValidateCollectorKind(t, contract, envelopes)
}

func TestEnvelopeCounts(t *testing.T) {
	envelopes := []factschema.Envelope{{FactKind: "aws_resource"}, {FactKind: "aws_resource"}, {FactKind: "aws_relationship"}}
	counts := EnvelopeCounts(envelopes)
	if counts["aws_resource"] != 2 {
		t.Errorf("aws_resource count = %d, want 2", counts["aws_resource"])
	}
	if counts["aws_relationship"] != 1 {
		t.Errorf("aws_relationship count = %d, want 1", counts["aws_relationship"])
	}
}
