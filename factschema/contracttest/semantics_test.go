// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package contracttest

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/eshu-hq/eshu-sdk/factschema"
)

func TestPositiveControls(t *testing.T) {
	contract := CollectorContract{CollectorKind: "kind", FactKinds: []FactKindShape{{Kind: "declared", RequiredPayloadKeys: []string{"required"}}}}
	AssertFactShape(t, contract, nil)
	AssertRequiredPayloadKeys(t, contract, []factschema.Envelope{{FactKind: "unknown"}})
	AssertFactShape(t, contract, []factschema.Envelope{{FactKind: "declared", CollectorKind: "other", Payload: map[string]any{"required": nil, "extra": true}}})
	counts := EnvelopeCounts([]factschema.Envelope{{FactKind: "declared"}, {FactKind: "declared"}})
	if counts["declared"] != 2 {
		t.Fatalf("count = %d, want 2", counts["declared"])
	}
	if len(EnvelopeCounts(nil)) != 0 {
		t.Fatal("empty input returned nonempty counts")
	}
	if contract.Lookup("unknown") != nil {
		t.Fatal("unknown kind resolved")
	}
	if shape := contract.Lookup("declared"); shape == nil || shape.Kind != "declared" {
		t.Fatal("declared kind missing")
	}
}

func TestAssertionFailureProcess(t *testing.T) {
	for _, tt := range []struct {
		name, diagnostic string
		fail             bool
	}{
		{"undeclared_standalone", `fact_kind "unknown" not declared`, true},
		{"undeclared_composed", `fact_kind "unknown" not declared`, true},
		{"missing_standalone", `missing required payload key "required"`, true},
		{"missing_composed", `missing required payload key "required"`, true},
		{"nil_payload", `missing required payload key "required"`, true},
		{"wrong_collector", `CollectorKind = "other", want "kind"`, true},
		{"valid", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestAssertionProbe$")
			command.Env = append(os.Environ(), "CONTRACTTEST_PROBE="+tt.name)
			output, err := command.CombinedOutput()
			if tt.fail && err == nil {
				t.Fatalf("probe passed unexpectedly: %s", output)
			}
			if !tt.fail && err != nil {
				t.Fatalf("valid probe failed: %v: %s", err, output)
			}
			if tt.fail && !strings.Contains(string(output), tt.diagnostic) {
				t.Fatalf("probe output %q lacks %q", output, tt.diagnostic)
			}
		})
	}
}

func TestAssertionProbe(t *testing.T) {
	mode := os.Getenv("CONTRACTTEST_PROBE")
	if mode == "" {
		return
	}
	contract := CollectorContract{CollectorKind: "kind", FactKinds: []FactKindShape{{Kind: "declared", RequiredPayloadKeys: []string{"required"}}}}
	unknown := []factschema.Envelope{{FactKind: "unknown"}}
	missing := []factschema.Envelope{{FactKind: "declared", Payload: map[string]any{"extra": true}}}
	switch mode {
	case "undeclared_standalone":
		AssertFactKinds(t, contract, unknown)
	case "undeclared_composed":
		AssertFactShape(t, contract, unknown)
	case "missing_standalone":
		AssertRequiredPayloadKeys(t, contract, missing)
	case "missing_composed":
		AssertFactShape(t, contract, missing)
	case "nil_payload":
		AssertRequiredPayloadKeys(t, contract, []factschema.Envelope{{FactKind: "declared"}})
	case "wrong_collector":
		ValidateCollectorKind(t, contract, []factschema.Envelope{{CollectorKind: "other"}})
	case "valid":
		AssertFactShape(t, contract, []factschema.Envelope{{FactKind: "declared", CollectorKind: "other", Payload: map[string]any{"required": nil, "extra": true}}})
	default:
		t.Fatalf("unknown probe mode %q", mode)
	}
}
