// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package contracttest

import (
	"testing"

	"github.com/eshu-hq/eshu-sdk/factschema"
)

// FactKindShape describes the expected payload shape for one fact kind.
type FactKindShape struct {
	Kind                string
	RequiredPayloadKeys []string
}

// CollectorContract describes the expected fact shape for one collector.
type CollectorContract struct {
	CollectorKind string
	FactKinds     []FactKindShape
}

// Lookup returns the FactKindShape for kind, or nil.
func (c CollectorContract) Lookup(kind string) *FactKindShape {
	for i := range c.FactKinds {
		if c.FactKinds[i].Kind == kind {
			return &c.FactKinds[i]
		}
	}
	return nil
}

// AssertFactKinds asserts every emitted fact kind is declared in the contract.
func AssertFactKinds(t *testing.T, contract CollectorContract, envelopes []factschema.Envelope) {
	t.Helper()
	for i, envelope := range envelopes {
		if contract.Lookup(envelope.FactKind) == nil {
			t.Errorf("envelope[%d]: fact_kind %q not declared in contract for collector %q", i, envelope.FactKind, contract.CollectorKind)
		}
	}
}

// AssertRequiredPayloadKeys asserts every fact has all required payload keys
// declared in the contract. It only checks facts whose kind is in the contract.
func AssertRequiredPayloadKeys(t *testing.T, contract CollectorContract, envelopes []factschema.Envelope) {
	t.Helper()
	for i, envelope := range envelopes {
		shape := contract.Lookup(envelope.FactKind)
		if shape == nil {
			continue
		}
		for _, key := range shape.RequiredPayloadKeys {
			if _, exists := envelope.Payload[key]; !exists {
				t.Errorf("envelope[%d] fact_kind=%q: missing required payload key %q", i, envelope.FactKind, key)
			}
		}
	}
}

// AssertFactShape runs both fact-kind and required-payload-key assertions in
// a single call. This is the primary entry point.
func AssertFactShape(t *testing.T, contract CollectorContract, envelopes []factschema.Envelope) {
	t.Helper()
	AssertFactKinds(t, contract, envelopes)
	AssertRequiredPayloadKeys(t, contract, envelopes)
}

// EnvelopeCounts returns a map of fact_kind -> count for diagnostic output.
func EnvelopeCounts(envelopes []factschema.Envelope) map[string]int {
	counts := make(map[string]int)
	for _, envelope := range envelopes {
		counts[envelope.FactKind]++
	}
	return counts
}

// ValidateCollectorKind asserts every envelope has the expected collector kind.
func ValidateCollectorKind(t *testing.T, contract CollectorContract, envelopes []factschema.Envelope) {
	t.Helper()
	for i, envelope := range envelopes {
		if envelope.CollectorKind != contract.CollectorKind {
			t.Errorf("envelope[%d] CollectorKind = %q, want %q", i, envelope.CollectorKind, contract.CollectorKind)
		}
	}
}
