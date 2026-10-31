#!/usr/bin/env bash
# Self-test for scripts/ci/release-verify.sh and the shape of
# .github/workflows/release.yml.
#
# Builds a throwaway origin and clone, stubs gh and the build, and runs the
# verifier against one passing case and a set of seeded violations (each one
# must fail, with the reason named). The workflow shape checks pin the
# properties the workflow's safety rests on: dispatch-only, no default token
# permissions, SHA-pinned actions, write access on the tag job alone, an
# approval-capable environment, and a module list equal to the script's.
#
# Run: bash scripts/ci/test-release-verify.sh [workflow]
# Runs in CI in the `shape` job and locally via `make shape`.

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORKFLOW="${1:-$ROOT/.github/workflows/release.yml}"
VERIFY="$ROOT/scripts/ci/release-verify.sh"
failures=0

fail() { printf 'FAIL: %s\n' "$*" >&2; failures=$((failures + 1)); }
pass() { printf 'PASS: %s\n' "$*"; }

tmp="$(mktemp -d "${TMPDIR:-/tmp}/release-verify-test.XXXXXX")"
trap 'rm -rf "$tmp"' EXIT

export GIT_AUTHOR_NAME=test GIT_AUTHOR_EMAIL=test@example.invalid
export GIT_COMMITTER_NAME=test GIT_COMMITTER_EMAIL=test@example.invalid
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_SYSTEM=/dev/null

# --- Fixture repository: origin (bare), a clone on main, and a side branch. ---
git init -q --bare "$tmp/origin.git"
git clone -q "$tmp/origin.git" "$tmp/work" 2>/dev/null
cd "$tmp/work" || exit 1
git checkout -q -b main
mkdir factschema collector wirecontract badpath
for m in factschema collector wirecontract; do
  printf 'module github.com/eshu-hq/eshu-sdk/%s\n\ngo 1.26.10\n' "$m" >"$m/go.mod"
done
printf 'module github.com/eshu-hq/eshu-sdk/other\n\ngo 1.26.10\n' >badpath/go.mod
mkdir -p factschema/fixturepack/payloads factschema/nested
printf 'golden bytes\n' >factschema/fixturepack/payloads/g.json
printf 'module example.com/nested\n' >factschema/nested/go.mod
printf 'nested bytes\n' >factschema/nested/n.json
printf '# Changelog\n\n## [Unreleased]\n\n## [0.1.0-rc.11] - Prepared (unpublished)\n\n## [0.1.0-rc.1] - 2026-10-01\n\n## [0.3.0]junk\n' >factschema/CHANGELOG.md
printf '# Changelog\n\n## [0.1.0-rc.1] - 2026-10-01\n' >collector/CHANGELOG.md
git add -A
git commit -q -m "main commit"
main_sha="$(git rev-parse HEAD)"
git push -q origin main 2>/dev/null
git checkout -q -b side
printf 'x\n' >side.txt
git add -A
git commit -q -m "side commit"
side_sha="$(git rev-parse HEAD)"
git checkout -q --detach "$main_sha"
git fetch -q origin

# --- gh stub: serves $STUB_CHECKS for the check-runs read. ---
cat >"$tmp/gh" <<'EOF'
#!/usr/bin/env bash
if [ -n "${STUB_GH_FAIL:-}" ]; then echo "gh: simulated failure" >&2; exit 1; fi
case "$*" in
  api\ repos/*/commits/*/check-runs\?check_name=ci-complete\&per_page=100) cat "$STUB_CHECKS" ;;
  *) echo "gh stub: unexpected call: $*" >&2; exit 99 ;;
esac
EOF
chmod +x "$tmp/gh"

checks_file="$tmp/checks.json"
set_checks() { printf '%s\n' "$1" >"$checks_file"; }
green='{"check_runs":[{"name":"ci-complete","app":{"slug":"github-actions"},"status":"completed","conclusion":"success","started_at":"2026-10-31T10:00:00Z"}]}'

