// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package conformance_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"testing"

	collector "github.com/eshu-hq/eshu-sdk/collector"
	"github.com/eshu-hq/eshu-sdk/collector/conformance"
)

const (
	syntheticPhase       = "proof.synthetic:context" // Mechanism token, not a core registry phase.
	opaqueSyntheticPhase = "proof_phase_v2"          // Mechanism token, not a core registry phase.
)

func coreContextRequest() conformance.Request {
	manifest := validManifest()
	manifest.Spec.EmittedFacts = []conformance.FactFamily{
		{Kind: "incident.record", SchemaVersions: []string{"1.0.0"}, SourceConfidence: []string{"reported"}},
		{Kind: "work_item.record", SchemaVersions: []string{"1.0.0"}, SourceConfidence: []string{"reported"}},
		{Kind: "dev.example.scorecard.snapshot", SchemaVersions: []string{"1.0.0"}, SourceConfidence: []string{"reported"}},
	}
	manifest.Spec.ConsumerContracts.Reducer.Phases = []string{syntheticPhase, conformance.SourceEvidenceOnlyReducerPhase}
	incident := validResult()
	incident.Facts[0].Kind = "incident.record"
	workItem := validResult()
	workItem.Facts[0].Kind = "work_item.record"
	workItem.Facts[0].StableKey = "work:1"
	workItem.Facts[0].SourceRef.FactKey = "work:1"
	empty := validResult()
	empty.State = collector.ResultUnchanged
	empty.Facts = nil
	return conformance.Request{
		Manifest:          manifest,
		Fixtures:          []collector.Result{incident, workItem, empty, validResult()},
		Mode:              conformance.ModeFixture,
		ReservedFactKinds: []string{"incident.record", "work_item.record", "documentation_source"},
		CoreFacts: &conformance.CoreFactContext{
			ComponentID:      manifest.Metadata.ID,
			ComponentVersion: manifest.Metadata.Version,
			Facts: []conformance.CoreFactAuthorization{
				{Kind: "incident.record", Emissions: []conformance.CoreFactEmission{{CollectorKind: "scorecard", SchemaVersions: []string{"1.0.0"}}}, ReducerPhases: []string{syntheticPhase}},
				{Kind: "work_item.record", Emissions: []conformance.CoreFactEmission{{CollectorKind: "scorecard", SchemaVersions: []string{"1.0.0"}}}, ReducerPhases: []string{syntheticPhase}},
			},
		},
	}
}

func TestCoreContextAcceptsExactTrustedCoverageAndMixedFixtures(t *testing.T) {
	req := coreContextRequest()
	before, _ := json.Marshal(req.Manifest)
	report := conformance.Run(req)
	if !report.OK() || report.Summary.FixtureCount != 4 || report.Summary.FactCount != 3 {
		t.Fatalf("report = %#v", report)
	}
	after, _ := json.Marshal(req.Manifest)
	if string(before) != string(after) {
		t.Fatal("Run mutated the manifest")
	}
	coreRaw, _ := json.Marshal(req)
	if strings.Contains(string(coreRaw), "CoreFacts") || strings.Contains(string(coreRaw), "incident.record\",\"emissions") {
		t.Fatalf("trusted context leaked to JSON: %s", coreRaw)
	}
	field, ok := reflect.TypeOf(conformance.Request{}).FieldByName("CoreFacts")
	if !ok || field.Tag.Get("json") != "-" || field.Tag.Get("yaml") != "-" {
		t.Fatalf("CoreFacts wire tags = %q", field.Tag)
	}
	var untrusted conformance.Request
	if err := json.Unmarshal([]byte(`{"CoreFacts":{"ComponentID":"dev.example.collector.scorecard"}}`), &untrusted); err != nil {
		t.Fatal(err)
	}
	if untrusted.CoreFacts != nil {
		t.Fatal("JSON supplied trusted context")
	}
}

