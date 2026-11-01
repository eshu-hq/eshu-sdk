// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package conformance_test

import (
	"testing"

	collector "github.com/eshu-hq/eshu-sdk/collector"
	"github.com/eshu-hq/eshu-sdk/collector/conformance"
)

// TestRunAdmitsExplicitCollectorSDKAlpha3 proves an explicitly configured
// collector-sdk/v1alpha3 runtime passes the out-of-tree conformance mirror
// for both a pageless result and a paged continuation result, mirroring the
// core manifest validation that admits the paged protocol (eshu-hq/eshu#9385).
func TestRunAdmitsExplicitCollectorSDKAlpha3(t *testing.T) {
	t.Parallel()

	for _, protocol := range []string{collector.ProtocolVersionV1Alpha3, " " + collector.ProtocolVersionV1Alpha3 + " "} {
		manifest := validManifest()
		manifest.Spec.Runtime.SDKProtocol = protocol
		pageless := validResult()
		pageless.ProtocolVersion = collector.ProtocolVersionV1Alpha3
		paged := validResult()
		paged.ProtocolVersion = collector.ProtocolVersionV1Alpha3
		paged.Page = &collector.ResultPage{Cursor: "page-cursor-1", More: true, TotalFacts: 2}
		report := conformance.Run(conformance.Request{
			Manifest: manifest,
			Fixtures: []collector.Result{pageless, paged},
			Mode:     conformance.ModeFixture,
		})
		if !report.OK() {
			t.Fatalf("protocol %q: findings = %#v, want passed", protocol, report.Findings)
		}
	}
}

// TestRunRejectsPageOnOlderProtocol proves a paged fixture on the v1alpha1
// default fails conformance instead of passing with the page ignored.
func TestRunRejectsPageOnOlderProtocol(t *testing.T) {
	t.Parallel()

	paged := validResult()
	paged.Page = &collector.ResultPage{Cursor: "page-cursor-1", More: true}
	report := conformance.Run(conformance.Request{
		Manifest: validManifest(),
		Fixtures: []collector.Result{paged},
		Mode:     conformance.ModeFixture,
	})
	if report.OK() {
		t.Fatal("v1alpha1 paged fixture passed conformance, want rejection")
	}
}
