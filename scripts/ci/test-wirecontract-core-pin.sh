#!/usr/bin/env bash
# Core-pin drift guard for the wirecontract module (eshu-hq/eshu#9509).
#
# The byte-parity claim needs an oracle that lives in this repo's gates:
# goldens committed next to the port prove only self-consistency. This guard
# rebuilds the same bytes from the CORE live types at the pinned commit and
# from the PORTED live types, and requires golden == core == port for every
# marshal golden. A mistranscribed tag, a dropped field, or golden rot on
# either side fails here. Core-side truth stays covered by core's own suite.
#
# Pin: core sdk/go/apicontract at c606960c97, served by the Go module proxy
# as v0.0.0-20261029223121-c606960c9771. Bump the pin only with a re-port.
# Network to proxy.golang.org is required; the guard fails closed without it.
#
# Run: bash scripts/ci/test-wirecontract-core-pin.sh (also via `make shape`).

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PIN="v0.0.0-20261029223121-c606960c9771"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

cat > "$WORK/go.mod" <<EOF
module wirecontractpincheck

go 1.26.5

require (
	github.com/eshu-hq/eshu-sdk/wirecontract v0.0.0
	github.com/eshu-hq/eshu/sdk/go/apicontract $PIN
)

replace github.com/eshu-hq/eshu-sdk/wirecontract => $ROOT/wirecontract
EOF

cat > "$WORK/main.go" <<'EOF'
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	core "github.com/eshu-hq/eshu/sdk/go/apicontract"
	port "github.com/eshu-hq/eshu-sdk/wirecontract"
)

func remarshal(v any) []byte {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(raw, '\n')
}

var failures int

func check(name string, golden []byte, coreV, portV func([]byte) []byte) {
	want := golden
	if got := coreV(golden); !bytes.Equal(got, want) {
		failures++
		fmt.Printf("FAIL: core remarshal of %s differs from golden\n", name)
	}
	if got := portV(golden); !bytes.Equal(got, want) {
		failures++
		fmt.Printf("FAIL: port remarshal of %s differs from golden\n", name)
	}
}

func main() {
	goldenDir := filepath.Join(os.Args[1], "wirecontract", "testdata", "golden")
	read := func(name string) []byte {
		raw, err := os.ReadFile(filepath.Join(goldenDir, name))
		if err != nil {
			panic(err)
		}
		return raw
	}
	check("envelope_truth.json", read("envelope_truth.json"),
		func(raw []byte) []byte {
			var v core.ResponseEnvelope
			must(json.Unmarshal(raw, &v))
			return remarshal(v)
		},
		func(raw []byte) []byte {
			var v port.ResponseEnvelope
			must(json.Unmarshal(raw, &v))
			return remarshal(v)
		})
	check("envelope_error.json", read("envelope_error.json"),
		func(raw []byte) []byte {
			var v core.ResponseEnvelope
			must(json.Unmarshal(raw, &v))
			return remarshal(v)
		},
		func(raw []byte) []byte {
			var v port.ResponseEnvelope
			must(json.Unmarshal(raw, &v))
			return remarshal(v)
		})
	check("tool_definition.json", read("tool_definition.json"),
		func(raw []byte) []byte {
			var v core.ToolDefinition
			must(json.Unmarshal(raw, &v))
			return remarshal(v)
		},
		func(raw []byte) []byte {
			var v port.ToolDefinition
			must(json.Unmarshal(raw, &v))
			return remarshal(v)
		})
	check("env_client_entries.json", read("env_client_entries.json"),
		func(raw []byte) []byte {
			var v []core.Entry
			must(json.Unmarshal(raw, &v))
			return remarshal(v)
		},
		func(raw []byte) []byte {
			var v []port.Entry
			must(json.Unmarshal(raw, &v))
			return remarshal(v)
		})
	for _, name := range []string{
		"validate_client_valid_client_env.json",
		"validate_client_invalid_and_typo.json",
		"validate_client_strict_unknown.json",
	} {
		check(name, read(name),
			func(raw []byte) []byte {
				var v []core.Finding
				must(json.Unmarshal(raw, &v))
				return remarshal(v)
			},
			func(raw []byte) []byte {
				var v []port.Finding
				must(json.Unmarshal(raw, &v))
				return remarshal(v)
			})
	}
	// Guard B: fully-populated shape key sets. Guard A (above) is blind to
	// an additive omitempty field unset in every golden: both remarshals
	// still match. Populating every field forces every json key out, so a
	// field added on either side changes that side's key set and fails.
	shapePairs := []struct {
		name string
		core any
		port any
	}{
		{"ResponseEnvelope", core.ResponseEnvelope{}, port.ResponseEnvelope{}},
		{"ToolDefinition", core.ToolDefinition{}, port.ToolDefinition{}},
		{"Entry", core.Entry{}, port.Entry{}},
		{"Finding", core.Finding{}, port.Finding{}},
	}
	for _, p := range shapePairs {
		coreKeys := populatedKeys(p.core)
		portKeys := populatedKeys(p.port)
		if !reflect.DeepEqual(coreKeys, portKeys) {
			failures++
			fmt.Printf("FAIL: shape keys differ for %s:\n core: %v\n port: %v\n", p.name, coreKeys, portKeys)
		}
	}
	if failures != 0 {
		fmt.Printf("core-pin: %d mismatches (see above)\n", failures)
		os.Exit(1)
	}
	fmt.Println("core-pin: golden == core == port on all 7 goldens; shape keys match on all 4 types")
}

