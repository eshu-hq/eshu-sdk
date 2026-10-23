// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package conformance

import (
	"fmt"
	"sort"
	"strings"

	collector "github.com/eshu-hq/eshu-sdk/collector"
)

// CoreFactContext is a trusted snapshot of host-authorized core emissions for
// one component identity and version. The host owns grant evaluation and
// revocation; a manifest cannot supply this context.
type CoreFactContext struct {
	ComponentID      string
	ComponentVersion string
	Facts            []CoreFactAuthorization
}

// CoreFactAuthorization describes exact emission coverage and trusted reducer
// phase/schema metadata for one reserved fact kind.
type CoreFactAuthorization struct {
	Kind             string
	Emissions        []CoreFactEmission
	ReducerPhases    []string
	PayloadSchemaRef string
}

// CoreFactEmission permits listed schema versions only for one collector kind.
type CoreFactEmission struct {
	CollectorKind  string
	SchemaVersions []string
}

func prepareCoreFacts(req Request) (map[string]CoreFactAuthorization, error) {
	context := req.CoreFacts
	if context == nil {
		return nil, nil
	}
	if !canonicalIdentifier(context.ComponentID) || context.ComponentID != req.Manifest.Metadata.ID ||
		!canonicalVersion(context.ComponentVersion) || context.ComponentVersion != req.Manifest.Metadata.Version {
		return nil, fmt.Errorf("core fact context component identity or version does not match manifest")
	}
	if len(context.Facts) == 0 || len(req.ReservedFactKinds) == 0 {
		return nil, fmt.Errorf("core fact context requires coverage and a reserved fact inventory")
	}
	reserved := make(map[string]struct{}, len(req.ReservedFactKinds))
	for _, kind := range req.ReservedFactKinds {
		if !canonicalIdentifier(kind) {
			return nil, fmt.Errorf("reserved fact kind %q is not canonical", kind)
		}
		if _, exists := reserved[kind]; exists {
			return nil, fmt.Errorf("reserved fact kind %q is duplicated", kind)
		}
		reserved[kind] = struct{}{}
	}
	collectors := make(map[string]struct{}, len(req.Manifest.Spec.CollectorKinds))
	for _, kind := range req.Manifest.Spec.CollectorKinds {
		if !canonicalIdentifier(kind) {
			return nil, fmt.Errorf("manifest collector kind %q is not canonical", kind)
		}
		if _, exists := collectors[kind]; exists {
			return nil, fmt.Errorf("manifest collector kind %q is duplicated", kind)
		}
		collectors[kind] = struct{}{}
	}
	declarations := make(map[string]FactFamily, len(req.Manifest.Spec.EmittedFacts))
	for _, fact := range req.Manifest.Spec.EmittedFacts {
		if !canonicalIdentifier(fact.Kind) {
			return nil, fmt.Errorf("manifest fact kind %q is not canonical", fact.Kind)
		}
		if _, exists := declarations[fact.Kind]; exists {
			return nil, fmt.Errorf("manifest fact kind %q is duplicated", fact.Kind)
		}
		declarations[fact.Kind] = fact
	}
	coverage := make(map[string]CoreFactAuthorization, len(context.Facts))
	for _, authorization := range context.Facts {
		declaration, declared := declarations[authorization.Kind]
		_, hostOwned := reserved[authorization.Kind]
		if !canonicalIdentifier(authorization.Kind) || !declared || !hostOwned {
			return nil, fmt.Errorf("core fact authorization kind %q is undeclared or not reserved", authorization.Kind)
		}
		if _, exists := coverage[authorization.Kind]; exists {
			return nil, fmt.Errorf("core fact authorization kind %q is duplicated", authorization.Kind)
		}
		if err := validateCoreAuthorization(authorization, declaration, collectors); err != nil {
			return nil, err
		}
		coverage[authorization.Kind] = authorization
	}
	for _, declaration := range req.Manifest.Spec.EmittedFacts {
		if _, coreKind := reserved[declaration.Kind]; !coreKind {
			continue
		}
		authorization, covered := coverage[declaration.Kind]
		if !covered {
			return nil, fmt.Errorf("reserved fact kind %q has no core authorization", declaration.Kind)
		}
		versions := make(map[string]struct{}, len(declaration.SchemaVersions))
		for _, version := range declaration.SchemaVersions {
			if !canonicalVersion(version) {
				return nil, fmt.Errorf("fact kind %q schema version %q is not canonical", declaration.Kind, version)
			}
			if _, duplicate := versions[version]; duplicate {
				return nil, fmt.Errorf("fact kind %q schema version %q is duplicated", declaration.Kind, version)
			}
			versions[version] = struct{}{}
			if !coversAnyCollector(authorization, version) {
				return nil, fmt.Errorf("fact kind %q schema version %q has no collector coverage", declaration.Kind, version)
			}
		}
	}
	return coverage, nil
}

