// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Unknown native values need the same key walk as their wire representation.
// MarshalJSON may produce object keys that are invisible in the Go value.
var errCanonicalPayloadRequired = errors.New("canonical JSON payload required")

// ValidateFactPayload checks JSON serialization, duplicate member names, and
// fact-scoped sensitive-key rules for one payload. Kind and SchemaVersion select
// narrow exceptions such as legacy AWS redaction shapes. It does not validate
// payload presence, fact identity, declarations, source references, or a host
// claim; callers must keep native values and custom marshalers stable through
// validation and later use.
func ValidateFactPayload(fact Fact) error {
	return validateFactPayload(fact)
}

func validateFactPayload(fact Fact) error {
	raw, err := json.Marshal(fact.Payload)
	if err != nil {
		// A producer's MarshalJSON error may contain a raw credential. Preserve
		// the failure stage without carrying producer-controlled error text.
		return errors.New("must be JSON serializable")
	}
	if err := walkFactPayload(fact.Payload, "", fact.Kind, fact.SchemaVersion, 0); err == nil {
		return nil
	} else if !errors.Is(err, errCanonicalPayloadRequired) {
		return err
	}
	// Custom marshalers and unknown native types may emit hidden object keys.
	// Check their serialized representation before accepting the result.
	canonical, err := decodeCanonicalPayload(raw)
	if err != nil {
		return err
	}
	return walkFactPayload(canonical, "", fact.Kind, fact.SchemaVersion, 0)
}

func walkFactPayload(value any, path, kind, schema string, depth int) error {
	switch typed := value.(type) {
	case map[string]any:
		awsType := ""
		if depth == 0 && kind == "aws_resource" && schema == "1.0.0" {
			awsType = awsRedactionResourceType(typed)
		}
		for key, child := range typed {
			if err := validateNativeMemberName(key); err != nil {
				return err
			}
			childPath := payloadKeyPath(path, key)
			if depth == 0 && key == "attributes" && awsType != "" {
				if err := validateAWSResourceAttributes(child, childPath, kind, schema, depth+1, awsType); err != nil {
					return err
				}
				continue
			}
			if depth == 0 && key == "tag_value_fingerprints" && kind == "azure_tag_observation" && schema == "1.0.0" {
				if err := validateAzureTagFingerprints(child); err != nil {
					return fmt.Errorf("%s: %w", childPath, err)
				}
				continue
			}
			if depth == 0 {
				if handled, err := validateKubernetesPostureField(child, key, kind, schema); handled {
					if err != nil {
						return fmt.Errorf("%s: %w", childPath, err)
					}
					continue
				}
			}
			if err := rejectSensitivePayloadKey(key, child, childPath); err != nil {
				return err
			}
			if err := walkFactPayload(child, childPath, kind, schema, depth+1); err != nil {
				return err
			}
		}
	case map[string]string:
		for key, val := range typed {
			if err := validateNativeMemberName(key); err != nil {
				return err
			}
			if err := rejectSensitivePayloadKey(key, val, payloadKeyPath(path, key)); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range typed {
			if err := walkFactPayload(child, path, kind, schema, depth+1); err != nil {
				return err
			}
		}
	case []string, nil, string, bool, time.Time, json.Number,
		float64, float32, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		// These values cannot conceal JSON object keys after Marshal succeeds.
	default:
		return errCanonicalPayloadRequired
	}
	return nil
}

func validateAzureTagFingerprints(value any) error {
	switch fingerprints := value.(type) {
	case map[string]string:
		if fingerprints == nil {
			return errors.New("must be an object")
		}
		for key, fingerprint := range fingerprints {
			if err := validateNativeMemberName(key); err != nil {
				return err
			}
			if !validAzureTagMarker(fingerprint) {
				return fmt.Errorf("invalid marker for tag %q", key)
			}
		}
	case map[string]any:
		if fingerprints == nil {
			return errors.New("must be an object")
		}
		for key, raw := range fingerprints {
			if err := validateNativeMemberName(key); err != nil {
				return err
			}
			fingerprint, ok := raw.(string)
			if !ok {
				return errCanonicalPayloadRequired
			}
			if !validAzureTagMarker(fingerprint) {
				return fmt.Errorf("invalid marker for tag %q", key)
			}
		}
	default:
		return errCanonicalPayloadRequired
	}
	return nil
}

func validAzureTagMarker(value string) bool {
	const prefix = "redacted:hmac-sha256:"
	if len(value) != len(prefix)+64 || !strings.HasPrefix(value, prefix) {
		return false
	}
	for i := len(prefix); i < len(value); i++ {
		c := value[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// rejectSensitivePayloadKey fails closed on sensitive-looking keys unless the
// value falls inside the key's allowlist carve-out (see redactionSafeValue):
// env_from_secret admits only JSON booleans (or null), so a credential
// string under that name is still rejected here and in validatePayloadKeys.
func rejectSensitivePayloadKey(key string, value any, path string) error {
	if sensitiveQueryPattern.MatchString(key) && !redactionSafeValue(key, value) {
		return fmt.Errorf("sensitive-looking key %q must be redacted before emission", path)
	}
	return nil
}

func payloadKeyPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

// Go's JSON encoder replaces malformed UTF-8 in object member names with
// U+FFFD. Reject those names before the native fast path can treat two
// distinct map keys as distinct serialized members.
func validateNativeMemberName(key string) error {
	if !utf8.ValidString(key) {
		return errInvalidCanonicalPayload
	}
	return nil
}