# run <expect: pass|fail> <name> <reason-substring> <args...>
# Environment overrides go in front via env(1) at the call site.
run() {
  local expect="$1" name="$2" want="$3"
  shift 3
  local out rc
  out="$("$@" 2>&1)"
  rc=$?
  if [ "$expect" = pass ]; then
    if [ "$rc" -eq 0 ]; then pass "$name"; else fail "$name: expected pass, exit $rc: $out"; fi
    return
  fi
  if [ "$rc" -eq 0 ]; then
    fail "$name: seeded violation passed"
  elif ! printf '%s' "$out" | grep -Fq -- "$want"; then
    fail "$name: failed (exit $rc) without naming '$want': $out"
  else
    pass "$name"
  fi
}

verify() { # verify <module> <version> <sha>  (stubbed gh, passing build)
  env GH_BIN="$tmp/gh" REPO=eshu-hq/eshu-sdk STUB_CHECKS="$checks_file" BUILD_CMD=true \
    bash "$VERIFY" "$@"
}

set_checks "$green"
run pass "all checks pass" "" verify factschema v0.1.0-rc.11 "$main_sha"
run pass "a prefix of another entry is its own entry (rc.1)" "" verify collector v0.1.0-rc.1 "$main_sha"

sha256_text() { if command -v sha256sum >/dev/null 2>&1; then sha256sum | cut -d' ' -f1; else shasum -a 256 | cut -d' ' -f1; fi; }
good="$(printf 'golden bytes\n' | sha256_text)"
nested_hash="$(printf 'nested bytes\n' | sha256_text)"
vx() { env EXPECT_FILE="$1" EXPECT_SHA256="$2" GH_BIN="$tmp/gh" REPO=eshu-hq/eshu-sdk STUB_CHECKS="$checks_file" BUILD_CMD=true bash "$VERIFY" factschema v0.1.0-rc.11 "$main_sha"; }
run pass "a stable version is well formed (changelog gate stays separate)" "" bash -c "source '$ROOT/scripts/ci/release-lib.sh' && release_check_version v1.2.3 && release_check_version v0.1.0-rc.11"

# --- The optional file hash is true at the commit, before any tag exists ---
run pass "expect: the file hash matches" "has the expected SHA-256" vx factschema/fixturepack/payloads/g.json "$good"
run fail "expect: a well-formed but wrong hash" "has SHA-256" vx factschema/fixturepack/payloads/g.json "$(printf '%064d' 0)"
run fail "expect: a file that does not exist" "is not a file" vx factschema/fixturepack/payloads/nope.json "$good"
run fail "expect: a directory, not a file" "is not a file" vx factschema/fixturepack/payloads "$good"
run fail "expect: path escape with .." "clean relative path" vx factschema/../collector/go.mod "$good"
run fail "expect: doubled slash in the path" "clean relative path" vx factschema//fixturepack/payloads/g.json "$good"
run fail "expect: a dot segment in the path" "clean relative path" vx factschema/./fixturepack/payloads/g.json "$good"
run fail "expect: path outside the module" "must start with" vx collector/go.mod "$good"
run fail "expect: a file in a nested module" "nested module" vx factschema/nested/n.json "$nested_hash"
run fail "expect: file without a hash" "go together" vx factschema/fixturepack/payloads/g.json ""
run fail "expect: hash without a file" "go together" vx "" "$good"
run fail "expect: uppercase hash" "64 lowercase hex" vx factschema/fixturepack/payloads/g.json "$(printf '%s' "$good" | tr a-f A-F)"
run fail "expect: 65-hex hash" "64 lowercase hex" vx factschema/fixturepack/payloads/g.json "${good}0"
run fail "expect: hash with a trailing newline" "64 lowercase hex" vx factschema/fixturepack/payloads/g.json "$good"$'\n'

# --- Version and SHA strings (the lib is the only gate the write job relies on) ---
# shellcheck source=scripts/ci/release-lib.sh
. "$ROOT/scripts/ci/release-lib.sh"
for good_version in v0.1.0 v0.1.0-rc.11 v1.2.3 v0.0.1-alpha v0.1.0-rc.0 v10.20.30 v0.1.0-rc.1a2; do
  run pass "lib accepts version $good_version" "" release_check_version "$good_version"
