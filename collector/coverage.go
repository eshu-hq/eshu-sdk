// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"fmt"
	"strings"
)

// CoverageCollection is a host-recognized unit of AWS replacement ownership.
// Unknown collections are invalid so a producer cannot invent a selector.
type CoverageCollection string

const (
	CollectionAWSService                CoverageCollection = "aws_service"
	CollectionEC2Core                   CoverageCollection = "aws_ec2_core"
	CollectionEC2SecurityGroupRules     CoverageCollection = "aws_ec2_security_group_rules"
	CollectionLambdaFunction            CoverageCollection = "aws_lambda_function"
	CollectionLambdaEventSourceMappings CoverageCollection = "aws_lambda_event_source_mappings"
	CollectionEKSCluster                CoverageCollection = "aws_eks_cluster"
	CollectionEKSIAMEnrichment          CoverageCollection = "aws_eks_iam_enrichment"
	CollectionS3Bucket                  CoverageCollection = "aws_s3_bucket"
	CollectionS3BucketEncryption        CoverageCollection = "aws_s3_bucket_encryption"
)

// CoverageScope is the host-authorized AWS partition of one claim. The host
// must compare these fields with its own tenant, component, and cloud grants.
type CoverageScope struct {
	TenantID    string `json:"tenant_id"`
	ComponentID string `json:"component_id"`
	AccountID   string `json:"account_id"`
	Partition   string `json:"partition"`
	Service     string `json:"service"`
	Region      string `json:"region"`
}

// OwnershipSet is one typed, disjoint collection or resource. ResourceID is
// required only for resource-scoped collections, and has no query semantics.
type OwnershipSet struct {
	Collection CoverageCollection `json:"collection"`
	ResourceID string             `json:"resource_id,omitempty"`
}

// CoverageContract is the host-issued inventory of sets this claim must report.
// The host must derive it independently of collector-produced evidence.
type CoverageContract struct {
	Scope CoverageScope  `json:"scope"`
	Sets  []OwnershipSet `json:"sets"`
}

// CoverageState reports whether a declared set was completely observed.
type CoverageState string

const (
	CoverageComplete   CoverageState = "complete"
	CoverageIncomplete CoverageState = "incomplete"
)

// SetCoverage reports one host-declared ownership set. An incomplete set may
// not be treated as a deletion signal or a current complete observation.
type SetCoverage struct {
	Set    OwnershipSet   `json:"set"`
	State  CoverageState  `json:"state"`
	Reason CoverageReason `json:"reason,omitempty"`
}

// CoverageReason is a bounded explanation of incomplete source coverage.
type CoverageReason string

const (
	CoverageNotObserved      CoverageReason = "not_observed"
	CoveragePermissionDenied CoverageReason = "permission_denied"
	CoverageThrottled        CoverageReason = "throttled"
	CoverageSourceError      CoverageReason = "source_error"
	CoverageUnsupported      CoverageReason = "unsupported"
)

// CoverageReport accounts for every declared set in one result. It cannot
// select previous generations or carry historical payloads.
type CoverageReport struct {
	Sets []SetCoverage `json:"sets"`
}

const maxCoverageSets = 10000

func validateCoverageContract(contract *CoverageContract) (map[OwnershipSet]struct{}, error) {
	if contract == nil {
		return nil, fmt.Errorf("coverage contract is required")
	}
	scope := contract.Scope
	for name, value := range map[string]string{
		"tenant_id": scope.TenantID, "component_id": scope.ComponentID, "account_id": scope.AccountID,
		"partition": scope.Partition, "service": scope.Service,
		"region": scope.Region,
	} {
		if invalidCoverageValue(value) {
			return nil, fmt.Errorf("coverage scope %s is invalid", name)
		}
	}
	if len(contract.Sets) == 0 || len(contract.Sets) > maxCoverageSets {
		return nil, fmt.Errorf("coverage contract requires 1..%d sets", maxCoverageSets)
	}
	seen := make(map[OwnershipSet]struct{}, len(contract.Sets))
	for _, set := range contract.Sets {
		if err := validateOwnershipSet(scope.Service, set); err != nil {
			return nil, err
		}
		if _, duplicate := seen[set]; duplicate {
			return nil, fmt.Errorf("coverage contract has duplicate ownership set")
		}
		seen[set] = struct{}{}
	}
	return seen, nil
}

