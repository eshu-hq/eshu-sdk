// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import "encoding/json"

// CoverageJSONSchema returns the opt-in v1alpha3 result wire schema.
func CoverageJSONSchema() ([]byte, error) {
	raw, err := json.MarshalIndent(coverageResultSchema(), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func coverageResultSchema() map[string]any {
	schema := resultSchema()
	schema["$id"] = "https://eshu.dev/schemas/collector-sdk/v1alpha3/result.schema.json"
	properties := schema["properties"].(map[string]any)
	properties["protocol_version"] = map[string]any{"const": ProtocolVersionV1Alpha3}
	properties["coverage"] = ref("#/$defs/coverage_report")
	schema["allOf"] = []any{map[string]any{
		"if": map[string]any{"properties": map[string]any{"state": map[string]any{
			"enum": []string{string(ResultComplete), string(ResultPartial)},
		}}},
		"then": map[string]any{"required": []string{"coverage"}},
		"else": map[string]any{"not": map[string]any{"required": []string{"coverage"}}},
	}}
	defs := schema["$defs"].(map[string]any)
	defs["generation"].(map[string]any)["properties"].(map[string]any)["trusted_deletion"] = map[string]any{"const": false}
	fact := defs["fact"].(map[string]any)
	fact["properties"].(map[string]any)["ownership"] = ref("#/$defs/ownership_set")
	fact["required"] = append(fact["required"].([]string), "ownership")
	defs["ownership_set"] = ownershipSetSchema()
	defs["coverage_reason"] = coverageReasonSchema()
	defs["root_coverage"] = objectSchema([]string{"root", "state"}, map[string]any{
		"root":   coverageRootSchema(),
		"state":  coverageStateSchema(),
		"reason": ref("#/$defs/coverage_reason"),
	})
	defs["resource_observation"] = objectSchema([]string{"kind", "id", "state"}, map[string]any{
		"kind":   resourceKindSchema(),
		"id":     nonEmptyString(),
		"state":  map[string]any{"enum": []string{string(ResourcePresent), string(ResourceAbsent), string(ResourceIncomplete)}},
		"reason": ref("#/$defs/coverage_reason"),
	})
	defs["set_coverage"] = objectSchema([]string{"set", "state"}, map[string]any{
		"set":    ref("#/$defs/ownership_set"),
		"state":  coverageStateSchema(),
		"reason": ref("#/$defs/coverage_reason"),
	})
	defs["coverage_report"] = objectSchema([]string{"root", "resources", "sets"}, map[string]any{
		"root":      ref("#/$defs/root_coverage"),
		"resources": map[string]any{"type": "array", "maxItems": maxCoverageResources, "items": ref("#/$defs/resource_observation")},
		"sets":      map[string]any{"type": "array", "maxItems": 2 * maxCoverageResources, "items": ref("#/$defs/set_coverage")},
	})
	return schema
}

func coverageScopeSchema() map[string]any {
	properties := map[string]any{}
	for _, field := range []string{"tenant_id", "component_id", "account_id", "partition", "service", "region"} {
		properties[field] = nonEmptyString()
	}
	return objectSchema([]string{"tenant_id", "component_id", "account_id", "partition", "service", "region"}, properties)
}

func coverageRootSchema() map[string]any {
	return map[string]any{"enum": []string{
		string(RootEC2Resources), string(RootLambdaFunctions), string(RootEKSClusters),
		string(RootS3Buckets), string(RootRequiredService),
	}}
}

func resourceKindSchema() map[string]any {
	return map[string]any{"enum": []string{
		string(ResourceEC2Core), string(ResourceEC2SecurityGroup), string(ResourceLambdaFunction),
		string(ResourceEKSCluster), string(ResourceS3Bucket),
	}}
}

func coverageStateSchema() map[string]any {
	return map[string]any{"enum": []string{string(CoverageComplete), string(CoverageIncomplete)}}
}

func coverageReasonSchema() map[string]any {
	return map[string]any{"enum": []string{
		string(CoverageNotObserved), string(CoveragePermissionDenied), string(CoverageThrottled),
		string(CoverageSourceError), string(CoverageUnsupported), string(CoverageCapacityExceeded),
	}}
}

func ownershipSetSchema() map[string]any {
	return objectSchema([]string{"collection", "resource_kind", "resource_id"}, map[string]any{
		"collection": map[string]any{"enum": []string{
			string(CollectionEC2Core), string(CollectionEC2SecurityGroupRules),
			string(CollectionLambdaFunction), string(CollectionLambdaEventSourceMappings),
			string(CollectionEKSCluster), string(CollectionEKSIAMEnrichment),
			string(CollectionS3Bucket), string(CollectionS3BucketEncryption),
		}},
		"resource_kind": resourceKindSchema(),
		"resource_id":   nonEmptyString(),
	})
}

func coverageContractSchema() map[string]any {
	return objectSchema([]string{"scope", "root", "max_resources"}, map[string]any{
		"scope":         ref("#/$defs/coverage_scope"),
		"root":          coverageRootSchema(),
		"max_resources": map[string]any{"type": "integer", "minimum": 1, "maximum": maxCoverageResources},
	})
}