done
for bad_version in v2.0.0x V0.1.0 ' v0.1.0' 'v0.1.0 ' v1 v1.2 v1.2.3.4 v01.0.0 v0.01.0 v0.1.00 v0.1.0-rc.01 v0.1.0-01 'v0.1.0+meta' v0.1.0- v0.1.0-rc. v0.1.0-rc..1 v0.1.0-RC.1_x $'v0.1.0-rc.11\nx' $'x\nv0.1.0' ''; do
  run fail "lib rejects version '$bad_version'" "is not vMAJOR" release_check_version "$bad_version"
done
for good_sha in "$main_sha" 0123456789abcdef0123456789abcdef01234567; do
  run pass "lib accepts commit $good_sha" "" release_check_sha "$good_sha"
done
for bad_sha in "${main_sha}0" "${main_sha:1}" "x${main_sha}" "${main_sha}x" "${main_sha}"$'\n'main "main"$'\n'"${main_sha}" "${main_sha}"$'\n' ""; do
  run fail "lib rejects commit '${bad_sha:0:50}'" "40-hex" release_check_sha "$bad_sha"
done
run pass "lib: v0 and v1 fit an unsuffixed path" "" bash -c ". '$ROOT/scripts/ci/release-lib.sh'; release_check_major v0.1.0 github.com/eshu-hq/eshu-sdk/m github.com/eshu-hq/eshu-sdk/m && release_check_major v1.9.9 github.com/eshu-hq/eshu-sdk/m github.com/eshu-hq/eshu-sdk/m"
run fail "lib: v2 is refused on a plain path" "major >= 2" release_check_major v2.0.0 github.com/eshu-hq/eshu-sdk/m github.com/eshu-hq/eshu-sdk/m
run fail "lib: v2 is refused on a /v2 path too (prove cannot import it yet)" "major >= 2" release_check_major v2.1.0 github.com/eshu-hq/eshu-sdk/m/v2 github.com/eshu-hq/eshu-sdk/m
run fail "lib: v10 is refused" "major >= 2" release_check_major v10.0.0 github.com/eshu-hq/eshu-sdk/m github.com/eshu-hq/eshu-sdk/m
run fail "lib: v1 on a /v2 path is a path mismatch" "expected github.com/eshu-hq/eshu-sdk/m" release_check_major v1.0.0 github.com/eshu-hq/eshu-sdk/m/v2 github.com/eshu-hq/eshu-sdk/m
run fail "lib: a /v1 suffix is not accepted" "expected github.com/eshu-hq/eshu-sdk/m" release_check_major v1.0.0 github.com/eshu-hq/eshu-sdk/m/v1 github.com/eshu-hq/eshu-sdk/m
run fail "lib: a path that merely starts like the base" "expected github.com/eshu-hq/eshu-sdk/m" release_check_major v0.1.0 github.com/eshu-hq/eshu-sdk/mx github.com/eshu-hq/eshu-sdk/m
run fail "lib: the base with a dot wildcard" "expected github.com/eshu-hq/eshu-sdk/m" release_check_major v0.1.0 github.com/eshu-hq/eshu-sdkXm github.com/eshu-hq/eshu-sdk/m

# --- Inputs ---
run fail "module not in the list" "is not one of" verify evil v0.1.0-rc.11 "$main_sha"
run fail "module path traversal" "is not one of" verify ../collector v0.1.0-rc.11 "$main_sha"
run fail "module with a slash" "is not one of" verify factschema/../collector v0.1.0-rc.11 "$main_sha"
run fail "empty module" "is not one of" verify "" v0.1.0-rc.11 "$main_sha"
for bad in latest v1 v1.2 0.1.0-rc.11 v01.0.0 'v0.1.0-rc.11+meta' 'v0.1.0 ' main; do
  run fail "bad version '$bad'" "is not vMAJOR" verify factschema "$bad" "$main_sha"
