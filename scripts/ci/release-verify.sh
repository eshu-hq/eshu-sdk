#!/usr/bin/env bash
# Pre-tag checks for the release workflow (.github/workflows/release.yml).
#
# Usage: release-verify.sh <module> <version> <commit-sha>
#
# Run from a checkout of the release commit with the full history, the
# origin/main ref, and the existing tags fetched. Every check must pass before
# the tag job may create a tag; the first failure exits 1 with the reason, and
# a malformed invocation exits 2. The checks print as they pass, so the run log
# records what was verified.
#
# Environment (all optional; the self-test sets them):
#   GH_BIN         gh executable used for the check-runs read (default: gh)
#   REPO           owner/name for the check-runs read (default: $GITHUB_REPOSITORY)
#   MAIN_REF       ref the commit must be an ancestor of (default: origin/main)
#   TAG_REMOTE     remote probed for the tag (default: origin)
#   REQUIRED_CHECK check run that must have concluded success (default: ci-complete)
#   REQUIRED_APP   app slug that must own that check run (default: github-actions)
#   EXPECT_FILE    optional repo-relative file inside the module (with EXPECT_SHA256)
#   EXPECT_SHA256  the SHA-256 that file must have at the commit
#   BUILD_CMD      build run inside the module dir at the commit
#                  (default: GOWORK=off go build ./...)

set -euo pipefail

# shellcheck source=scripts/ci/release-lib.sh
. "$(dirname "${BASH_SOURCE[0]}")/release-lib.sh"

GH_BIN="${GH_BIN:-gh}"
REPO="${REPO:-${GITHUB_REPOSITORY:-}}"
MAIN_REF="${MAIN_REF:-origin/main}"
TAG_REMOTE="${TAG_REMOTE:-origin}"
REQUIRED_CHECK="${REQUIRED_CHECK:-ci-complete}"
REQUIRED_APP="${REQUIRED_APP:-github-actions}"
EXPECT_FILE="${EXPECT_FILE:-}"
EXPECT_SHA256="${EXPECT_SHA256:-}"
BUILD_CMD="${BUILD_CMD:-GOWORK=off go build ./...}"

usage() {
  printf 'usage: release-verify.sh <module> <version> <commit-sha>\n' >&2
  exit 2
}

fail() {
  printf 'release-verify: FAIL: %s\n' "$*" >&2
  exit 1
}

# need runs a lib check and turns its reason into the failure message.
need() {
  local reason
  reason="$("$@" 2>&1)" || fail "$reason"
}

ok() {
  checks=$((checks + 1))
  printf 'release-verify: ok: %s\n' "$*"
}

[ "$#" -eq 3 ] || usage
module="$1"
version="$2"
sha="$3"
checks=0

# 1. The module is one this repository tags.
need release_check_module "$module"
ok "module $module is in the allowed list"

# 2. The version is a plain semver tag.
need release_check_version "$version"
ok "version $version is well formed"

# 3. The commit is a full SHA that exists in this checkout.
need release_check_sha "$sha"
git cat-file -e "${sha}^{commit}" 2>/dev/null || fail "commit $sha is not in this checkout"
ok "commit $sha exists"

# 4. The commit is already on main, so a tag cannot name unreviewed work.
git rev-parse --verify --quiet "${MAIN_REF}^{commit}" >/dev/null || fail "$MAIN_REF is not fetched"
git merge-base --is-ancestor "$sha" "$MAIN_REF" || fail "commit $sha is not an ancestor of $MAIN_REF"
ok "commit $sha is an ancestor of $MAIN_REF"

