// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"errors"
	"fmt"
	"strings"
)

const awsRedactionRuleset = "aws-launch-2026-05-14"

// awsRedactionResourceType selects only the two legacy service-specific payloads.
// An inconsistent service_kind disables the exception, so the ordinary
// sensitive-key walk rejects the embedded environment or secret keys.
func awsRedactionResourceType(payload map[string]any) string {
	resourceType, _ := payload["resource_type"].(string)
	service, present := payload["service_kind"]
	switch resourceType {
	case "aws_lambda_function":
		if !present || service == "lambda" {
			return resourceType
		}
	case "aws_ecs_task_definition":
		if !present || service == "ecs" {
			return resourceType
		}
	}
	return ""
}

func validateAWSResourceAttributes(value any, path, kind, schema string, depth int, resourceType string) error {
	attributes, ok := value.(map[string]any)
	if !ok {
		return errCanonicalPayloadRequired
	}
	for key, child := range attributes {
		childPath := payloadKeyPath(path, key)
		switch {
		case resourceType == "aws_lambda_function" && key == "environment":
			if err := validateAWSLambdaEnvironment(child); err != nil {
				return fmt.Errorf("%s: %w", childPath, err)
			}
		case resourceType == "aws_ecs_task_definition" && key == "containers":
			if err := validateAWSContainers(child, childPath, kind, schema, depth+1); err != nil {
				return err
			}
		default:
			if err := rejectSensitivePayloadKey(key, child, childPath); err != nil {
				return err
			}
			if err := walkFactPayload(child, childPath, kind, schema, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateAWSLambdaEnvironment(value any) error {
	if value == nil { // Legacy scanner emits null when no variables were observed.
		return nil
	}
	environment, ok := value.(map[string]any)
	if !ok {
		return errCanonicalPayloadRequired
	}
	for name, marker := range environment {
		if name == "" || name != strings.TrimSpace(name) {
			return errors.New("invalid environment name")
		}
		if err := validateAWSMarker(marker, "lambda.function.environment."+name); err != nil {
			return fmt.Errorf("invalid environment marker: %w", err)
		}
	}
	return nil
}

func validateAWSContainers(value any, path, kind, schema string, depth int) error {
	if value == nil { // Legacy scanner emits null for no containers.
		return nil
	}
	containers, ok := value.([]any)
	if !ok {
		return errCanonicalPayloadRequired
	}
	for _, rawContainer := range containers {
		container, ok := rawContainer.(map[string]any)
		if !ok {
			return errCanonicalPayloadRequired
		}
		for key, child := range container {
			childPath := payloadKeyPath(path, key)
			switch key {
			case "environment":
				if err := validateAWSEnvironmentEntries(child); err != nil {
					return fmt.Errorf("%s: %w", childPath, err)
				}
			case "secrets":
				if err := validateAWSSecretReferences(child); err != nil {
					return fmt.Errorf("%s: %w", childPath, err)
				}
			default:
				if err := rejectSensitivePayloadKey(key, child, childPath); err != nil {
					return err
				}
				if err := walkFactPayload(child, childPath, kind, schema, depth+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateAWSEnvironmentEntries(value any) error {
	if value == nil { // Legacy scanner emits null for no environment variables.
		return nil
	}
	entries, ok := value.([]any)
	if !ok {
		return errCanonicalPayloadRequired
	}
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			return errCanonicalPayloadRequired
		}
		if len(entry) != 2 {
			return errors.New("environment entry must contain name and value")
		}
		name, ok := entry["name"].(string)
		if !ok || name != strings.TrimSpace(name) {
			return errors.New("invalid environment name")
		}
		if err := validateAWSMarker(entry["value"], "ecs.task_definition.container.environment."+name); err != nil {
			return fmt.Errorf("invalid environment marker: %w", err)
		}
	}
	return nil
}

func validateAWSSecretReferences(value any) error {
	if value == nil { // Legacy scanner emits null for no secret references.
		return nil
	}
	entries, ok := value.([]any)
	if !ok {
		return errCanonicalPayloadRequired
	}
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			return errCanonicalPayloadRequired
		}
		if len(entry) != 2 {
			return errors.New("secret reference must contain name and value_from")
		}
		name, nameOK := entry["name"].(string)
		from, fromOK := entry["value_from"].(string)
		if !nameOK || !fromOK || name != strings.TrimSpace(name) || from != strings.TrimSpace(from) {
			return errors.New("invalid secret reference")
		}
	}
	return nil
}

func validateAWSMarker(value any, source string) error {
	fields, ok := value.(map[string]any)
	if !ok {
		return errCanonicalPayloadRequired
	}
	if len(fields) != 4 {
		return errors.New("must contain exactly four marker fields")
	}
	marker, markerOK := fields["marker"].(string)
	reason, reasonOK := fields["reason"].(string)
	actualSource, sourceOK := fields["source"].(string)
	ruleset, rulesetOK := fields["ruleset_version"].(string)
	if !markerOK || !validAzureTagMarker(marker) || !reasonOK ||
		(reason != "known_sensitive_key" && reason != "unknown_provider_schema") ||
		!sourceOK || actualSource != source || !rulesetOK || ruleset != awsRedactionRuleset {
		return errors.New("invalid AWS redaction marker")
	}
	return nil
}