func validateOwnershipSet(service string, set OwnershipSet) error {
	var wantedService string
	var resource bool
	switch set.Collection {
	case CollectionAWSService:
		wantedService = service
	case CollectionEC2Core, CollectionEC2SecurityGroupRules:
		wantedService = "ec2"
	case CollectionLambdaFunction, CollectionLambdaEventSourceMappings:
		wantedService, resource = "lambda", true
	case CollectionEKSCluster, CollectionEKSIAMEnrichment:
		wantedService, resource = "eks", true
	case CollectionS3Bucket, CollectionS3BucketEncryption:
		wantedService, resource = "s3", true
	default:
		return fmt.Errorf("coverage collection %q is unsupported", set.Collection)
	}
	if service != wantedService {
		return fmt.Errorf("coverage collection %q is outside service %q", set.Collection, service)
	}
	if resource && invalidCoverageValue(set.ResourceID) {
		return fmt.Errorf("coverage collection %q requires resource_id", set.Collection)
	}
	if !resource && set.ResourceID != "" {
		return fmt.Errorf("coverage collection %q cannot select resource_id", set.Collection)
	}
	return nil
}

func invalidCoverageValue(value string) bool {
	return value == "" || len(value) > 2048 || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\x00\n\r\t")
}

func validateCoverageReport(contract *CoverageContract, result Result) error {
	declared, err := validateCoverageContract(contract)
	if err != nil {
		return err
	}
	if contract.Scope.ComponentID != result.Claim.ComponentID {
		return fmt.Errorf("coverage component does not match claim")
	}
	if result.State == ResultUnchanged || result.State == ResultRetryable || result.State == ResultTerminal {
		if result.Coverage != nil {
			return fmt.Errorf("%s result cannot report coverage", result.State)
		}
		return nil
	}
	if result.Coverage == nil || len(result.Coverage.Sets) != len(declared) {
		return fmt.Errorf("coverage report must account for every declared set")
	}
	seen := make(map[OwnershipSet]CoverageState, len(declared))
	incomplete := false
	for _, entry := range result.Coverage.Sets {
		if _, ok := declared[entry.Set]; !ok {
			return fmt.Errorf("coverage report names undeclared set")
		}
		if _, duplicate := seen[entry.Set]; duplicate {
			return fmt.Errorf("coverage report duplicates ownership set")
		}
		switch entry.State {
		case CoverageComplete:
			if entry.Reason != "" {
				return fmt.Errorf("complete coverage must not have reason")
			}
		case CoverageIncomplete:
			switch entry.Reason {
			case CoverageNotObserved, CoveragePermissionDenied, CoverageThrottled, CoverageSourceError, CoverageUnsupported:
			default:
				return fmt.Errorf("incomplete coverage requires recognized reason")
			}
			incomplete = true
		default:
			return fmt.Errorf("coverage state %q is unsupported", entry.State)
		}
		seen[entry.Set] = entry.State
	}
	if result.State == ResultComplete && incomplete || result.State == ResultPartial && !incomplete {
		return fmt.Errorf("result state and coverage completeness disagree")
	}
	for _, fact := range result.Facts {
		if fact.Ownership == nil {
			return fmt.Errorf("fact %q requires ownership set", fact.Kind)
		}
		if fact.Ownership.Collection == CollectionAWSService {
			return fmt.Errorf("fact %q cannot belong to service coverage sentinel", fact.Kind)
		}
		if _, ok := declared[*fact.Ownership]; !ok {
			return fmt.Errorf("fact %q names undeclared ownership set", fact.Kind)
		}
	}
	return nil
}