done
run fail "short commit" "40-hex" verify factschema v0.1.0-rc.11 "${main_sha:0:12}"
run fail "uppercase commit" "40-hex" verify factschema v0.1.0-rc.11 "$(printf '%s' "$main_sha" | tr a-f A-F)"
run fail "branch name as commit" "40-hex" verify factschema v0.1.0-rc.11 main
run fail "unknown commit" "not in this checkout" verify factschema v0.1.0-rc.11 0000000000000000000000000000000000000000
run fail "wrong argument count" "usage" verify factschema v0.1.0-rc.11
run fail "commit not on main" "not an ancestor" bash -c "cd '$tmp/work' && git checkout -q --detach '$side_sha' && env GH_BIN='$tmp/gh' REPO=eshu-hq/eshu-sdk STUB_CHECKS='$checks_file' BUILD_CMD=true bash '$VERIFY' factschema v0.1.0-rc.11 '$side_sha'; rc=\$?; git checkout -q --detach '$main_sha'; exit \$rc"
run fail "main ref not fetched" "is not fetched" env MAIN_REF=origin/nope GH_BIN="$tmp/gh" REPO=eshu-hq/eshu-sdk STUB_CHECKS="$checks_file" BUILD_CMD=true bash "$VERIFY" factschema v0.1.0-rc.11 "$main_sha"

# --- CI state ---
set_checks '{"check_runs":[{"name":"ci-complete","app":{"slug":"github-actions"},"status":"completed","conclusion":"failure","started_at":"2026-10-31T10:00:00Z"}]}'
run fail "ci-complete failed" "not 'completed success'" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks '{"check_runs":[{"name":"ci-complete","app":{"slug":"github-actions"},"status":"in_progress","conclusion":null,"started_at":"2026-10-31T10:00:00Z"}]}'
run fail "ci-complete still running" "in_progress" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks '{"check_runs":[{"name":"gates","status":"completed","conclusion":"success","started_at":"2026-10-31T10:00:00Z"}]}'
run fail "ci-complete absent" "'missing'" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks '{"check_runs":[]}'
run fail "no check runs" "'missing'" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks '{"check_runs":[{"name":"gates / ci-complete","app":{"slug":"github-actions"},"status":"completed","conclusion":"success","started_at":"2026-10-31T10:00:00Z"}]}'
run fail "a similarly named check does not count" "'missing'" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks '{"check_runs":[{"name":"ci-complete","app":{"slug":"some-other-app"},"status":"completed","conclusion":"success","started_at":"2026-10-31T10:00:00Z"}]}'
run fail "ci-complete from another app does not count" "'missing'" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks '{"check_runs":[{"name":"ci-complete","status":"completed","conclusion":"success","started_at":"2026-10-31T10:00:00Z"}]}'
run fail "ci-complete with no app does not count" "'missing'" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks '{"check_runs":[{"name":"ci-complete","app":{"slug":"github-actions"},"status":"completed","conclusion":"success","started_at":"2026-10-31T10:00:00Z"},{"name":"ci-complete","app":{"slug":"github-actions"},"status":"completed","conclusion":"failure","started_at":"2026-10-31T11:00:00Z"}]}'
run fail "a later failed rerun blocks" "not 'completed success'" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks '{"check_runs":[{"name":"ci-complete","app":{"slug":"github-actions"},"status":"completed","conclusion":"failure","started_at":"2026-10-31T10:00:00Z"},{"name":"ci-complete","app":{"slug":"github-actions"},"status":"completed","conclusion":"success","started_at":"2026-10-31T11:00:00Z"}]}'
run pass "a later green rerun clears an earlier failure" "" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks 'not json'
run fail "check-run response is not JSON" "not JSON" verify factschema v0.1.0-rc.11 "$main_sha"
set_checks "$green"
run fail "check-run read fails" "reading check runs" env STUB_GH_FAIL=1 GH_BIN="$tmp/gh" REPO=eshu-hq/eshu-sdk STUB_CHECKS="$checks_file" BUILD_CMD=true bash "$VERIFY" factschema v0.1.0-rc.11 "$main_sha"

# --- Changelog ---
run fail "changelog lacks the version" "no '## [0.1.0-rc.11]' entry" verify collector v0.1.0-rc.11 "$main_sha"
run fail "changelog does not exist" "does not exist" verify wirecontract v0.1.0-rc.1 "$main_sha"
run fail "a heading with trailing junk is not an entry" "no '## [0.3.0]' entry" verify factschema v0.3.0 "$main_sha"

