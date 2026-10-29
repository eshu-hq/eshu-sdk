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
	defs := schema["$defs"].(map[string]any)
	defs["generation"].(map[string]any)["properties"].(map[string]any)["trusted_deletion"] = map[string]any{"type": "boolean"}
	fact := defs["fact"].(map[string]any)
	fact["properties"].(map[string]any)["ownership"] = ref("#/$defs/ownership_set")
	required := fact["required"].([]string)
	fact["required"] = append(required, "ownership")
	defs["coverage_scope"] = coverageScopeSchema()
	defs["ownership_set"] = ownershipSetSchema()
	defs["set_coverage"] = objectSchema([]string{"set", "state"}, map[string]any{
		"set":   ref("#/$defs/ownership_set"),
		"state": map[string]any{"enum": []string{string(CoverageComplete), string(CoverageIncomplete)}},
		"reason": map[string]any{"enum": []string{
			string(CoverageNotObserved), string(CoveragePermissionDenied), string(CoverageThrottled),
			string(CoverageSourceError), string(CoverageUnsupported),
		}},
	})
	defs["coverage_report"] = objectSchema([]string{"sets"}, map[string]any{
		"sets": map[string]any{"type": "array", "minItems": 1, "maxItems": maxCoverageSets, "items": ref("#/$defs/set_coverage")},
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

func ownershipSetSchema() map[string]any {
	return objectSchema([]string{"collection"}, map[string]any{
		"collection": map[string]any{"enum": []string{
			string(CollectionAWSService), string(CollectionEC2Core), string(CollectionEC2SecurityGroupRules),
			string(CollectionLambdaFunction), string(CollectionLambdaEventSourceMappings),
			string(CollectionEKSCluster), string(CollectionEKSIAMEnrichment),
			string(CollectionS3Bucket), string(CollectionS3BucketEncryption),
		}},
		"resource_id": map[string]any{"type": "string", "maxLength": 2048},
	})
}

func coverageContractSchema() map[string]any {
	return objectSchema([]string{"scope", "sets"}, map[string]any{
		"scope": ref("#/$defs/coverage_scope"),
		"sets":  map[string]any{"type": "array", "minItems": 1, "maxItems": maxCoverageSets, "items": ref("#/$defs/ownership_set")},
	})
}
