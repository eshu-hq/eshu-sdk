#!/usr/bin/env bash
# Fail when a Go coverprofile total drops below the module floor.
#
# Usage: check-coverage.sh <coverprofile> <floor-percent>
# Example: check-coverage.sh collector/coverage.out 85.0
#
# Coverage exclusion verdict (issue #34, epic #27): neither module excludes
# any file from its total, so the floor sees every committed statement.
# Census (2026-10-28, origin/main d62cbec): no committed Go file carries its
# own generated-code marker —
#   rg -n "Code generated|DO NOT EDIT" --glob '*.go' factschema/  # exit 1, no hits
#   rg -n "Code generated|DO NOT EDIT" --glob '*.go' collector/   # one hit, below
# The single collector hit is collector/entrypoints/generate.go:39, a template
# string for downstream provider-owned mains, not a file-owned marker. The
# factschema //go:generate directive (factschema/decode.go:6) emits only
# *.schema.json under factschema/schema/ (see
# factschema/internal/schemagen/cmd/main.go and
# factschema/schema_gen_test.go), which never enters Go coverage. A future
# exclusion needs a reviewed file-by-file list; silent package exclusions
# stay banned.
set -euo pipefail

if [ "$#" -ne 2 ]; then
    printf 'usage: check-coverage.sh <coverprofile> <floor-percent>\n' >&2
    exit 2
fi
profile=$1
floor=$2

# Fail closed on a non-numeric floor: awk would coerce it to 0 and pass any
# total, so reject anything that is not a plain decimal percentage first.
case "$floor" in
'' | *[!0-9.]* | .* | *.) printf 'check-coverage: bad floor %s\n' "$floor" >&2; exit 2 ;;
esac
case "$floor" in
*.*.*) printf 'check-coverage: bad floor %s\n' "$floor" >&2; exit 2 ;;
esac

total=$(go tool cover -func="$profile" | awk '/^total:/ { sub(/%$/, "", $3); print $3 }')
if [ -z "$total" ]; then
    printf 'check-coverage: no total line in %s\n' "$profile" >&2
    exit 1
fi
if awk -v total="$total" -v floor="$floor" 'BEGIN { exit (total + 0 >= floor + 0) ? 0 : 1 }'; then
    printf 'check-coverage: %s%% >= floor %s%% (%s)\n' "$total" "$floor" "$profile"
else
    printf 'check-coverage: %s%% < floor %s%% (%s)\n' "$total" "$floor" "$profile" >&2
    exit 1
fi