func TestCoreContextRejectsInvalidTrustedInput(t *testing.T) {
	cases := []struct {
		name   string
		change func(*conformance.Request)
		code   conformance.FindingCode
	}{
		{"missing reserved inventory", func(r *conformance.Request) { r.ReservedFactKinds = nil }, conformance.FindingManifestInvalid},
		{"no trusted context", func(r *conformance.Request) { r.CoreFacts = nil }, conformance.FindingManifestInvalid},
		{"missing coverage", func(r *conformance.Request) { r.CoreFacts.Facts = r.CoreFacts.Facts[:1] }, conformance.FindingManifestInvalid},
		{"producer mismatch", func(r *conformance.Request) { r.CoreFacts.ComponentID = "other" }, conformance.FindingManifestInvalid},
		{"version mismatch", func(r *conformance.Request) { r.CoreFacts.ComponentVersion = "2.0.0" }, conformance.FindingManifestInvalid},
		{"undeclared authorization", func(r *conformance.Request) { r.CoreFacts.Facts[0].Kind = "documentation_source" }, conformance.FindingManifestInvalid},
		{"unreserved authorization", func(r *conformance.Request) { r.ReservedFactKinds = []string{"work_item.record"} }, conformance.FindingManifestInvalid},
		{"duplicate authorization", func(r *conformance.Request) { r.CoreFacts.Facts = append(r.CoreFacts.Facts, r.CoreFacts.Facts[0]) }, conformance.FindingManifestInvalid},
		{"duplicate declaration", func(r *conformance.Request) {
			r.Manifest.Spec.EmittedFacts = append(r.Manifest.Spec.EmittedFacts, r.Manifest.Spec.EmittedFacts[0])
		}, conformance.FindingManifestInvalid},
		{"empty emission", func(r *conformance.Request) { r.CoreFacts.Facts[0].Emissions = nil }, conformance.FindingManifestInvalid},
		{"unknown collector kind", func(r *conformance.Request) { r.CoreFacts.Facts[0].Emissions[0].CollectorKind = "unknown" }, conformance.FindingManifestInvalid},
		{"duplicate emission", func(r *conformance.Request) {
			r.CoreFacts.Facts[0].Emissions = append(r.CoreFacts.Facts[0].Emissions, r.CoreFacts.Facts[0].Emissions[0])
		}, conformance.FindingManifestInvalid},
		{"uncovered version", func(r *conformance.Request) {
			r.Manifest.Spec.EmittedFacts[0].SchemaVersions = []string{"1.0.0", "1.1.0"}
		}, conformance.FindingManifestInvalid},
		{"undeclared version", func(r *conformance.Request) { r.CoreFacts.Facts[0].Emissions[0].SchemaVersions = []string{"1.1.0"} }, conformance.FindingManifestInvalid},
		{"duplicate version", func(r *conformance.Request) {
			r.CoreFacts.Facts[0].Emissions[0].SchemaVersions = []string{"1.0.0", "1.0.0"}
		}, conformance.FindingManifestInvalid},
		{"noncanonical kind", func(r *conformance.Request) { r.CoreFacts.Facts[0].Kind = " incident.record" }, conformance.FindingManifestInvalid},
		{"noncanonical collector", func(r *conformance.Request) { r.CoreFacts.Facts[0].Emissions[0].CollectorKind = " scorecard" }, conformance.FindingManifestInvalid},
		{"noncanonical phase", func(r *conformance.Request) { r.CoreFacts.Facts[0].ReducerPhases[0] = " proof.synthetic:context" }, conformance.FindingManifestInvalid},
		{"blank phase token", func(r *conformance.Request) { r.CoreFacts.Facts[0].ReducerPhases[0] = "" }, conformance.FindingManifestInvalid},
		{"whitespace phase token", func(r *conformance.Request) { r.CoreFacts.Facts[0].ReducerPhases[0] = " " }, conformance.FindingManifestInvalid},
		{"empty phase", func(r *conformance.Request) { r.CoreFacts.Facts[0].ReducerPhases = nil }, conformance.FindingManifestInvalid},
		{"duplicate phase", func(r *conformance.Request) {
			r.CoreFacts.Facts[0].ReducerPhases = []string{syntheticPhase, syntheticPhase}
		}, conformance.FindingManifestInvalid},
		{"missing trusted phase", func(r *conformance.Request) {
			r.Manifest.Spec.ConsumerContracts.Reducer.Phases = []string{conformance.SourceEvidenceOnlyReducerPhase}
		}, conformance.FindingMissingReducerConsumer},
		{"extra phase", func(r *conformance.Request) {
			r.Manifest.Spec.ConsumerContracts.Reducer.Phases = append(r.Manifest.Spec.ConsumerContracts.Reducer.Phases, "proof.untrusted:phase")
		}, conformance.FindingMissingReducerConsumer},
		{"exact phase mismatch", func(r *conformance.Request) {
			r.Manifest.Spec.ConsumerContracts.Reducer.Phases[0] = "Proof.synthetic:context"
		}, conformance.FindingMissingReducerConsumer},
		{"wrong fixture component", func(r *conformance.Request) { r.Fixtures[2].Claim.ComponentID = "other" }, conformance.FindingFixtureContractFailed},
		{"wrong fixture collector", func(r *conformance.Request) { r.Fixtures[2].Claim.CollectorKind = "other" }, conformance.FindingFixtureContractFailed},
		{"missing trusted schema", func(r *conformance.Request) { r.CoreFacts.Facts[0].PayloadSchemaRef = "incident.record" }, conformance.FindingPayloadSchemaInvalid},
		{"invalid trusted schema ref", func(r *conformance.Request) { r.CoreFacts.Facts[0].PayloadSchemaRef = " incident.record" }, conformance.FindingPayloadSchemaInvalid},
		{"invalid supplied schema", func(r *conformance.Request) {
			r.CoreFacts.Facts[0].PayloadSchemaRef = "incident.record"
			r.PayloadSchemas = map[string]json.RawMessage{"incident.record": json.RawMessage(`{"unsupported":true}`)}
		}, conformance.FindingPayloadSchemaInvalid},
		{"invalid payload", func(r *conformance.Request) {
			r.CoreFacts.Facts[0].PayloadSchemaRef = "incident.record"
			r.PayloadSchemas = map[string]json.RawMessage{"incident.record": json.RawMessage(`{"type":"object","properties":{"score":{"type":"number"}},"required":["score"],"additionalProperties":true}`)}
			r.Fixtures[0].Facts[0].Payload = map[string]any{"wrong": true}
		}, conformance.FindingPayloadSchemaInvalid},
		{"conflicting schema ref", func(r *conformance.Request) {
			r.CoreFacts.Facts[0].PayloadSchemaRef = "incident.record"
			r.Manifest.Spec.EmittedFacts[0].PayloadSchemaRef = "other.record"
		}, conformance.FindingPayloadSchemaInvalid},
		{"manifest schema ref absent in trusted context", func(r *conformance.Request) {
			r.Manifest.Spec.EmittedFacts[0].PayloadSchemaRef = "incident.record"
		}, conformance.FindingPayloadSchemaInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := coreContextRequest()
			tc.change(&req)
			report := conformance.Run(req)
			if report.OK() {
				t.Fatalf("invalid context passed: %#v", report)
			}
			assertFinding(t, report, tc.code)
		})
	}
}