// populatedKeys marshals a fully-populated value and returns its sorted JSON
// key paths. Every settable field gets a non-zero dummy (recursed into
// structs, slices, maps, and pointers), so omitempty drops nothing and an
// added field always changes the set.
func populatedKeys(v any) []string {
	rv := reflect.New(reflect.TypeOf(v)).Elem()
	populate(rv)
	raw, err := json.Marshal(rv.Interface())
	must(err)
	var decoded any
	must(json.Unmarshal(raw, &decoded))
	var keys []string
	collectKeys(decoded, "", &keys)
	sort.Strings(keys)
	return keys
}

func populate(rv reflect.Value) {
	switch rv.Kind() {
	case reflect.String:
		rv.SetString("x")
	case reflect.Bool:
		rv.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		rv.SetUint(1)
	case reflect.Float32, reflect.Float64:
		rv.SetFloat(1.5)
	case reflect.Slice:
		rv.Set(reflect.MakeSlice(rv.Type(), 1, 1))
		populate(rv.Index(0))
	case reflect.Map:
		rv.Set(reflect.MakeMapWithSize(rv.Type(), 1))
		k := reflect.New(rv.Type().Key()).Elem()
		populateMapKey(k)
		e := reflect.New(rv.Type().Elem()).Elem()
		populateInterface(e)
		rv.SetMapIndex(k, e)
	case reflect.Ptr:
		rv.Set(reflect.New(rv.Type().Elem()))
		populate(rv.Elem())
	case reflect.Interface:
		populateInterface(rv)
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			if rv.Type().Field(i).PkgPath != "" {
				continue
			}
			populate(rv.Field(i))
		}
	}
}

func populateMapKey(rv reflect.Value) {
	if rv.Kind() == reflect.String {
		rv.SetString("k")
		return
	}
	populate(rv)
}

func populateInterface(rv reflect.Value) {
	if rv.Kind() != reflect.Interface {
		populate(rv)
		return
	}
	m := map[string]any{"k": "x"}
	rv.Set(reflect.ValueOf(m))
}

func collectKeys(v any, prefix string, out *[]string) {
	switch t := v.(type) {
	case map[string]any:
		for k, e := range t {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			*out = append(*out, p)
			collectKeys(e, p, out)
		}
	case []any:
		for _, e := range t {
			collectKeys(e, prefix, out)
		}
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
EOF

cd "$WORK"
if ! GOWORK=off GOFLAGS=-mod=mod go mod tidy >/dev/null 2>&1; then
	echo "core-pin: cannot resolve the pinned core module (proxy.golang.org unreachable?)" >&2
	exit 1
fi
GOWORK=off go run . "$ROOT"
