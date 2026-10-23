// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 eshu-hq

package factschema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestPayloadContractsCoverAllSchemas(t *testing.T) {
	t.Parallel()

	contractKinds := make(map[string]bool, len(payloadContracts))
	for _, contract := range payloadContracts {
		if contractKinds[contract.factKind] {
			t.Fatalf("duplicate payloadContracts fact kind %q", contract.factKind)
		}
		contractKinds[contract.factKind] = true
	}
	decodedKinds := make(map[string]bool, len(allDecodedKinds))
	for _, kind := range allDecodedKinds {
		if decodedKinds[kind] {
			t.Fatalf("duplicate allDecodedKinds fact kind %q", kind)
		}
		decodedKinds[kind] = true
		if !contractKinds[kind] {
			t.Fatalf("allDecodedKinds fact kind %q has no payloadContracts row", kind)
		}
	}
	for kind := range contractKinds {
		if !decodedKinds[kind] {
			t.Fatalf("payloadContracts fact kind %q has no allDecodedKinds row", kind)
		}
		decode, ok := decoderByKind[kind]
		if !ok || decode == nil {
			t.Fatalf("payloadContracts fact kind %q has no public decoder", kind)
		}
	}
	for kind, decode := range decoderByKind {
		if decode == nil || !contractKinds[kind] || !decodedKinds[kind] {
			t.Fatalf("public decoder for fact kind %q is nil or has no matching contract and decoded-kind row", kind)
		}
	}

	entries, err := os.ReadDir("schema")
	if err != nil {
		t.Fatalf("read schema dir: %v", err)
	}
	schemaFiles := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schema.json") {
			continue
		}
		schemaFiles[entry.Name()] = true
	}

	registered := map[string]bool{}
	for _, contract := range payloadContracts {
		if registered[contract.schemaFile] {
			t.Fatalf("payloadContracts registers %q more than once", contract.schemaFile)
		}
		registered[contract.schemaFile] = true
		if !schemaFiles[contract.schemaFile] {
			t.Fatalf("payloadContracts row for %q names schema file %q, which does not exist under schema/", contract.factKind, contract.schemaFile)
		}
	}
	for name := range schemaFiles {
		if !registered[name] {
			t.Fatalf("schema file %q has no payloadContracts row; add one so its key set is drift-checked", name)
		}
	}
}

// TestDerivedKeySetsMatchGeneratedSchemas is the definitive single-source-of-
// truth lock: for each registered fact kind it derives the required and known
// key sets from the payload struct (via the same payloadKeySetOf the decode
// path uses) and asserts they equal the generated schema's "required" array
// and "properties" keys. Because the schema is generated from the same struct,
// the two agree automatically unless the two derivation rules diverge — for
// example an invopop upgrade changing its required semantics, or a change to
// parseJSONTag. That divergence, not a hand-map edit, is the only remaining
// drift axis, and this test is what catches it.
func TestDerivedKeySetsMatchGeneratedSchemas(t *testing.T) {
	t.Parallel()

	for _, contract := range payloadContracts {
		t.Run(contract.factKind, func(t *testing.T) {
			t.Parallel()

			raw, err := os.ReadFile(filepath.Join("schema", contract.schemaFile))
			if err != nil {
				t.Fatalf("read schema %q: %v", contract.schemaFile, err)
			}
			var doc struct {
				Properties map[string]json.RawMessage `json:"properties"`
				Required   []string                   `json:"required"`
			}
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("unmarshal schema %q: %v", contract.schemaFile, err)
			}

			ks := payloadKeySetOf(contract.typ)

			schemaProperties := map[string]bool{}
			for name := range doc.Properties {
				schemaProperties[name] = true
			}
			if got := keySet(ks.Known); !reflect.DeepEqual(got, schemaProperties) {
				t.Fatalf("derived known keys = %v, want schema properties %v", sortedKeys(got), sortedKeys(schemaProperties))
			}

			schemaRequired := map[string]bool{}
			for _, name := range doc.Required {
				schemaRequired[name] = true
			}
			if got := keySet(ks.Required); !reflect.DeepEqual(got, schemaRequired) {
				t.Fatalf("derived required keys = %v, want schema required %v", sortedKeys(got), sortedKeys(schemaRequired))
			}
		})
	}
}

