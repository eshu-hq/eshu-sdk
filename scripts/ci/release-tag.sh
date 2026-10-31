#!/usr/bin/env bash
# Create the lightweight release tag <module>/<version> at <commit-sha> through
# the GitHub git refs API (.github/workflows/release.yml, tag job).
#
# Usage: release-tag.sh <module> <version> <commit-sha>
#
# The tag is created by POSTing a new ref, which GitHub rejects when the ref
# exists, so an existing tag is never moved. The script probes first. A ref
# that already points at the verified commit as a commit object is accepted
# without a POST, so re-running a failed tag job (for example after a failed
# read-back) is safe; a ref anywhere else is a refusal. It reads the ref back to
# confirm it points at the verified commit as a plain commit object, like the
# rc.8 to rc.10 tags.
# Exit 1 is a refusal or a failed check, 2 a malformed invocation.
#
# Environment (all optional; the self-test sets them):
#   GH_BIN  gh executable (default: gh)
#   REPO    owner/name (default: $GITHUB_REPOSITORY)

set -euo pipefail

# shellcheck source=scripts/ci/release-lib.sh
. "$(dirname "${BASH_SOURCE[0]}")/release-lib.sh"

GH_BIN="${GH_BIN:-gh}"
REPO="${REPO:-${GITHUB_REPOSITORY:-}}"

fail() {
  printf 'release-tag: FAIL: %s\n' "$*" >&2
  exit 1
}

# need runs a lib check and turns its reason into the failure message.
need() {
  local reason
  reason="$("$@" 2>&1)" || fail "$reason"
}

if [ "$#" -ne 3 ]; then
  printf 'usage: release-tag.sh <module> <version> <commit-sha>\n' >&2
  exit 2
fi
module="$1"
version="$2"
sha="$3"
need release_check_module "$module"
need release_check_version "$version"
need release_check_sha "$sha"
[ -n "$REPO" ] || fail "REPO (or GITHUB_REPOSITORY) is not set"

tag="${module}/${version}"
ref="refs/tags/${tag}"

# matching-refs matches by prefix (rc.1 also matches rc.10), so compare whole.
existing="$("$GH_BIN" api "repos/${REPO}/git/matching-refs/tags/${tag}" 2>&1)" || fail "probing for tag $tag failed: $existing"
found="$(printf '%s' "$existing" | jq -r --arg ref "$ref" '[.[]? | select(.ref == $ref)] | first // empty | "\(.object.sha) \(.object.type)"')" || fail "tag probe response is not JSON"
if [ -n "$found" ]; then
  [ "$found" = "$sha commit" ] || fail "tag $tag already exists at '$found', not the verified commit $sha; a published tag is never retargeted"
  printf 'release-tag: %s already points at %s; nothing to create\n' "$ref" "$sha"
else
  created="$("$GH_BIN" api -X POST "repos/${REPO}/git/refs" -f "ref=${ref}" -f "sha=${sha}" 2>&1)" || fail "creating $ref failed: $created"
  printf 'release-tag: created %s at %s\n' "$ref" "$sha"
fi

readback="$("$GH_BIN" api "repos/${REPO}/git/ref/tags/${tag}" 2>&1)" || fail "reading back $ref failed: $readback"
got_sha="$(printf '%s' "$readback" | jq -r '.object.sha // empty')"
got_type="$(printf '%s' "$readback" | jq -r '.object.type // empty')"
[ "$got_sha" = "$sha" ] || fail "$ref points at '$got_sha', not the verified commit $sha"
[ "$got_type" = "commit" ] || fail "$ref is a '$got_type' object; release tags are lightweight (commit)"
printf 'release-tag: confirmed %s -> %s (%s)\n' "$ref" "$got_sha" "$got_type"
