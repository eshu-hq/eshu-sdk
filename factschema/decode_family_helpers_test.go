// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package factschema

import (
	"reflect"
	"testing"
)

func requiredFieldsForKind(t *testing.T, factKind string) []string {
	t.Helper()
	for _, contract := range payloadContracts {
		if contract.factKind == factKind {
			return payloadKeySetOf(contract.typ).Required
		}
	}
	t.Fatalf("requiredFieldsForKind: no payloadContracts row for fact kind %q", factKind)
	return nil
}

// structTypeForKind returns the payload struct reflect.Type for one fact kind
// from the payloadContracts registry, so a required-field value can be built to
// match the field's real Go type rather than assuming every required field is a
// string.
func structTypeForKind(t *testing.T, factKind string) reflect.Type {
	t.Helper()
	for _, contract := range payloadContracts {
		if contract.factKind == factKind {
			return contract.typ
		}
	}
	t.Fatalf("structTypeForKind: no payloadContracts row for fact kind %q", factKind)
	return nil
}

// requiredFieldValue returns a payload value for a required field that decodes
// cleanly into its Go type. nonEmpty selects a populated value (for the
// full-payload positive test) versus a type-appropriate empty-but-present value
// (for the present-but-empty test): "" or [] or {} rather than always "".
// Deriving the value from the field's kind keeps the generic tests correct for
// any required non-string field (for example gcp_iam_policy_observation.members,
// a required []map[string]string), not just string identity fields.
func requiredFieldValue(t *testing.T, typ reflect.Type, jsonName string, nonEmpty bool) any {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name, _, skip := parseJSONTag(field.Tag.Get("json"), field.Name)
		if skip || name != jsonName {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Slice:
			if nonEmpty {
				// One element whose own shape is valid: a []map[string]string
				// (members) needs a map element; any other slice gets a string
				// element. Either way this is a present, non-empty collection.
				if field.Type.Elem().Kind() == reflect.Map {
					return []any{map[string]any{"k": "v"}}
				}
				return []any{"x"}
			}
			return []any{} // present-but-empty collection
		case reflect.Map:
			if nonEmpty {
				return map[string]any{"k": "v"}
			}
			return map[string]any{}
		case reflect.Struct:
			return requiredPayloadForType(t, field.Type, nonEmpty)
		case reflect.Bool:
			// A required bool has no "empty" value distinct from its zero value
			// (false is both a valid observed value and the Go zero value), so
			// both the nonEmpty and present-but-empty cases return a concrete
			// bool rather than a string: assignField rejects a string payload
			// value for a bool field.
			if nonEmpty {
				return true
			}
			return false
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if nonEmpty {
				return 1
			}
			return 0
		case reflect.Float32, reflect.Float64:
			if nonEmpty {
				return 1.0
			}
			return 0.0
		default:
			if nonEmpty {
				return "x"
			}
			return "" // present-but-empty scalar (a valid observed value)
		}
	}
	t.Fatalf("requiredFieldValue: fact struct %s has no required field with json name %q", typ.Name(), jsonName)
	return nil
}

func requiredPayloadForType(t *testing.T, typ reflect.Type, nonEmpty bool) map[string]any {
	t.Helper()
	out := map[string]any{}
	for _, key := range payloadKeySetOf(typ).Required {
		out[key] = requiredFieldValue(t, typ, key, nonEmpty)
	}
	return out
}

// fullPayloadForKind returns a minimal valid payload map (every required key
// present, non-empty) for one fact kind, so a per-kind test can delete a single
// required key and prove decode dead-letters on exactly that field. Each value
// matches the required field's Go type (a required slice gets a non-empty
// slice, not the string "x").
func fullPayloadForKind(t *testing.T, factKind string) map[string]any {
	t.Helper()
	typ := structTypeForKind(t, factKind)
	return requiredPayloadForType(t, typ, true)
}

// decodeByKind dispatches to the kind's public Decode function so the test
// exercises the real production seam, not a re-implementation. It returns the
// error only, which is all the required-field tests assert on.
