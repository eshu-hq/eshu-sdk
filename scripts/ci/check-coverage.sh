#!/usr/bin/env bash
# Fail when a Go coverprofile total drops below the module floor.
#
# Usage: check-coverage.sh <coverprofile> <floor-percent>
# Example: check-coverage.sh collector/coverage.out 85.0
set -euo pipefail

if [ "$#" -ne 2 ]; then
    printf 'usage: check-coverage.sh <coverprofile> <floor-percent>\n' >&2
    exit 2
fi
profile=$1
floor=$2

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
