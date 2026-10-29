// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/eshu-hq/eshu-sdk/collector/marker"
)

func awsRedactionResult(t *testing.T, resourceType, service string, attributes map[string]any) (Result, Validator) {
	t.Helper()
	result := readFixture(t, "complete")
	result.Facts[0].Kind = "aws_resource"
	result.Facts[0].SchemaVersion = "1.0.0"
	result.Facts[0].Payload = map[string]any{
		"resource_type": resourceType,
		"service_kind":  service,
		"attributes":    attributes,
	}
	contract := Contract{ProtocolVersion: ProtocolVersionV1Alpha1, Facts: []FactDeclaration{{
		Kind: "aws_resource", SchemaVersions: []string{"1.0.0", "2.0.0"}, SourceConfidence: []SourceConfidence{SourceConfidenceReported},
	}}}
	return result, NewValidator(contract)
}

func awsMarker(t *testing.T, source string) map[string]any {
	t.Helper()
	key, err := marker.NewKey([]byte("synthetic-aws-redaction-test-key"))
	if err != nil {
		t.Fatal(err)
	}
	value := marker.String("synthetic-secret", "known_sensitive_key", source, key)
	return map[string]any{"marker": value.Marker, "reason": value.Reason, "source": value.Source, "ruleset_version": "aws-launch-2026-05-14"}
}

func TestAWSLegacyRedactedEnvironmentAdmission(t *testing.T) {
	lambda, validator := awsRedactionResult(t, "aws_lambda_function", "lambda", map[string]any{
		"environment": map[string]any{"API_KEY": awsMarker(t, "lambda.function.environment.API_KEY")},
	})
	ecs, ecsValidator := awsRedactionResult(t, "aws_ecs_task_definition", "ecs", map[string]any{
		"containers": []any{map[string]any{
			"environment": []any{map[string]any{"name": "API_KEY", "value": awsMarker(t, "ecs.task_definition.container.environment.API_KEY")}},
			"secrets":     []any{map[string]any{"name": "SECRET", "value_from": "arn:aws:secretsmanager:us-east-1:123456789012:secret:synthetic"}},
		}},
	})
	for _, tc := range []struct {
		name      string
		result    Result
		validator Validator
	}{
		{"lambda", lambda, validator},
		{"ecs", ecs, ecsValidator},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, candidate := range []Result{tc.result, wireResult(t, tc.result)} {
				if _, err := tc.validator.ValidateResult(candidate); err != nil {
					t.Fatalf("legacy redacted payload rejected: %v", err)
				}
			}
		})
	}
	if err := ValidateShareSafeKeys(lambda.Facts[0].Payload); err == nil {
		t.Fatal("context-free sensitive-key policy was relaxed")
	}
}

