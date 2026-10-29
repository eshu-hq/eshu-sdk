// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"bytes"
	"encoding/json"
	"os"
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
		Sets:  []OwnershipSet{{Collection: CollectionAWSService}, {Collection: CollectionEC2Core}, {Collection: CollectionEC2SecurityGroupRules}},
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

func TestCoverageRejectsMissingDeclaredSet(t *testing.T) {
	request, result := coverageFixture(t)
	result.Coverage.Sets = result.Coverage.Sets[:2]
	if _, err := NewValidator(request.Contract).ValidateResult(result); err == nil {
		t.Fatal("missing coverage report was accepted")
	}
}

func TestCoverageFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Request, *Result)
	}{
		{"duplicate set", func(_ *Request, r *Result) { r.Coverage.Sets[2] = r.Coverage.Sets[1] }},
		{"undeclared set", func(_ *Request, r *Result) { r.Coverage.Sets[2].Set.Collection = CollectionS3BucketEncryption }},
		{"complete with incomplete set", func(_ *Request, r *Result) { r.State = ResultComplete }},
		{"partial without incomplete set", func(_ *Request, r *Result) {
			r.Coverage.Sets[2].State = CoverageComplete
			r.Coverage.Sets[2].Reason = ""
		}},
		{"unknown collection", func(q *Request, _ *Result) { q.Contract.Coverage.Sets[1].Collection = "database_where" }},
		{"cross service collection", func(q *Request, _ *Result) { q.Contract.Coverage.Sets[1].Collection = CollectionS3Bucket }},
		{"duplicate declaration", func(q *Request, _ *Result) { q.Contract.Coverage.Sets[2] = q.Contract.Coverage.Sets[1] }},
		{"missing component", func(q *Request, _ *Result) { q.Contract.Coverage.Scope.ComponentID = "other" }},
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

func TestCoverageResourceBoundaries(t *testing.T) {
	for _, tc := range []struct {
		service    string
		collection CoverageCollection
		resource   string
		valid      bool
	}{
		{"lambda", CollectionLambdaEventSourceMappings, "arn:aws:lambda:us-east-1:123456789012:function:orders", true},
		{"lambda", CollectionLambdaEventSourceMappings, "", false},
		{"eks", CollectionEKSIAMEnrichment, "cluster-a", true},
		{"s3", CollectionS3BucketEncryption, "bucket-a", true},
		{"s3", CollectionS3BucketEncryption, "", false},
		{"ec2", CollectionEC2SecurityGroupRules, "sg-1", false},
	} {
		err := validateOwnershipSet(tc.service, OwnershipSet{Collection: tc.collection, ResourceID: tc.resource})
		if (err == nil) != tc.valid {
			t.Errorf("%s/%s/%s: error=%v, valid=%t", tc.service, tc.collection, tc.resource, err, tc.valid)
		}
	}
}

func TestCoverageValidatorSnapshotsHostDeclaration(t *testing.T) {
	request, result := coverageFixture(t)
	validator := NewValidator(request.Contract)
	request.Contract.Coverage.Sets[1] = OwnershipSet{Collection: CollectionS3Bucket, ResourceID: "bucket-a"}
	if _, err := validator.ValidateResult(result); err != nil {
		t.Fatalf("host declaration changed after validator construction: %v", err)
	}
}

func TestCoverageRequestRejectsWrongComponent(t *testing.T) {
	request, _ := coverageFixture(t)
	request.Contract.Coverage.Scope.ComponentID = "other-component"
	if err := request.Validate(); err == nil {
		t.Fatal("request admitted coverage for another component")
	}
}
