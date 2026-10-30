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
	if failures != 0 {
		fmt.Printf("core-pin: %d mismatches (see above)\n", failures)
		os.Exit(1)
	}
	fmt.Println("core-pin: golden == core == port on all 4 marshal goldens")
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