func validateCoreAuthorization(auth CoreFactAuthorization, declared FactFamily, collectors map[string]struct{}) error {
	if len(auth.Emissions) == 0 || len(auth.ReducerPhases) == 0 {
		return fmt.Errorf("core fact kind %q requires emissions and reducer phases", auth.Kind)
	}
	phaseSet := make(map[string]struct{}, len(auth.ReducerPhases))
	for _, phase := range auth.ReducerPhases {
		if !canonicalPhase(phase) {
			return fmt.Errorf("core fact kind %q reducer phase %q is not canonical", auth.Kind, phase)
		}
		if _, duplicate := phaseSet[phase]; duplicate {
			return fmt.Errorf("core fact kind %q reducer phase %q is duplicated", auth.Kind, phase)
		}
		phaseSet[phase] = struct{}{}
	}
	emissionSet := make(map[string]struct{}, len(auth.Emissions))
	for _, emission := range auth.Emissions {
		_, declaredCollector := collectors[emission.CollectorKind]
		if !canonicalIdentifier(emission.CollectorKind) || !declaredCollector || len(emission.SchemaVersions) == 0 {
			return fmt.Errorf("core fact kind %q collector emission is invalid", auth.Kind)
		}
		if _, duplicate := emissionSet[emission.CollectorKind]; duplicate {
			return fmt.Errorf("core fact kind %q collector kind %q is duplicated", auth.Kind, emission.CollectorKind)
		}
		emissionSet[emission.CollectorKind] = struct{}{}
		versions := make(map[string]struct{}, len(emission.SchemaVersions))
		for _, version := range emission.SchemaVersions {
			if !canonicalVersion(version) || !containsString(declared.SchemaVersions, version) {
				return fmt.Errorf("core fact kind %q schema version %q is undeclared or not canonical", auth.Kind, version)
			}
			if _, duplicate := versions[version]; duplicate {
				return fmt.Errorf("core fact kind %q schema version %q is duplicated", auth.Kind, version)
			}
			versions[version] = struct{}{}
		}
	}
	return nil
}

func canonicalIdentifier(value string) bool {
	return value != "" && strings.TrimSpace(value) == value && identifierPattern.MatchString(value)
}

func canonicalVersion(value string) bool {
	return value != "" && strings.TrimSpace(value) == value && semverPattern.MatchString(value)
}

func canonicalPhase(value string) bool {
	return value != "" && strings.TrimSpace(value) == value
}

func containsString(values []string, sought string) bool {
	for _, value := range values {
		if value == sought {
			return true
		}
	}
	return false
}

func coversAnyCollector(auth CoreFactAuthorization, version string) bool {
	for _, emission := range auth.Emissions {
		if containsString(emission.SchemaVersions, version) {
			return true
		}
	}
	return false
}

func coversTuple(auth CoreFactAuthorization, collectorKind, version string) bool {
	for _, emission := range auth.Emissions {
		if emission.CollectorKind == collectorKind && containsString(emission.SchemaVersions, version) {
			return true
		}
	}
	return false
}

func addCorePhaseFindings(report *Report, manifest Manifest, coverage map[string]CoreFactAuthorization) {
	trusted := make(map[string]struct{})
	for _, auth := range coverage {
		for _, phase := range auth.ReducerPhases {
			trusted[phase] = struct{}{}
		}
	}
	declared := make(map[string]struct{})
	duplicates := make(map[string]struct{})
	for _, phase := range manifest.Spec.ConsumerContracts.Reducer.Phases {
		if _, duplicate := declared[phase]; duplicate {
			duplicates[phase] = struct{}{}
		}
		declared[phase] = struct{}{}
	}
	for _, phase := range sortedKeys(duplicates) {
		addBlockingFinding(report, FindingMissingReducerConsumer, fmt.Sprintf("reducer phase %q is duplicated", phase), -1)
	}
	for _, phase := range sortedKeys(trusted) {
		if _, ok := declared[phase]; !ok {
			addBlockingFinding(report, FindingMissingReducerConsumer, fmt.Sprintf("trusted reducer phase %q is missing", phase), -1)
		}
	}
	for _, phase := range sortedKeys(declared) {
		if phase == SourceEvidenceOnlyReducerPhase {
			continue
		}
		if _, ok := trusted[phase]; !ok {
			addBlockingFinding(report, FindingMissingReducerConsumer, fmt.Sprintf("reducer phase %q has no trusted core authorization", phase), -1)
		}
	}
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func checkCoreFixture(req Request, coverage map[string]CoreFactAuthorization, fixture collector.Result) error {
	if fixture.Claim.ComponentID != req.Manifest.Metadata.ID || fixture.Claim.ComponentID != req.CoreFacts.ComponentID {
		return fmt.Errorf("fixture claim component does not match trusted context")
	}
	if !containsString(req.Manifest.Spec.CollectorKinds, fixture.Claim.CollectorKind) {
		return fmt.Errorf("fixture claim collector kind is not declared by manifest")
	}
	for _, fact := range fixture.Facts {
		if !containsString(req.ReservedFactKinds, fact.Kind) {
			continue
		}
		auth, ok := coverage[fact.Kind]
		if !ok || !coversTuple(auth, fixture.Claim.CollectorKind, fact.SchemaVersion) {
			return fmt.Errorf("fixture reserved fact kind %q lacks exact collector/schema authorization", fact.Kind)
		}
	}
	return nil
}

func requireCorePayloadSchemas(req Request, coverage map[string]CoreFactAuthorization) error {
	kinds := make([]string, 0, len(coverage))
	for kind := range coverage {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		auth := coverage[kind]
		if auth.PayloadSchemaRef != "" && !canonicalIdentifier(auth.PayloadSchemaRef) {
			return fmt.Errorf("core fact kind %q payloadSchemaRef is not canonical", kind)
		}
		for _, declared := range req.Manifest.Spec.EmittedFacts {
			if declared.Kind == kind && declared.PayloadSchemaRef != "" && declared.PayloadSchemaRef != auth.PayloadSchemaRef {
				return fmt.Errorf("core fact kind %q payloadSchemaRef conflicts with manifest", kind)
			}
		}
		if auth.PayloadSchemaRef == "" {
			continue
		}
		if _, ok := req.PayloadSchemas[kind]; !ok {
			return fmt.Errorf("core fact kind %q trusted payloadSchemaRef has no PayloadSchemas entry", kind)
		}
	}
	return nil
}
