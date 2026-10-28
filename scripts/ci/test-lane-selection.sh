#!/usr/bin/env bash
# Shape test for .github/workflows/ci.yml lane selection.
#
# Pins the wire-contract fixes from PR #44 review:
#   1. Fixture-only changes select consumer-compat and both race lanes.
#   2. Agent-instruction files (AGENTS.md/CLAUDE.md/CONTRIBUTING.md) select
#      the shared lane instead of skipping every lane.
#   3. Go changes select the gosec lane (advisory SARIF, never blocking).
#   4. govulncheck is version-pinned (no floating @latest).
#   5. actionlint tarball is checksum-verified.
#   6. Race lanes carry no `actions: write` permission.
#
# The classifier cases are extracted from ci.yml itself and eval'd, so this
# test exercises the real shipped patterns, not a copy of them.

set -euo pipefail

CI_YML="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)/.github/workflows/ci.yml"
failures=0

fail() { printf 'FAIL: %s\n' "$*" >&2; failures=$((failures + 1)); }
pass() { printf 'PASS: %s\n' "$*"; }

# Extract the real classifier case blocks from ci.yml (from the `known=false`
# init through the fail-closed misc line) for eval against sample paths.
CASE_BLOCKS="$(awk '/^[[:space:]]*known=false$/{on=1} on{print} /Anything unrecognized/{on=0}' "$CI_YML")"
[ -n "$CASE_BLOCKS" ] || { fail "could not extract classifier blocks from ci.yml"; }

# Classify one path with the extracted blocks; prints KEY=value lines.
classify_file() {
  local f="$1"
  local go_collector=false fixture_collector=false
  local go_factschema=false fixture_factschema=false
  local workflows=false shell=false shared=false misc=false known=false
  eval "$CASE_BLOCKS" >/dev/null 2>&1
  printf 'go_collector=%s fixture_collector=%s go_factschema=%s fixture_factschema=%s workflows=%s shell=%s shared=%s misc=%s\n' \
    "$go_collector" "$fixture_collector" "$go_factschema" "$fixture_factschema" \
    "$workflows" "$shell" "$shared" "$misc"
}

# Print the `if:` line of a lane job in ci.yml.
lane_if() {
  awk -v job="^  $1:" '$0 ~ job{injob=1} injob && /if:/{print; injob=0}' "$CI_YML"
}

# --- 1. Fixture-only changes select consumer-compat and both race lanes. ---
out="$(classify_file 'factschema/schema/x.json')"
[ "$out" = "go_collector=false fixture_collector=false go_factschema=false fixture_factschema=true workflows=false shell=false shared=false misc=false" ] \
  && pass "factschema fixture classifies to fixture_factschema only" \
  || fail "factschema fixture classification wrong: $out"

out="$(classify_file 'collector/testdata/y.json')"
[ "$(printf '%s' "$out" | grep -o 'fixture_collector=true')" = "fixture_collector=true" ] \
  && [ "$(printf '%s' "$out" | grep -o 'go_collector=true' || true)" = "" ] \
  && pass "collector fixture classifies to fixture_collector only" \
  || fail "collector fixture classification wrong: $out"

for lane in consumer-compat race-collector race-factschema; do
  cond="$(lane_if "$lane")"
  case "$cond" in
    *"fixture_collector == 'true'"*"fixture_factschema == 'true'"*)
      pass "$lane selects on fixture outputs" ;;
    *) fail "$lane gate misses fixture outputs: $cond" ;;
  esac
done

# --- 2. Agent-instruction files select the shared lane. ---
for f in AGENTS.md collector/AGENTS.md factschema/AGENTS.md CLAUDE.md CONTRIBUTING.md; do
  out="$(classify_file "$f")"
  case "$out" in *"shared=true"*) pass "$f selects shared" ;; *) fail "$f skips every lane: $out" ;; esac
done

# Pure-docs skip is preserved for inert prose.
out="$(classify_file 'README.md')"
case "$out" in *"=true"*) fail "README.md selects a lane: $out" ;; *) pass "README.md still skips every lane" ;; esac

# The agent-instruction case must precede the docs-skip case (first match wins).
agent_line="$(grep -n 'AGENTS.md' "$CI_YML" | head -1 | cut -d: -f1)"
docs_line="$(grep -n 'docs/\*' "$CI_YML" | head -1 | cut -d: -f1)"
[ -n "$agent_line" ] && [ -n "$docs_line" ] && [ "$agent_line" -lt "$docs_line" ] \
  && pass "agent-instruction case precedes docs-skip case" \
  || fail "agent-instruction case missing or after docs-skip (agent=$agent_line docs=$docs_line)"

# --- 3. Go changes select gosec; gosec stays advisory with SARIF. ---
for lane in gosec; do
  cond="$(lane_if "$lane")"
  case "$cond" in
    *"go_collector == 'true'"*"go_factschema == 'true'"*)
      pass "gosec selects on Go changes" ;;
    *) fail "gosec gate misses Go outputs: $cond" ;;
  esac
done
grep -q '|| true' "$CI_YML" && pass "gosec keeps non-blocking semantics" || fail "gosec lost || true"
grep -q 'upload-artifact' "$CI_YML" && pass "gosec keeps SARIF artifact upload" || fail "gosec lost artifact upload"

# --- 4. govulncheck pinned. ---
grep -q 'govulncheck@latest' "$CI_YML" \
  && fail "govulncheck still floats @latest" \
  || pass "govulncheck has no floating @latest"
grep -Eq 'govulncheck@v[0-9]+\.[0-9]+\.[0-9]+' "$CI_YML" \
  && pass "govulncheck carries an exact version pin" \
  || fail "govulncheck has no exact @vX.Y.Z pin"

# --- 5. actionlint checksum. ---
grep -q 'sha256sum' "$CI_YML" \
  && pass "actionlint install verifies sha256" \
  || fail "actionlint install has no checksum verification"

# --- 6. Race lanes least privilege. ---
for lane in race-collector race-factschema; do
  block="$(awk -v job="^  $lane:" '$0 ~ job{injob=1; print; next} injob && /^  [a-z-]+:/{exit} injob{print}' "$CI_YML")"
  case "$block" in
    *"actions: write"*) fail "$lane still requests actions: write" ;;
    *) pass "$lane has no actions: write" ;;
  esac
done

if [ "$failures" -gt 0 ]; then
  printf '%d shape assertion(s) failed\n' "$failures" >&2
  exit 1
fi
printf 'lane-selection shape: all assertions green\n'
