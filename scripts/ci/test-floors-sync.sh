#!/usr/bin/env bash
# Floors-sync guard (issue #39 finding F1, epic #27).
#
# The reusable race lanes enforce coverage floors inline from the
# `coverage-floors` workflow input, while local `make race` enforces the
# Makefile COVERAGE_FLOOR_* defaults via scripts/ci/check-coverage.sh.
# This test fails when the two drift apart, so a floor bump in one place
# without the other cannot go green.
#
# Run: bash scripts/ci/test-floors-sync.sh [Makefile] [workflow]

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MAKEFILE="${1:-$ROOT/Makefile}"
WORKFLOW="${2:-$ROOT/.github/workflows/ci.yml}"
failures=0

fail() { printf 'FAIL: %s\n' "$*" >&2; failures=$((failures + 1)); }
pass() { printf 'PASS: %s\n' "$*"; }

floors_json="$(sed -nE "s/^[[:space:]]*coverage-floors:[[:space:]]*'(\{[^']*\})'.*$/\1/p" "$WORKFLOW" | head -n 1)"
if [ -z "$floors_json" ]; then
  fail "no coverage-floors input in $WORKFLOW"
else
  pass "coverage-floors input present in workflow"
fi

for module in collector factschema wirecontract; do
  upper="$(printf '%s' "$module" | tr 'a-z' 'A-Z')"
  mk_floor="$(grep -E "^COVERAGE_FLOOR_${upper}[[:space:]]*\\?= " "$MAKEFILE" | sed -E 's/.*\\?=[[:space:]]*//' || true)"
  if [ -z "$mk_floor" ]; then
    fail "no COVERAGE_FLOOR_${upper} default in $MAKEFILE"
    continue
  fi
  if [ -z "$floors_json" ]; then
    fail "cannot compare $module floor without workflow input"
    continue
  fi
  wf_floor="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get(sys.argv[2], ""))' <(printf '%s' "$floors_json") "$module")"
  if [ -z "$wf_floor" ]; then
    fail "no $module entry in workflow coverage-floors"
  elif [ "$mk_floor" = "$wf_floor" ]; then
    pass "$module floor in sync ($mk_floor)"
  else
    fail "$module floor drift: Makefile $mk_floor vs workflow $wf_floor"
  fi
done

if [ "$failures" -gt 0 ]; then
  printf '%d floors-sync assertion(s) failed\n' "$failures" >&2
  exit 1
fi
printf 'floors-sync: all assertions green\n'
