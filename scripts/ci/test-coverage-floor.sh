#!/usr/bin/env bash
# RED/GREEN test for the per-module coverage floors (issue #34, epic #27).
#
# Exercises the real scripts/ci/check-coverage.sh against a synthetic
# coverprofile with a known total (50.0%: 2 of 4 statements covered):
# a floor at or below the total passes (baseline floors pass), a floor above
# it fails (ratcheting floors fail the lane). Floor comparison is inclusive:
# total == floor passes.
#
# Run: bash scripts/ci/test-coverage-floor.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/ci/check-coverage.sh"
failures=0

fail() { printf 'FAIL: %s\n' "$*" >&2; failures=$((failures + 1)); }
pass() { printf 'PASS: %s\n' "$*"; }

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# go tool cover resolves profile paths against real source, so the synthetic
# profile points at a real temp file: 2 covered + 2 uncovered => total 50.0%.
cat > "$WORK/a.go" <<'EOF'
package p

func Covered() int {
	return 1
}

func Uncovered() int {
	return 2
}
EOF
cat > "$WORK/synthetic.out" <<EOF
mode: set
$WORK/a.go:3.19,5.2 2 1
$WORK/a.go:7.21,9.2 2 0
EOF

total="$(go tool cover -func="$WORK/synthetic.out" | awk '/^total:/ { sub(/%$/, "", $3); print $3 }')"
[ "$total" = "50.0" ] \
  && pass "synthetic profile totals 50.0%" \
  || fail "synthetic profile total wrong: $total"

# --- GREEN: floor at or below the total passes. ---
for floor in 50.0 49.9 48.3; do
  if bash "$SCRIPT" "$WORK/synthetic.out" "$floor" >/dev/null 2>&1; then
    pass "floor $floor passes at total 50.0%"
  else
    fail "floor $floor should pass at total 50.0%"
  fi
done

# --- RED: floor above the total fails the lane. ---
for floor in 50.1 85.0 90.0; do
  if bash "$SCRIPT" "$WORK/synthetic.out" "$floor" >/dev/null 2>&1; then
    fail "floor $floor should fail at total 50.0%"
  else
    pass "floor $floor fails at total 50.0%"
  fi
done

# --- Missing profile and bad usage fail. ---
if bash "$SCRIPT" "$WORK/does-not-exist.out" 70.0 >/dev/null 2>&1; then
  fail "missing coverprofile should fail"
else
  pass "missing coverprofile fails"
fi
if bash "$SCRIPT" only-one-arg >/dev/null 2>&1; then
  fail "single-arg usage should fail"
else
  pass "single-arg usage fails"
fi

if [ "$failures" -gt 0 ]; then
  printf '%d coverage-floor assertion(s) failed\n' "$failures" >&2
  exit 1
fi
printf 'coverage-floor: all assertions green\n'