func TestCoreContextRequiresExactCollectorVersionTuple(t *testing.T) {
	req := coreContextRequest()
	req.Manifest.Spec.CollectorKinds = []string{"scorecard", "other"}
	req.Manifest.Spec.EmittedFacts[0].SchemaVersions = []string{"1.0.0", "1.1.0"}
	req.CoreFacts.Facts[0].Emissions = []conformance.CoreFactEmission{
		{CollectorKind: "scorecard", SchemaVersions: []string{"1.0.0"}},
		{CollectorKind: "other", SchemaVersions: []string{"1.1.0"}},
	}
	if report := conformance.Run(req); !report.OK() {
		t.Fatalf("split coverage: %#v", report.Findings)
	}
	req.Fixtures[0].Facts[0].SchemaVersion = "1.1.0" // Each dimension is granted somewhere, but this tuple is not.
	report := conformance.Run(req)
	assertFinding(t, report, conformance.FindingFixtureContractFailed)
}

func TestCoreContextAcceptsUndottedReservedKindAndTrustedSchema(t *testing.T) {
	req := coreContextRequest()
	req.Manifest.Spec.EmittedFacts[0].Kind = "documentation_source"
	req.CoreFacts.Facts[0].Kind = "documentation_source"
	req.CoreFacts.Facts[0].PayloadSchemaRef = "documentation_source"
	req.Fixtures[0].Facts[0].Kind = "documentation_source"
	req.PayloadSchemas = map[string]json.RawMessage{
		"documentation_source": json.RawMessage(`{"type":"object","properties":{"score":{"type":"number"}},"required":["score"],"additionalProperties":true}`),
	}
	if report := conformance.Run(req); !report.OK() {
		t.Fatalf("undotted core kind/schema: %#v", report.Findings)
	}
}