// TestPayloadStructShapeConvention enforces the two struct-shape bans that keep
// the required rule ("no omitempty ⇒ required") unambiguous per field. A
// pointer field without omitempty is required by the schema yet nullable, a
// required-but-nullable contradiction decodeAndValidate would reject at runtime
// as a null required field. A non-pointer, non-slice, non-map field with
// omitempty collapses the absent and zero-value states, discarding the
// observed/not-observed distinction the pointer-and-omitempty optional fields
// exist to preserve. Slice and map fields are exempt from that second ban: a
// nil slice/map already round-trips through omitempty exactly like a nil
// pointer (encoding/json omits a nil or empty slice/map with omitempty and a
// json.Unmarshal of an absent key leaves it nil), so a []string field tagged
// `json:"x,omitempty"` is not ambiguous the way a bare string field tagged
// `json:"x,omitempty"` would be. Banning the pointer and scalar shapes means
// the schema generator's "no omitempty ⇒ required" rule and the intuition
// "pointer/slice/map ⇒ optional" can never disagree.
func TestPayloadStructShapeConvention(t *testing.T) {
	t.Parallel()

	for _, contract := range payloadContracts {
		t.Run(contract.factKind, func(t *testing.T) {
			t.Parallel()

			typ := contract.typ
			seen := map[string]string{} // json name -> first Go field that declared it
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				if field.Anonymous {
					t.Fatalf("field %q is embedded; payload structs must be flat", field.Name)
				}
				if field.PkgPath != "" {
					continue
				}
				name, omitEmpty, skip := parseJSONTag(field.Tag.Get("json"), field.Name)
				if skip {
					continue
				}
				if prior, dup := seen[name]; dup {
					t.Fatalf("fields %q and %q both serialize to json key %q; payload key names must be unique", prior, field.Name, name)
				}
				seen[name] = field.Name
				switch field.Type.Kind() {
				case reflect.Pointer:
					if !omitEmpty {
						t.Fatalf("field %q is a pointer without omitempty (required-but-nullable); add omitempty or make it a value type", field.Name)
					}
				case reflect.Slice, reflect.Map:
					// Nil is both the zero value and the absent-key decode
					// result for a slice/map, so omitempty does not collapse
					// a distinction the way it would for a scalar; either
					// tagging is unambiguous. Most slice/map fields carry
					// omitempty by convention (see aws/v1, iam/v1) and are
					// therefore optional. A slice/map WITHOUT omitempty is a
					// required collection: the schema lists it in "required" and
					// the decode seam dead-letters an absent or null key. That is
					// only correct when the emitter unconditionally writes the
					// key, so it must be opted in explicitly via
					// intentionalRequiredCollections; a required collection not on
					// that list is a bug (a field that would dead-letter a valid
					// fact the emitter can produce without it).
					_, allowed := intentionalRequiredCollections[requiredCollectionKey{contract.factKind, name}]
					if !omitEmpty && !allowed {
						t.Fatalf("field %q is a required slice/map (no omitempty) but is not in intentionalRequiredCollections; add it there with a justification or add omitempty", field.Name)
					}
				default:
					if omitEmpty {
						t.Fatalf("field %q is a non-pointer with omitempty (collapses absent and zero value); make it a pointer or drop omitempty", field.Name)
					}
				}
			}
		})
	}
}

// BenchmarkDecodeAWSResource records the baseline cost of the current
// json.Marshal/Unmarshal decode path so any future move to a lower-allocation
// decoder is justified by a before/after measurement rather than intuition.
func BenchmarkDecodeAWSResource(b *testing.B) {
	env := testEnvelope(fullAWSResourcePayload())
	b.ReportAllocs()
	for b.Loop() {
		if _, err := DecodeAWSResource(env); err != nil {
			b.Fatalf("DecodeAWSResource() error = %v", err)
		}
	}
}

// keySet collects a slice of key names into a set so the drift test can
// compare the derived keys against the generated schema's keys order-
// independently. Duplicate json names are caught separately, by
// TestPayloadStructShapeConvention, not here.
func keySet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