# --- Tag state ---
git -C "$tmp/work" tag factschema/v0.1.0-rc.11 "$main_sha"
run fail "tag exists locally" "already exists locally" verify factschema v0.1.0-rc.11 "$main_sha"
git -C "$tmp/work" tag -d factschema/v0.1.0-rc.11 >/dev/null
git -C "$tmp/work" push -q origin "$main_sha:refs/tags/factschema/v0.1.0-rc.11" 2>/dev/null
run fail "tag exists on the remote only" "already exists on origin" verify factschema v0.1.0-rc.11 "$main_sha"
git -C "$tmp/work" push -q origin ":refs/tags/factschema/v0.1.0-rc.11" 2>/dev/null
run fail "remote unreachable" "could not probe" env TAG_REMOTE="$tmp/missing.git" GH_BIN="$tmp/gh" REPO=eshu-hq/eshu-sdk STUB_CHECKS="$checks_file" BUILD_CMD=true bash "$VERIFY" factschema v0.1.0-rc.11 "$main_sha"
run pass "tag freed again" "" verify factschema v0.1.0-rc.11 "$main_sha"

# --- Module tree ---
run fail "build fails" "does not build" env GH_BIN="$tmp/gh" REPO=eshu-hq/eshu-sdk STUB_CHECKS="$checks_file" BUILD_CMD=false bash "$VERIFY" factschema v0.1.0-rc.11 "$main_sha"
git -C "$tmp/work" mv factschema/go.mod factschema/go.mod.off
git -C "$tmp/work" commit -q -m "drop go.mod"
nogomod_sha="$(git -C "$tmp/work" rev-parse HEAD)"
git -C "$tmp/work" push -q origin HEAD:main 2>/dev/null
git -C "$tmp/work" fetch -q origin
run fail "go.mod missing at the commit" "go.mod does not exist" verify factschema v0.1.0-rc.11 "$nogomod_sha"
# HEAD is now nogomod_sha; asking for main_sha must refuse to build another tree.
run fail "checkout is not at the commit" "not the release commit" verify factschema v0.1.0-rc.11 "$main_sha"
printf 'module github.com/eshu-hq/eshu-sdk/other\n\ngo 1.26.10\n' >"$tmp/work/wirecontract/go.mod"
printf '# Changelog\n\n## [0.1.0-rc.1] - 2026-10-01\n' >"$tmp/work/wirecontract/CHANGELOG.md"
git -C "$tmp/work" add -A
git -C "$tmp/work" commit -q -m "wirecontract with another module path"
wrong_sha="$(git -C "$tmp/work" rev-parse HEAD)"
git -C "$tmp/work" push -q origin HEAD:main 2>/dev/null
git -C "$tmp/work" fetch -q origin
run fail "go.mod declares another module" "expected github.com/eshu-hq/eshu-sdk/wirecontract" verify wirecontract v0.1.0-rc.1 "$wrong_sha"
mod_commit() { # mod_commit <module line> <changelog headings...> -> sets mod_sha
  printf 'module %s\n\ngo 1.26.10\n' "$1" >"$tmp/work/wirecontract/go.mod"
  shift
  { printf '# Changelog\n\n'; for h in "$@"; do printf '## [%s] - x\n\n' "$h"; done; } >"$tmp/work/wirecontract/CHANGELOG.md"
  git -C "$tmp/work" add -A
  git -C "$tmp/work" commit -q -m "go.mod variant"
  mod_sha="$(git -C "$tmp/work" rev-parse HEAD)"
  git -C "$tmp/work" push -q origin HEAD:main 2>/dev/null
  git -C "$tmp/work" fetch -q origin
}
mod_commit github.com/eshu-hq/eshu-sdk/wirecontract-evil 0.1.0-rc.1
run fail "go.mod path that only starts like the module" "expected github.com/eshu-hq/eshu-sdk/wirecontract" verify wirecontract v0.1.0-rc.1 "$mod_sha"
mod_commit github.com/eshu-hq/eshu-sdk/wirecontract 2.0.0 1.0.0 0.1.0-rc.1
run fail "v2.0.0 on an unsuffixed path is refused" "major >= 2" verify wirecontract v2.0.0 "$mod_sha"
run pass "v1.0.0 on an unsuffixed path" "" verify wirecontract v1.0.0 "$mod_sha"
mod_commit github.com/eshu-hq/eshu-sdk/wirecontract/v2 2.0.0 1.0.0
run fail "v2.0.0 on a /v2 path is refused until prove supports it" "major >= 2" verify wirecontract v2.0.0 "$mod_sha"
run fail "v1.0.0 on a /v2 path" "expected github.com/eshu-hq/eshu-sdk/wirecontract" verify wirecontract v1.0.0 "$mod_sha"

