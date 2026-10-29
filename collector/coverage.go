// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"fmt"
	"strings"
)

// CoverageCollection is a fixed AWS fact collection, never a database selector.
type CoverageCollection string

// Collection constants identify the fixed core and enrichment fact families.
const (
	CollectionEC2Core                   CoverageCollection = "aws_ec2_core"
	CollectionEC2SecurityGroupRules     CoverageCollection = "aws_ec2_security_group_rules"
	CollectionLambdaFunction            CoverageCollection = "aws_lambda_function"
	CollectionLambdaEventSourceMappings CoverageCollection = "aws_lambda_event_source_mappings"
	CollectionEKSCluster                CoverageCollection = "aws_eks_cluster"
	CollectionEKSIAMEnrichment          CoverageCollection = "aws_eks_iam_enrichment"
	CollectionS3Bucket                  CoverageCollection = "aws_s3_bucket"
	CollectionS3BucketEncryption        CoverageCollection = "aws_s3_bucket_encryption"
)

// CoverageRoot is the host-selected discovery root for one AWS service claim.
type CoverageRoot string

// Root constants name the bounded discovery shape selected by the host.
const (
	RootEC2Resources    CoverageRoot = "aws_ec2_resources"
	RootLambdaFunctions CoverageRoot = "aws_lambda_functions"
	RootEKSClusters     CoverageRoot = "aws_eks_clusters"
	RootS3Buckets       CoverageRoot = "aws_s3_buckets"
	// RootRequiredService reports a failed required service with no resource
	// ownership. It cannot authorize complete publication or carry facts.
	RootRequiredService CoverageRoot = "aws_required_service"
)

// CoverageResourceKind names a typed identity within a discovery root.
type CoverageResourceKind string

// Resource constants distinguish identities within each root.
const (
	ResourceEC2Core          CoverageResourceKind = "aws_ec2_resource"
	ResourceEC2SecurityGroup CoverageResourceKind = "aws_ec2_security_group"
	ResourceLambdaFunction   CoverageResourceKind = "aws_lambda_function"
	ResourceEKSCluster       CoverageResourceKind = "aws_eks_cluster"
	ResourceS3Bucket         CoverageResourceKind = "aws_s3_bucket"
)

// CoverageScope is the host-authorized cloud and component partition. The host
// must compare it with its own tenant, account, region, and component grants.
type CoverageScope struct {
	TenantID    string `json:"tenant_id"`
	ComponentID string `json:"component_id"`
	AccountID   string `json:"account_id"`
	Partition   string `json:"partition"`
	Service     string `json:"service"`
	Region      string `json:"region"`
}

// OwnershipSet identifies one whole-fact collection for one typed resource.
// A producer cannot put a selector, prior generation, or payload in this type.
type OwnershipSet struct {
	Collection   CoverageCollection   `json:"collection"`
	ResourceKind CoverageResourceKind `json:"resource_kind"`
	ResourceID   string               `json:"resource_id"`
}

// CoverageContract is the host-issued root and capacity policy for one claim.
// The root determines the fixed required core and enrichment collections.
type CoverageContract struct {
	Scope        CoverageScope `json:"scope"`
	Root         CoverageRoot  `json:"root"`
	MaxResources int           `json:"max_resources"`
}

// CoverageState reports whether a root or collection was fully observed.
type CoverageState string

// Coverage states distinguish complete from incomplete observation.
const (
	CoverageComplete   CoverageState = "complete"
	CoverageIncomplete CoverageState = "incomplete"
)

// CoverageReason gives a bounded cause for incomplete observation.
type CoverageReason string

// Coverage reasons are bounded source failure classes.
const (
	CoverageNotObserved      CoverageReason = "not_observed"
	CoveragePermissionDenied CoverageReason = "permission_denied"
	CoverageThrottled        CoverageReason = "throttled"
	CoverageSourceError      CoverageReason = "source_error"
	CoverageUnsupported      CoverageReason = "unsupported"
	CoverageCapacityExceeded CoverageReason = "capacity_exceeded"
)

// RootCoverage reports complete or incomplete source enumeration. A complete
// declaration remains untrusted until the host verifies the producer and source.
type RootCoverage struct {
	Root   CoverageRoot   `json:"root"`
	State  CoverageState  `json:"state"`
	Reason CoverageReason `json:"reason,omitempty"`
}

// ResourceState distinguishes a seen resource, a proven absence, and a listed
// resource whose required core detail could not be collected.
type ResourceState string

// Resource states separate source presence from absence and failed detail.
const (
	ResourcePresent    ResourceState = "present"
	ResourceAbsent     ResourceState = "absent"
	ResourceIncomplete ResourceState = "incomplete"
)

