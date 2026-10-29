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

func coverageFixture(t *testing.T) (Request, Result) {
	t.Helper()
	raw, err := os.ReadFile("testdata/fixtures/coverage-v1alpha3.json")
	if err != nil {
		t.Fatal(err)
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	contract := Contract{ProtocolVersion: ProtocolVersionV1Alpha3, Coverage: &CoverageContract{
		Scope: CoverageScope{TenantID: "tenant", ComponentID: result.Claim.ComponentID, AccountID: "123456789012", Partition: "aws", Service: "ec2", Region: "us-east-1"},
		Root:  RootEC2Resources, MaxResources: 4,
	}}
	return Request{ProtocolVersion: ProtocolVersionV1Alpha3, Claim: result.Claim, Contract: contract}, result
}

func TestCoverageFixtureAndSchema(t *testing.T) {
	request, result := coverageFixture(t)
	if err := request.Validate(); err != nil {
		t.Fatalf("request: %v", err)
	}
	if _, err := NewValidator(request.Contract).ValidateResult(result); err != nil {
		t.Fatalf("result: %v", err)
	}
	raw, err := CoverageJSONSchema()
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("schema/collector-sdk-v1alpha3.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, golden) {
		t.Fatal("v1alpha3 result schema drift")
	}
}

func TestCoverageRejectsTrustedDeletion(t *testing.T) {
	request, result := coverageFixture(t)
	result.Generation.TrustedDeletion = true
	if _, err := NewValidator(request.Contract).ValidateResult(result); err == nil {
		t.Fatal("trusted deletion overrode incomplete coverage")
	}
}

func TestCoverageRootAndResourceOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Request, *Result)
	}{
		{"complete observed resource", func(_ *Request, r *Result) {
			r.Coverage.Sets[1].State = CoverageComplete
			r.Coverage.Sets[1].Reason = ""
			r.State = ResultComplete
		}},
		{"explicit absence after complete root", func(_ *Request, r *Result) {
			r.Coverage.Resources[0].State = ResourceAbsent
			r.Coverage.Sets = []SetCoverage{}
			r.State = ResultComplete
		}},
		{"listed resource with missing core detail", func(_ *Request, r *Result) {
			r.Coverage.Resources[0].State = ResourceIncomplete
			r.Coverage.Resources[0].Reason = CoverageSourceError
			r.Coverage.Sets[0].State = CoverageIncomplete
			r.Coverage.Sets[0].Reason = CoverageSourceError
		}},
		{"capacity failure", func(_ *Request, r *Result) {
			r.Coverage.Root.State = CoverageIncomplete
			r.Coverage.Root.Reason = CoverageCapacityExceeded
			r.Coverage.Resources = []ResourceObservation{}
			r.Coverage.Sets = []SetCoverage{}
		}},
		{"incomplete root", func(_ *Request, r *Result) {
			r.Coverage.Root.State = CoverageIncomplete
			r.Coverage.Root.Reason = CoverageThrottled
			r.Coverage.Resources = []ResourceObservation{}
			r.Coverage.Sets = []SetCoverage{}
		}},
		{"required service failure", func(q *Request, r *Result) {
			q.Contract.Coverage.Scope.Service = "rds"
			q.Contract.Coverage.Root = RootRequiredService
			r.Coverage.Root = RootCoverage{Root: RootRequiredService, State: CoverageIncomplete, Reason: CoverageSourceError}
			r.Coverage.Resources = []ResourceObservation{}
			r.Coverage.Sets = []SetCoverage{}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request, result := coverageFixture(t)
			tc.change(&request, &result)
			if err := request.Validate(); err != nil {
				t.Fatalf("request: %v", err)
			}
			if _, err := NewValidator(request.Contract).ValidateResult(result); err != nil {
				t.Fatalf("result: %v", err)
			}
		})
	}
}

func TestCoverageFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Request, *Result)
	}{
		{"missing root", func(_ *Request, r *Result) { r.Coverage.Root.Root = "" }},
		{"missing collection", func(_ *Request, r *Result) { r.Coverage.Sets = r.Coverage.Sets[:1] }},
		{"null resources", func(_ *Request, r *Result) { r.Coverage.Resources = nil }},
		{"null sets", func(_ *Request, r *Result) { r.Coverage.Sets = nil }},
		{"duplicate collection", func(_ *Request, r *Result) { r.Coverage.Sets[1] = r.Coverage.Sets[0] }},
		{"collection outside resource", func(_ *Request, r *Result) { r.Coverage.Sets[1].Set.ResourceID = "sg-other" }},
		{"complete result with incomplete rules", func(_ *Request, r *Result) { r.State = ResultComplete }},
		{"missing core detail marked present", func(_ *Request, r *Result) {
			r.Coverage.Sets[0].State = CoverageIncomplete
			r.Coverage.Sets[0].Reason = CoverageSourceError
		}},
		{"absence under incomplete root", func(_ *Request, r *Result) {
			r.Coverage.Root.State = CoverageIncomplete
			r.Coverage.Root.Reason = CoverageSourceError
			r.Coverage.Resources[0].State = ResourceAbsent
			r.Coverage.Sets = []SetCoverage{}
		}},
		{"wildcard identity", func(_ *Request, r *Result) { r.Coverage.Resources[0].ID = "sg-*" }},
		{"duplicate identity", func(_ *Request, r *Result) {
			r.Coverage.Resources = append(r.Coverage.Resources, r.Coverage.Resources[0])
		}},
		{"capacity exceeded", func(q *Request, r *Result) {
			q.Contract.Coverage.MaxResources = 1
			r.Coverage.Resources = append(r.Coverage.Resources, ResourceObservation{Kind: ResourceEC2Core, ID: "i-2", State: ResourcePresent})
		}},
		{"invalid capacity", func(q *Request, _ *Result) { q.Contract.Coverage.MaxResources = maxCoverageResources + 1 }},
		{"required service claims completion", func(q *Request, r *Result) {
			q.Contract.Coverage.Scope.Service = "rds"
			q.Contract.Coverage.Root = RootRequiredService
			r.Coverage.Root = RootCoverage{Root: RootRequiredService, State: CoverageComplete}
			r.Coverage.Resources = []ResourceObservation{}
			r.Coverage.Sets = []SetCoverage{}
			r.State = ResultComplete
		}},
		{"legacy protocol with coverage", func(q *Request, r *Result) {
			q.ProtocolVersion = ProtocolVersionV1Alpha2
			q.Contract.ProtocolVersion = ProtocolVersionV1Alpha2
			r.ProtocolVersion = ProtocolVersionV1Alpha2
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request, result := coverageFixture(t)
			tc.change(&request, &result)
			if _, err := NewValidator(request.Contract).ValidateResult(result); err == nil {
				t.Fatal("unsafe coverage accepted")
			}
		})
	}
}

func TestCoverageTypedRootBoundaries(t *testing.T) {
	for _, tc := range []struct {
		root  CoverageRoot
		kind  CoverageResourceKind
		count int
	}{
		{RootEC2Resources, ResourceEC2Core, 1},
		{RootEC2Resources, ResourceEC2SecurityGroup, 2},
		{RootLambdaFunctions, ResourceLambdaFunction, 2},
		{RootEKSClusters, ResourceEKSCluster, 2},
		{RootS3Buckets, ResourceS3Bucket, 2},
		{RootS3Buckets, ResourceLambdaFunction, 0},
	} {
		sets, err := requiredSets(tc.root, ResourceObservation{Kind: tc.kind, ID: "resource"})
		if len(sets) != tc.count || (err == nil) != (tc.count > 0) {
			t.Errorf("root=%s kind=%s: sets=%v err=%v", tc.root, tc.kind, sets, err)
		}
	}
}

