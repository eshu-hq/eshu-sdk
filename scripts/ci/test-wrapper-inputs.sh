#!/usr/bin/env bash
# Wrapper-input pin for the org reusable (issue #39, epic #27).
#
# The wrapper passes a fixed set of inputs to
# eshu-hq/.github/.github/workflows/go-multi-module-ci.yml. This test fails
# when the set drifts (a new input sneaks in, a required one goes missing),
# when the reusable or satellite pins are not full SHAs (the .github#3 dead-
# pin class), or when the lane inputs stop parsing, so the contract the PR
# body claims is pinned in-repo and runs in CI via the shape job.
#
# Run: bash scripts/ci/test-wrapper-inputs.sh [workflow] [Makefile]

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORKFLOW="${1:-$ROOT/.github/workflows/ci.yml}"
MAKEFILE="${2:-$ROOT/Makefile}"
failures=0

fail() { printf 'FAIL: %s\n' "$*" >&2; failures=$((failures + 1)); }
pass() { printf 'PASS: %s\n' "$*"; }

# --- The reusable pin is a full SHA. ---
pin="$(grep -E 'uses: eshu-hq/\.github/\.github/workflows/go-multi-module-ci\.yml@' "$WORKFLOW" | sed -E 's/.*@([0-9a-f]+).*/\1/' | head -n 1)"
if [ -z "$pin" ]; then
  fail "no reusable uses: pin in $WORKFLOW"
elif [ "${#pin}" -ne 40 ] || printf '%s' "$pin" | grep -q '[^0-9a-f]'; then
  fail "reusable pin is not 40-hex: $pin"
else
  pass "reusable pin is 40-hex ($pin)"
fi

# --- The wrapper input set is exactly the twelve known inputs. ---
with_block="$(awk '/^  gates:/{injob=1; next} injob && /^    with:/{inwith=1; next} inwith && /^      [a-z-]+:/{print} inwith && /^[^ ]/{exit}' "$WORKFLOW")"
keys="$(printf '%s\n' "$with_block" | sed -E 's/^      ([a-z-]+):.*/\1/' | sort | tr '\n' ' ')"
expected="coverage-floors file-cap-command file-cap-deps-command fixture-subpaths generate-check-command generate-check-go-module generate-check-modules modules primary-module satellite-ref satellite-repository tidy-check-command "
if [ "$keys" = "$expected" ]; then
  pass "wrapper input set matches the twelve known inputs"
else
  fail "wrapper input drift: got [$keys]"
fi

# --- modules is required and non-empty JSON. ---
mods="$(grep -E '^      modules:' "$WORKFLOW" | sed -E "s/.*modules: *'([^']+)'.*/\1/")"
if [ -z "$mods" ]; then
  fail "required modules input missing"
elif python3 -c 'import json,sys; ms=json.load(open(sys.argv[1])); sys.exit(0 if isinstance(ms,list) and ms else 1)' <(printf '%s' "$mods"); then
  pass "modules input is a non-empty list ($mods)"
else
  fail "modules input is not a non-empty JSON list: $mods"
fi

# --- The satellite pin is a full SHA. ---
sat="$(grep -E "^      satellite-ref:" "$WORKFLOW" | sed -E "s/.*satellite-ref: *'([^']+)'.*/\1/")"
if [ -z "$sat" ]; then
  fail "satellite-ref input missing"
elif [ "${#sat}" -ne 40 ] || printf '%s' "$sat" | grep -q '[^0-9a-f]'; then
  fail "satellite-ref is not 40-hex: $sat"
else
  pass "satellite-ref is 40-hex ($sat)"
fi

# --- coverage-floors parses and names both modules (values: floors-sync). ---
floors="$(grep -E '^      coverage-floors:' "$WORKFLOW" | sed -E "s/.*coverage-floors: *'([^']+)'.*/\1/")"
if [ -z "$floors" ]; then
  fail "coverage-floors input missing"
elif python3 -c 'import json,sys; f=json.load(open(sys.argv[1])); sys.exit(0 if set(("collector","factschema")) <= set(f) else 1)' <(printf '%s' "$floors"); then
  pass "coverage-floors names both modules"
else
  fail "coverage-floors does not name both modules: $floors"
fi

# --- Every floor in the workflow has a Makefile default to sync against. ---
if grep -q '^COVERAGE_FLOOR_COLLECTOR' "$MAKEFILE" && grep -q '^COVERAGE_FLOOR_FACTSCHEMA' "$MAKEFILE"; then
  pass "Makefile carries both COVERAGE_FLOOR_* defaults"
else
  fail "Makefile misses a COVERAGE_FLOOR_* default"
fi

if [ "$failures" -gt 0 ]; then
  printf '%d wrapper-input assertion(s) failed\n' "$failures" >&2
  exit 1
fi
printf 'wrapper-inputs: all assertions green\n'