// ResourceObservation is one typed source identity, not a deletion command.
// The host alone decides whether an absent observation can retire old evidence.
type ResourceObservation struct {
	Kind   CoverageResourceKind `json:"kind"`
	ID     string               `json:"id"`
	State  ResourceState        `json:"state"`
	Reason CoverageReason       `json:"reason,omitempty"`
}

// SetCoverage reports core or enrichment completeness for one observed resource.
type SetCoverage struct {
	Set    OwnershipSet   `json:"set"`
	State  CoverageState  `json:"state"`
	Reason CoverageReason `json:"reason,omitempty"`
}

// CoverageReport accounts for root discovery, resource identities, and each
// required collection. It contains no historical payload or generation IDs.
type CoverageReport struct {
	Root      RootCoverage          `json:"root"`
	Resources []ResourceObservation `json:"resources"`
	Sets      []SetCoverage         `json:"sets"`
}

const maxCoverageResources = 10000

func validateCoverageContract(contract *CoverageContract) error {
	if contract == nil {
		return fmt.Errorf("coverage contract is required")
	}
	for name, value := range map[string]string{
		"tenant_id": contract.Scope.TenantID, "component_id": contract.Scope.ComponentID,
		"account_id": contract.Scope.AccountID, "partition": contract.Scope.Partition,
		"service": contract.Scope.Service, "region": contract.Scope.Region,
	} {
		if invalidCoverageValue(value) {
			return fmt.Errorf("coverage scope %s is invalid", name)
		}
	}
	if contract.MaxResources < 1 || contract.MaxResources > maxCoverageResources {
		return fmt.Errorf("coverage max_resources must be 1..%d", maxCoverageResources)
	}
	var want CoverageRoot
	switch contract.Scope.Service {
	case "ec2":
		want = RootEC2Resources
	case "lambda":
		want = RootLambdaFunctions
	case "eks":
		want = RootEKSClusters
	case "s3":
		want = RootS3Buckets
	default:
		want = RootRequiredService
	}
	if contract.Root != want {
		return fmt.Errorf("coverage root %q is invalid for service %q", contract.Root, contract.Scope.Service)
	}
	return nil
}

func invalidCoverageValue(value string) bool {
	return value == "" || len(value) > 2048 || strings.TrimSpace(value) != value || strings.ContainsAny(value, "\x00\n\r\t*?%")
}

func validCoverageReason(reason CoverageReason) bool {
	switch reason {
	case CoverageNotObserved, CoveragePermissionDenied, CoverageThrottled, CoverageSourceError, CoverageUnsupported, CoverageCapacityExceeded:
		return true
	default:
		return false
	}
}

func validateCoverageState(state CoverageState, reason CoverageReason) error {
	switch state {
	case CoverageComplete:
		if reason != "" {
			return fmt.Errorf("complete coverage cannot have reason")
		}
	case CoverageIncomplete:
		if !validCoverageReason(reason) {
			return fmt.Errorf("incomplete coverage requires recognized reason")
		}
	default:
		return fmt.Errorf("coverage state %q is unsupported", state)
	}
	return nil
}

func requiredSets(root CoverageRoot, resource ResourceObservation) ([]OwnershipSet, error) {
	set := func(collection CoverageCollection) OwnershipSet {
		return OwnershipSet{Collection: collection, ResourceKind: resource.Kind, ResourceID: resource.ID}
	}
	switch root {
	case RootEC2Resources:
		switch resource.Kind {
		case ResourceEC2Core:
			return []OwnershipSet{set(CollectionEC2Core)}, nil
		case ResourceEC2SecurityGroup:
			return []OwnershipSet{set(CollectionEC2Core), set(CollectionEC2SecurityGroupRules)}, nil
		}
	case RootLambdaFunctions:
		if resource.Kind == ResourceLambdaFunction {
			return []OwnershipSet{set(CollectionLambdaFunction), set(CollectionLambdaEventSourceMappings)}, nil
		}
	case RootEKSClusters:
		if resource.Kind == ResourceEKSCluster {
			return []OwnershipSet{set(CollectionEKSCluster), set(CollectionEKSIAMEnrichment)}, nil
		}
	case RootS3Buckets:
		if resource.Kind == ResourceS3Bucket {
			return []OwnershipSet{set(CollectionS3Bucket), set(CollectionS3BucketEncryption)}, nil
		}
	}
	return nil, fmt.Errorf("resource kind %q is outside root %q", resource.Kind, root)
}

