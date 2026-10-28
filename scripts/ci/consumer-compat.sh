#!/usr/bin/env bash
# Run the pinned k8s-live satellite suite against a candidate SDK checkout
# and judge the issue #26 outcome (consumer-compat lane).
#
# Usage: consumer-compat.sh <satellite-dir> <sdk-dir>
#
# Both SDK modules in <sdk-dir> are wired into the satellite via temporary
# `go mod edit -replace` directives, so <satellite-dir> must be a throwaway
# checkout (the lane clones one per run). <sdk-dir> is left untouched.
#
# Pass criteria, in order:
#   1. Satellite suite fully green with no StillBlocked pin result: fixed
#      SDK against an updated satellite.
#   2. Suite green except TestExtractedPodTemplateFullHarnessStillBlocked
#      failing with "the SDK allowlist has landed": fixed SDK against a
#      satellite whose block pin still records the pre-fix expectation.
# Anything else fails. In particular a green suite that INCLUDES a passing
# StillBlocked pin fails this script: the pin passing means the
# containers.env_from_secret finding is back, which is the #26 failure
# shape reproduced against a bad SDK.
set -euo pipefail

if [ "$#" -ne 2 ]; then
    printf 'usage: consumer-compat.sh <satellite-dir> <sdk-dir>\n' >&2
    exit 2
fi
satellite=$1
sdk=$2
pin_test='TestExtractedPodTemplateFullHarnessStillBlocked'
landed_msg='the SDK allowlist has landed'

cd "$satellite"
GOWORK=off go mod edit \
    -replace github.com/eshu-hq/eshu-sdk/collector="$sdk/collector" \
    -replace github.com/eshu-hq/eshu-sdk/factschema="$sdk/factschema"

log=$(mktemp /tmp/sdk-compat-XXXXXX.log)
trap 'rm -f "$log"' EXIT

status=0
GOWORK=off go test ./... -count=1 -v >"$log" 2>&1 || status=$?

pin_passed=$(grep -c "^--- PASS: $pin_test" "$log" || true)
pin_failed=$(grep -c "^--- FAIL: $pin_test" "$log" || true)
failures=$(grep -c '^--- FAIL: ' "$log" || true)
landed=$(grep -c "$landed_msg" "$log" || true)

if [ "$status" -eq 0 ] && [ "$pin_passed" -gt 0 ]; then
    printf 'consumer-compat: FAIL - suite green WITH a passing %s pin: the env_from_secret finding is back (#26 shape)\n' "$pin_test" >&2
    exit 1
fi
if [ "$status" -eq 0 ]; then
    printf 'consumer-compat: satellite suite green, no block pin outstanding\n'
    exit 0
fi
if [ "$failures" -eq 1 ] && [ "$pin_failed" -eq 1 ] && [ "$landed" -gt 0 ]; then
    printf 'consumer-compat: suite green except the stale %s pin, which reports the allowlist landed (fixed SDK)\n' "$pin_test"
    exit 0
fi
printf 'consumer-compat: FAIL - satellite suite exited %s with %s failing test(s):\n' "$status" "$failures" >&2
grep '^--- FAIL: ' "$log" >&2 || true
tail -n 20 "$log" >&2
exit 1