func TestAWSLegacyRedactedEnvironmentRejectsUnsafeShapes(t *testing.T) {
	base, validator := awsRedactionResult(t, "aws_lambda_function", "lambda", map[string]any{
		"environment": map[string]any{"API_KEY": awsMarker(t, "lambda.function.environment.API_KEY")},
	})
	for _, tc := range []struct {
		name   string
		mutate func(*Result)
	}{
		{"raw value", func(r *Result) {
			r.Facts[0].Payload["attributes"].(map[string]any)["environment"] = map[string]any{"API_KEY": "raw-secret"}
		}},
		{"wrong source", func(r *Result) {
			r.Facts[0].Payload["attributes"].(map[string]any)["environment"].(map[string]any)["API_KEY"].(map[string]any)["source"] = "lambda.function.environment.OTHER"
		}},
		{"wrong reason", func(r *Result) {
			r.Facts[0].Payload["attributes"].(map[string]any)["environment"].(map[string]any)["API_KEY"].(map[string]any)["reason"] = "known_provider_schema"
		}},
		{"wrong ruleset", func(r *Result) {
			r.Facts[0].Payload["attributes"].(map[string]any)["environment"].(map[string]any)["API_KEY"].(map[string]any)["ruleset_version"] = "later"
		}},
		{"extra marker member", func(r *Result) {
			r.Facts[0].Payload["attributes"].(map[string]any)["environment"].(map[string]any)["API_KEY"].(map[string]any)["raw"] = "raw-secret"
		}},
		{"uppercase digest", func(r *Result) {
			r.Facts[0].Payload["attributes"].(map[string]any)["environment"].(map[string]any)["API_KEY"].(map[string]any)["marker"] = marker.Prefix + strings.Repeat("A", 64)
		}},
		{"wrong service", func(r *Result) { r.Facts[0].Payload["service_kind"] = "ecs" }},
		{"wrong resource", func(r *Result) { r.Facts[0].Payload["resource_type"] = "aws_ec2_instance" }},
		{"wrong schema", func(r *Result) { r.Facts[0].SchemaVersion = "2.0.0" }},
		{"dotted impersonation", func(r *Result) {
			r.Facts[0].Payload["attributes.environment"] = r.Facts[0].Payload["attributes"].(map[string]any)["environment"]
			delete(r.Facts[0].Payload, "attributes")
		}},
		{"nested sensitive key", func(r *Result) {
			r.Facts[0].Payload["attributes"].(map[string]any)["nested"] = map[string]any{"password": "raw-secret"}
		}},
		{"duplicate JSON member", func(r *Result) {
			r.Facts[0].Payload["attributes"] = json.RawMessage(`{"environment":{"API_KEY":{"marker":"x","marker":"y"}}}`)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var result Result
			raw, err := json.Marshal(base)
			if err != nil || json.Unmarshal(raw, &result) != nil {
				t.Fatalf("clone result: %v", err)
			}
			tc.mutate(&result)
			if _, err := validator.ValidateResult(result); err == nil {
				t.Fatal("unsafe payload accepted")
			}
		})
	}
}

func TestAWSECSEnvironmentAndSecretReferenceBoundaries(t *testing.T) {
	valid := []struct {
		name       string
		containers any
	}{
		{"absent", []any{map[string]any{"name": "app"}}},
		{"null", []any{map[string]any{"environment": nil, "secrets": nil}}},
		{"empty", []any{map[string]any{"environment": []any{}, "secrets": []any{}}}},
		{"no containers", nil},
	}
	for _, tc := range valid {
		t.Run(tc.name, func(t *testing.T) {
			result, validator := awsRedactionResult(t, "aws_ecs_task_definition", "ecs", map[string]any{"containers": tc.containers})
			for _, candidate := range []Result{result, wireResult(t, result)} {
				if _, err := validator.ValidateResult(candidate); err != nil {
					t.Fatalf("legacy empty/null state rejected: %v", err)
				}
			}
		})
	}
	bad := []struct {
		name      string
		container map[string]any
	}{
		{"raw env", map[string]any{"environment": []any{map[string]any{"name": "PASSWORD", "value": "raw-secret"}}}},
		{"extra env field", map[string]any{"environment": []any{map[string]any{"name": "PASSWORD", "value": awsMarker(t, "ecs.task_definition.container.environment.PASSWORD"), "raw": "raw-secret"}}}},
		{"wrong env source", map[string]any{"environment": []any{map[string]any{"name": "PASSWORD", "value": awsMarker(t, "ecs.task_definition.container.environment.OTHER")}}}},
		{"secret extra", map[string]any{"secrets": []any{map[string]any{"name": "TOKEN", "value_from": "arn:aws:secretsmanager:us-east-1:123456789012:secret:synthetic", "raw": "raw-secret"}}}},
		{"secret nested", map[string]any{"secrets": []any{map[string]any{"name": "TOKEN", "value_from": map[string]any{"password": "raw-secret"}}}}},
		{"secret scalar", map[string]any{"secrets": "raw-secret"}},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			result, validator := awsRedactionResult(t, "aws_ecs_task_definition", "ecs", map[string]any{"containers": []any{tc.container}})
			for _, candidate := range []Result{result, wireResult(t, result)} {
				if _, err := validator.ValidateResult(candidate); err == nil {
					t.Fatal("unsafe ECS payload accepted")
				}
			}
		})
	}
}