func isCoreCollection(collection CoverageCollection) bool {
	switch collection {
	case CollectionEC2Core, CollectionLambdaFunction, CollectionEKSCluster, CollectionS3Bucket:
		return true
	default:
		return false
	}
}

func validateCoverageReport(contract *CoverageContract, result Result) error {
	if err := validateCoverageContract(contract); err != nil {
		return err
	}
	if contract.Scope.ComponentID != result.Claim.ComponentID {
		return fmt.Errorf("coverage component does not match claim")
	}
	if result.Generation.TrustedDeletion {
		return fmt.Errorf("trusted_deletion cannot override coverage")
	}
	if result.State == ResultUnchanged || result.State == ResultRetryable || result.State == ResultTerminal {
		if result.Coverage != nil {
			return fmt.Errorf("%s result cannot report coverage", result.State)
		}
		return nil
	}
	report := result.Coverage
	if report == nil {
		return fmt.Errorf("coverage report is required")
	}
	if report.Resources == nil || report.Sets == nil {
		return fmt.Errorf("coverage resources and sets must be explicit arrays")
	}
	if report.Root.Root != contract.Root {
		return fmt.Errorf("coverage root does not match host contract")
	}
	if err := validateCoverageState(report.Root.State, report.Root.Reason); err != nil {
		return err
	}
	if contract.Root == RootRequiredService && report.Root.State != CoverageIncomplete {
		return fmt.Errorf("required service root cannot declare completion")
	}
	if len(report.Resources) > contract.MaxResources {
		return fmt.Errorf("coverage resource capacity exceeded")
	}
	if len(report.Sets) > 2*contract.MaxResources {
		return fmt.Errorf("coverage set capacity exceeded")
	}
	type resourceKey struct {
		kind CoverageResourceKind
		id   string
	}
	resources := make(map[resourceKey]ResourceState, len(report.Resources))
	expected := make(map[OwnershipSet]ResourceState, len(report.Resources)*2)
	incomplete := report.Root.State == CoverageIncomplete
	for _, resource := range report.Resources {
		if invalidCoverageValue(resource.ID) {
			return fmt.Errorf("coverage resource id is invalid")
		}
		key := resourceKey{resource.Kind, resource.ID}
		if _, duplicate := resources[key]; duplicate {
			return fmt.Errorf("coverage resource is duplicated")
		}
		sets, err := requiredSets(contract.Root, resource)
		if err != nil {
			return err
		}
		switch resource.State {
		case ResourcePresent, ResourceAbsent:
			if resource.Reason != "" {
				return fmt.Errorf("present or absent resource cannot have reason")
			}
			if resource.State == ResourceAbsent && report.Root.State != CoverageComplete {
				return fmt.Errorf("absence requires complete root discovery")
			}
		case ResourceIncomplete:
			if !validCoverageReason(resource.Reason) {
				return fmt.Errorf("incomplete resource requires recognized reason")
			}
			incomplete = true
		default:
			return fmt.Errorf("resource state %q is unsupported", resource.State)
		}
		resources[key] = resource.State
		if resource.State != ResourceAbsent {
			for _, set := range sets {
				expected[set] = resource.State
			}
		}
	}
	if len(report.Sets) != len(expected) {
		return fmt.Errorf("coverage report must account for every required collection")
	}
	states := make(map[OwnershipSet]CoverageState, len(expected))
	for _, entry := range report.Sets {
		resourceState, ok := expected[entry.Set]
		if !ok {
			return fmt.Errorf("coverage report names undeclared resource collection")
		}
		if _, duplicate := states[entry.Set]; duplicate {
			return fmt.Errorf("coverage report duplicates resource collection")
		}
		if err := validateCoverageState(entry.State, entry.Reason); err != nil {
			return err
		}
		if resourceState == ResourceIncomplete && entry.State != CoverageIncomplete {
			return fmt.Errorf("incomplete resource cannot declare complete collection")
		}
		if resourceState == ResourcePresent && isCoreCollection(entry.Set.Collection) && entry.State != CoverageComplete {
			return fmt.Errorf("missing core detail requires incomplete resource state")
		}
		if entry.State == CoverageIncomplete {
			incomplete = true
		}
		states[entry.Set] = entry.State
	}
	if result.State == ResultComplete && incomplete || result.State == ResultPartial && !incomplete {
		return fmt.Errorf("result state and coverage completeness disagree")
	}
	for _, fact := range result.Facts {
		if fact.Ownership == nil {
			return fmt.Errorf("fact %q requires ownership set", fact.Kind)
		}
		if states[*fact.Ownership] != CoverageComplete {
			return fmt.Errorf("fact %q requires complete owned collection", fact.Kind)
		}
	}
	return nil
}