# 5. The required check concluded success on that exact SHA. The newest run of
#    the check decides, so a failed rerun after an earlier success blocks.
[ -n "$REPO" ] || fail "REPO (or GITHUB_REPOSITORY) is not set"
runs="$("$GH_BIN" api "repos/${REPO}/commits/${sha}/check-runs?check_name=${REQUIRED_CHECK}&per_page=100" 2>&1)" || fail "reading check runs for $sha failed: $runs"
verdict="$(printf '%s' "$runs" | jq -r --arg name "$REQUIRED_CHECK" --arg app "$REQUIRED_APP" '
  [.check_runs[]? | select(.name == $name and .app.slug == $app)]
  | sort_by(.started_at // "")
  | last
  | if . == null then "missing" else "\(.status) \(.conclusion // "none")" end')" || fail "check-run response for $sha is not JSON"
[ "$verdict" = "completed success" ] || fail "check run $REQUIRED_CHECK on $sha is '$verdict', not 'completed success'"
ok "check run $REQUIRED_CHECK (app $REQUIRED_APP) concluded success on $sha"

# 6. The module's changelog, as committed at the release commit, has an entry
#    for exactly this version. The heading is matched whole, so [0.1.0-rc.1]
#    cannot satisfy [0.1.0-rc.11].
changelog="$(git show "${sha}:${module}/CHANGELOG.md" 2>/dev/null)" || fail "${module}/CHANGELOG.md does not exist at $sha"
heading="## [${version#v}]"
printf '%s\n' "$changelog" | awk -v h="$heading" '
  index($0, h) == 1 && (length($0) == length(h) || substr($0, length(h) + 1, 1) == " ") { found = 1 }
  END { exit found ? 0 : 1 }' || fail "${module}/CHANGELOG.md at $sha has no '$heading' entry"
ok "${module}/CHANGELOG.md names $version"

# 7. The tag does not exist anywhere: not locally, not on the remote. A tag is
#    immutable once published, so a clash is a stop, not a retarget.
tag="${module}/${version}"
if git rev-parse --verify --quiet "refs/tags/${tag}" >/dev/null; then
  fail "tag $tag already exists locally"
fi
remote_probe=0
git ls-remote --exit-code --tags "$TAG_REMOTE" "refs/tags/${tag}" >/dev/null 2>&1 || remote_probe=$?
case "$remote_probe" in
  2) ;; # exit 2 is "no matching ref": the tag is free
  0) fail "tag $tag already exists on $TAG_REMOTE" ;;
  *) fail "could not probe $TAG_REMOTE for tag $tag (git ls-remote exit $remote_probe)" ;;
esac
ok "tag $tag does not exist locally or on $TAG_REMOTE"

# 8. The module is a Go module at the commit with the expected path, the
#    version's major matches that path (Go serves v0 and v1 only for a path with
#    no /vN suffix), and the tree builds on its own (GOWORK=off).
gomod="$(git show "${sha}:${module}/go.mod" 2>/dev/null)" || fail "${module}/go.mod does not exist at $sha"
declared="$(printf '%s\n' "$gomod" | awk '$1 == "module" { print $2; exit }')"
need release_check_major "$version" "$declared" "github.com/eshu-hq/eshu-sdk/${module}"
ok "${module}/go.mod declares $declared, which fits $version"

# 9. The optional file-hash pair is valid and true at the commit, so a wrong
#    value stops the run here, before a tag exists that prove could never match.
need release_check_expect "$module" "$EXPECT_FILE" "$EXPECT_SHA256"
if [ -n "$EXPECT_FILE" ]; then
  [ "$(git cat-file -t "${sha}:${EXPECT_FILE}" 2>/dev/null)" = "blob" ] || fail "${EXPECT_FILE} is not a file at $sha"
  # A file under a nested module is not in this module's zip.
  dir="$(dirname "$EXPECT_FILE")"
  while [ "$dir" != "$module" ] && [ "$dir" != "." ]; do
    if git cat-file -e "${sha}:${dir}/go.mod" 2>/dev/null; then
      fail "${EXPECT_FILE} is inside the nested module ${dir}, which the ${module} zip omits"
    fi
    dir="$(dirname "$dir")"
  done
  got="$(git show "${sha}:${EXPECT_FILE}" | { sha256sum 2>/dev/null || shasum -a 256; } | cut -d' ' -f1)"
  [ "$got" = "$EXPECT_SHA256" ] || fail "${EXPECT_FILE} at $sha has SHA-256 $got, expected $EXPECT_SHA256"
  ok "${EXPECT_FILE} at $sha has the expected SHA-256"
fi

# The build runs against the worktree, so require it to be the release commit.
head_sha="$(git rev-parse HEAD)"
[ "$head_sha" = "$sha" ] || fail "checkout is at $head_sha, not the release commit $sha"
if ! build_out="$(cd "$module" && bash -c "$BUILD_CMD" 2>&1)"; then
  printf '%s\n' "$build_out" | tail -n 20 >&2
  fail "${module} at $sha does not build with: $BUILD_CMD"
fi
ok "${module} at $sha builds ($BUILD_CMD)"

printf 'release-verify: all %d checks passed for %s at %s\n' "$checks" "$tag" "$sha"