# --- Workflow shape ---
if [ ! -f "$WORKFLOW" ]; then
  fail "workflow $WORKFLOW does not exist"
else
  # Dispatch is the only trigger.
  triggers="$(awk '/^on:/{on=1; next} on && /^[^ #]/{exit} on && /^  [a-z_]+:/{gsub(/[ :]/,""); print}' "$WORKFLOW" | tr '\n' ' ')"
  if [ "$triggers" = "workflow_dispatch " ]; then pass "workflow_dispatch is the only trigger"; else fail "triggers are [$triggers]"; fi

  # No default token permissions at the top level.
  if grep -Eq '^permissions: \{\}$' "$WORKFLOW"; then pass "top-level permissions are empty"; else fail "top-level permissions are not '{}'"; fi

  # Every action is pinned to a full commit SHA.
  unpinned="$(grep -E '^\s+(- )?uses:' "$WORKFLOW" | grep -Ev 'uses: [^@ ]+@[0-9a-f]{40}( |$)' || true)"
  if [ -z "$unpinned" ]; then pass "every uses: is a 40-hex pin"; else fail "unpinned: $unpinned"; fi

  # contents: write appears once, in the tag job, which carries the environment.
  writes="$(grep -v '^[[:space:]]*#' "$WORKFLOW" | grep -c 'contents: write' || true)"
  # shellcheck disable=SC2034  # read through ${!job_block} below
  verify_block="$(awk '/^  verify:/{t=1; next} t && /^  [a-z-]+:/{exit} t{print}' "$WORKFLOW")"
  tag_block="$(awk '/^  tag:/{t=1; next} t && /^  [a-z-]+:/{exit} t{print}' "$WORKFLOW")"
  if [ "$writes" -eq 1 ] && printf '%s\n' "$tag_block" | grep -q 'contents: write'; then pass "contents: write is on the tag job alone"; else fail "contents: write count $writes or not in the tag job"; fi
  if printf '%s\n' "$tag_block" | grep -Eq '^    environment: release$'; then pass "the tag job uses the release environment"; else fail "the tag job lacks environment: release"; fi
  if printf '%s\n' "$tag_block" | grep -Eq '^    needs: \[?verify\]?$'; then pass "the tag job needs verify"; else fail "the tag job does not need verify"; fi
  prove_block="$(awk '/^  prove:/{t=1; next} t && /^  [a-z-]+:/{exit} t{print}' "$WORKFLOW")"
  if printf '%s\n' "$prove_block" | grep -Eq '^    needs: \[?tag\]?$'; then pass "the prove job needs tag"; else fail "the prove job does not need tag"; fi

  # Each job's token permissions are exactly these, so a new scope is a test
  # change a reviewer sees.
  job_perms() { awk -v job="$1" '$0 == "  " job ":" {t=1; next} t && /^  [a-z-]+:/ {exit} t && /^    permissions:/ {p=1; next} t && p && /^    [a-z]/ {exit} t && p {gsub(/^ +/, ""); print}' "$WORKFLOW" | tr '\n' ' '; }
  if [ "$(job_perms verify)" = "contents: read checks: read " ]; then pass "verify has exactly contents: read, checks: read"; else fail "verify permissions are [$(job_perms verify)]"; fi
  if [ "$(job_perms tag)" = "contents: write " ]; then pass "tag has exactly contents: write"; else fail "tag permissions are [$(job_perms tag)]"; fi
  if [ "$(job_perms prove)" = "contents: read " ]; then pass "prove has exactly contents: read"; else fail "prove permissions are [$(job_perms prove)]"; fi

  # Every run step names its shell: the default shell has no pipefail, which
  # would let `| tee` hide a failed proof.
  missing_shell="$(awk '/^      - /{ if (run && !shell) n++; run=0; shell=0 } /^        run:/{run=1} /^        shell: bash$/{shell=1} END{ if (run && !shell) n++; print n+0 }' "$WORKFLOW")"
  if [ "$missing_shell" = "0" ]; then pass "every run step sets shell: bash"; else fail "$missing_shell run step(s) lack shell: bash"; fi

  # No checkout keeps the token in git config.
  checkouts="$(grep -c 'uses: actions/checkout@' "$WORKFLOW" || true)"
  no_persist="$(grep -c 'persist-credentials: false' "$WORKFLOW" || true)"
  if [ "$checkouts" -ge 1 ] && [ "$checkouts" = "$no_persist" ]; then pass "every checkout sets persist-credentials: false"; else fail "$checkouts checkout(s), $no_persist persist-credentials: false"; fi

  # The proof reaches the artifact even on failure.
  # shellcheck disable=SC2016  # a literal Actions expression, not a shell variable
  if grep -Fq 'name: release-proof-${{ inputs.module }}-${{ inputs.version }}' "$WORKFLOW"; then pass "the proof log is uploaded under a module-version name"; else fail "no release-proof artifact name"; fi

  # verify receives the file-hash inputs, so a wrong value cannot reach the tag job.
  verify_env_ok=1
  # shellcheck disable=SC2016  # literal Actions expressions, not shell variables
  for want in 'EXPECT_FILE: ${{ inputs.expect_file }}' 'EXPECT_SHA256: ${{ inputs.expect_sha256 }}'; do
    printf '%s\n' "$verify_block" | grep -Fq "$want" || verify_env_ok=0
  done
  if [ "$verify_env_ok" -eq 1 ]; then pass "verify receives expect_file and expect_sha256"; else fail "verify does not receive both file-hash inputs"; fi

  # Exactly one job names an environment, and it is the tag job: an environment on verify
  # would make the read-only checks wait for an approval.
  env_lines="$(grep -v '^[[:space:]]*#' "$WORKFLOW" | grep -c '^    environment:' || true)"
  if [ "$env_lines" = "1" ] && printf '%s\n' "$tag_block" | grep -Eq '^    environment: release$'; then pass "only the tag job names an environment"; else fail "environment lines: $env_lines (want 1, in the tag job)"; fi

  # Releases of one module serialize and are never cancelled midway.
  if grep -Eq 'group: release-\$\{\{ inputs.module \}\}' "$WORKFLOW" && grep -Eq 'cancel-in-progress: false' "$WORKFLOW"; then pass "same-module releases serialize"; else fail "concurrency group or cancel-in-progress is wrong"; fi

  # Every job refuses to run from another ref. This stops an accidental
  # dispatch of an unedited copy from a branch; it is not a security boundary,
  # because a branch can edit the file and drop the guard.
  for job_block in verify_block tag_block prove_block; do
    if printf '%s\n' "${!job_block}" | grep -Fxq "    if: github.ref == 'refs/heads/main'"; then
      pass "${job_block%_block} refuses to run from another ref"
    else
      fail "${job_block%_block} lacks the github.ref == refs/heads/main guard"
    fi
  done

  # The module choice equals the script's allowed list.
  # shellcheck source=scripts/ci/release-lib.sh
  . "$ROOT/scripts/ci/release-lib.sh"
  options="$(awk '/^      module:/{m=1; next} m && /^      [a-z_]+:/{exit} m && /^          - /{sub(/^          - /,""); print}' "$WORKFLOW" | sort | tr '\n' ' ')"
  allowed="$(printf '%s\n' "${RELEASE_ALLOWED_MODULES[@]}" | sort | tr '\n' ' ')"
  if [ "$options" = "$allowed" ]; then pass "module choice equals the script's allowed list"; else fail "module choice [$options] differs from [$allowed]"; fi
fi

if [ "$failures" -ne 0 ]; then
  printf 'test-release-verify: %d failure(s)\n' "$failures" >&2
  exit 1
fi
printf 'test-release-verify: ok\n'