func TestCoreContextAcceptsOpaqueTrustedPhase(t *testing.T) {
	for _, phase := range []string{opaqueSyntheticPhase, "Proof Phase V2"} {
		req := coreContextRequest()
		for index := range req.CoreFacts.Facts {
			req.CoreFacts.Facts[index].ReducerPhases = []string{phase}
		}
		req.Manifest.Spec.ConsumerContracts.Reducer.Phases = []string{phase, conformance.SourceEvidenceOnlyReducerPhase}
		if report := conformance.Run(req); !report.OK() {
			t.Fatalf("matching opaque phase %q rejected: %#v", phase, report.Findings)
		}
	}
}

func TestCoreContextReportOrderIsStable(t *testing.T) {
	req := coreContextRequest()
	req.Manifest.Spec.ConsumerContracts.Reducer.Phases = []string{"proof.z:extra", "proof.a:extra"}
	req.CoreFacts.Facts[1].ReducerPhases = []string{"proof.z:trusted"}
	req.Fixtures[0].Claim.ComponentID = "bad"
	req.Fixtures[1].Claim.CollectorKind = "bad"
	var want []conformance.Finding
	for range 20 {
		got := conformance.Run(req).Findings
		if want == nil {
			want = got
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("nondeterministic findings:\n%#v\n%#v", want, got)
		}
	}
	if len(want) != 6 {
		t.Fatalf("finding count = %d, want four phase and two fixture findings: %#v", len(want), want)
	}
	for index, finding := range want {
		if index < 4 && finding.Code != conformance.FindingMissingReducerConsumer {
			t.Fatalf("phase finding %d = %#v", index, finding)
		}
		if index >= 4 && finding.Code != conformance.FindingFixtureContractFailed {
			t.Fatalf("fixture finding %d = %#v", index, finding)
		}
	}
}

func TestCoreContextConcurrentRunsKeepIndependentReports(t *testing.T) {
	req := coreContextRequest()
	want, err := json.Marshal(conformance.Run(req))
	if err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for range 16 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 20 {
				got, err := json.Marshal(conformance.Run(req))
				if err != nil {
					t.Errorf("marshal report: %v", err)
					return
				}
				if string(got) != string(want) {
					t.Errorf("concurrent report changed: %s", got)
					return
				}
			}
		}()
	}
	workers.Wait()
}

func TestCoreContextLeavesGrantlessReportBytes(t *testing.T) {
	for _, tc := range []struct {
		name    string
		request conformance.Request
		want    string
	}{
		{"passed", conformance.Request{Manifest: validManifest(), Fixtures: []collector.Result{validResult()}}, `{"schema_version":"eshu.extension.conformance.v1","mode":"fixture","status":"passed","component_id":"dev.example.collector.scorecard","component_version":"0.1.0","summary":{"fixture_count":1,"fact_count":1,"duplicate_count":0,"redaction_count":0,"tombstone_count":0,"status_count":0,"idempotent_reemission_checked":true}}`},
		{"failed", conformance.Request{Manifest: validManifest(), Fixtures: []collector.Result{validResult()}, ReservedFactKinds: []string{"dev.example.scorecard.snapshot"}}, `{"schema_version":"eshu.extension.conformance.v1","mode":"fixture","status":"failed","component_id":"dev.example.collector.scorecard","component_version":"0.1.0","findings":[{"code":"manifest_invalid","message":"fact kind \"dev.example.scorecard.snapshot\" is host-owned and cannot be claimed by a component","blocks_publication":true,"blocks_hosted_activation":true}],"summary":{"fixture_count":0,"fact_count":0,"duplicate_count":0,"redaction_count":0,"tombstone_count":0,"status_count":0,"idempotent_reemission_checked":false}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(conformance.Run(tc.request))
			if err != nil {
				t.Fatal(err)
			}
			if string(raw) != tc.want {
				t.Fatalf("report bytes changed:\n%s\nwant:\n%s", raw, tc.want)
			}
		})
	}
}