func TestCoverageValidatorSnapshotsHostPolicy(t *testing.T) {
	request, result := coverageFixture(t)
	validator := NewValidator(request.Contract)
	request.Contract.Coverage.Root = RootS3Buckets
	if _, err := validator.ValidateResult(result); err != nil {
		t.Fatalf("host policy changed after construction: %v", err)
	}
}

func TestCoverageRequestRejectsWrongComponent(t *testing.T) {
	request, _ := coverageFixture(t)
	request.Contract.Coverage.Scope.ComponentID = "other-component"
	if err := request.Validate(); err == nil {
		t.Fatal("request admitted coverage for another component")
	}
}

func TestCoverageRejectsFactsFromIncompleteOrAbsentResources(t *testing.T) {
	for _, absent := range []bool{false, true} {
		request, result := coverageFixture(t)
		ownership := result.Coverage.Sets[1].Set
		if absent {
			result.Coverage.Resources[0].State = ResourceAbsent
			result.Coverage.Sets = []SetCoverage{}
			result.State = ResultComplete
		}
		result.Facts = []Fact{{Kind: "aws_resource", Ownership: &ownership}}
		_, err := NewValidator(request.Contract).ValidateResult(result)
		if err == nil || !strings.Contains(err.Error(), "complete owned collection") {
			t.Errorf("absent=%t: error=%v, want owned-collection refusal", absent, err)
		}
	}
}

func TestCoverageRejectsInvalidUTF8Scope(t *testing.T) {
	bad := string([]byte{'x', 0xff})
	for _, tc := range []struct {
		name   string
		change func(*Request)
	}{
		{"tenant_id", func(r *Request) { r.Contract.Coverage.Scope.TenantID = bad }},
		{"component_id", func(r *Request) {
			r.Contract.Coverage.Scope.ComponentID = bad
			r.Claim.ComponentID = bad
		}},
		{"account_id", func(r *Request) { r.Contract.Coverage.Scope.AccountID = bad }},
		{"partition", func(r *Request) { r.Contract.Coverage.Scope.Partition = bad }},
		{"service", func(r *Request) {
			r.Contract.Coverage.Scope.Service = bad
			r.Contract.Coverage.Root = RootRequiredService
		}},
		{"region", func(r *Request) { r.Contract.Coverage.Scope.Region = bad }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request, _ := coverageFixture(t)
			tc.change(&request)
			err := request.Validate()
			if err == nil || !strings.Contains(err.Error(), "coverage scope "+tc.name+" is invalid") {
				t.Fatalf("Validate() error=%v, want invalid %s", err, tc.name)
			}
		})
	}
}

func TestCoverageRejectsNativeIdentityWireCollision(t *testing.T) {
	first := "sg-" + string([]byte{0xff})
	second := "sg-" + string([]byte{0xfe})
	firstWire, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondWire, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !bytes.Equal(firstWire, secondWire) {
		t.Fatal("test identities must differ natively and collide after JSON encoding")
	}

	request, result := coverageFixture(t)
	result.Coverage.Resources[0].ID = first
	for i := range result.Coverage.Sets {
		result.Coverage.Sets[i].Set.ResourceID = first
	}
	result.Coverage.Resources = append(result.Coverage.Resources,
		ResourceObservation{Kind: ResourceEC2SecurityGroup, ID: second, State: ResourcePresent})
	result.Coverage.Sets = append(result.Coverage.Sets,
		SetCoverage{Set: OwnershipSet{Collection: CollectionEC2Core, ResourceKind: ResourceEC2SecurityGroup, ResourceID: second}, State: CoverageComplete},
		SetCoverage{Set: OwnershipSet{Collection: CollectionEC2SecurityGroupRules, ResourceKind: ResourceEC2SecurityGroup, ResourceID: second}, State: CoverageComplete})
	_, err = NewValidator(request.Contract).ValidateResult(result)
	if err == nil || !strings.Contains(err.Error(), "coverage resource id is invalid") {
		t.Fatalf("ValidateResult() error=%v, want invalid resource id", err)
	}
}
