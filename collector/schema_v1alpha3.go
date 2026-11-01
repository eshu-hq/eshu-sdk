// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import "encoding/json"

const (
	schemaIDV1Alpha3        = "https://eshu.dev/schemas/collector-sdk/v1alpha3/result.schema.json"
	requestSchemaIDV1Alpha3 = "https://eshu.dev/schemas/collector-sdk/v1alpha3/request.schema.json"
)

// JSONSchemaV1Alpha3 returns the collector-sdk/v1alpha3 result JSON Schema:
// the v1alpha1 result shape plus the optional paged-result continuation in
// Result.Page. The v1alpha1 artifact returned by JSONSchema is unchanged.
func JSONSchemaV1Alpha3() ([]byte, error) {
	schema := resultSchema()
	schema["$id"] = schemaIDV1Alpha3
	properties := schema["properties"].(map[string]any)
	properties["protocol_version"] = map[string]any{"const": ProtocolVersionV1Alpha3}
	properties["page"] = ref("#/$defs/page")
	schema["$defs"].(map[string]any)["page"] = pageSchema()
	raw, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

// RequestJSONSchemaV1Alpha3 returns the collector-sdk/v1alpha3 request JSON
// Schema: the request shape pinned to the v1alpha3 protocol on both the
// request and its contract, plus the optional Request.Continuation resume
// cursor. The artifact returned by RequestJSONSchema is unchanged.
func RequestJSONSchemaV1Alpha3() ([]byte, error) {
	schema := requestSchema()
	schema["$id"] = requestSchemaIDV1Alpha3
	delete(schema, "oneOf")
	properties := schema["properties"].(map[string]any)
	properties["protocol_version"] = map[string]any{"const": ProtocolVersionV1Alpha3}
	properties["continuation"] = ref("#/$defs/continuation")
	defs := schema["$defs"].(map[string]any)
	contract := defs["contract"].(map[string]any)
	contract["properties"].(map[string]any)["protocol_version"] = map[string]any{"const": ProtocolVersionV1Alpha3}
	defs["continuation"] = objectSchema(
		[]string{"cursor"},
		map[string]any{"cursor": nonEmptyString()},
	)
	raw, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

// pageSchema describes the optional v1alpha3 paged-result continuation.
// Cross-field rules (cursor required while more is true, empty on the final
// page, continuation only on complete or partial) are enforced by
// Validator.ValidateResult and the reassembling host.
func pageSchema() map[string]any {
	return objectSchema(
		[]string{},
		map[string]any{
			"cursor":      map[string]any{"type": "string"},
			"more":        map[string]any{"type": "boolean"},
			"page_index":  map[string]any{"type": "integer", "minimum": 0},
			"total_facts": map[string]any{"type": "integer", "minimum": 0},
		},
	)
}
