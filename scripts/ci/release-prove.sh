#!/usr/bin/env bash
# Prove a published release tag from the outside (.github/workflows/release.yml,
# prove job): a fresh consumer module on an empty module cache, resolving
# github.com/eshu-hq/eshu-sdk/<module>@<version> through the public proxy with
# no replace, no workspace, and the public checksum database.
#
# Usage: release-prove.sh <module> <version> <commit-sha>
#
# Optional file check, both or neither (environment):
#   EXPECT_FILE    path of a file inside the module, repo-relative, so it starts
#                  with "<module>/" (e.g. factschema/fixturepack/payloads/x.json)
#   EXPECT_SHA256  the 64-hex SHA-256 the downloaded file must have
#
# The run log is the record: it prints the proxy's module and go.mod checksums,
# the origin commit the proxy reports (and fails if that is not the verified
# commit), the consumer's go.sum lines, and the result of building and running a
# tiny program that imports the module. A tag can reach the proxy a little after
# it is created, so the resolve step retries.
#
# Environment (all optional; the self-test sets them):
#   GO_BIN          go executable (default: go)
#   PROVE_ATTEMPTS  resolve attempts before giving up (default: 12)
#   PROVE_SLEEP     seconds between attempts (default: 30)
#
# Exit 1 is a failed proof, 2 a malformed invocation.

set -euo pipefail

# shellcheck source=scripts/ci/release-lib.sh
. "$(dirname "${BASH_SOURCE[0]}")/release-lib.sh"

GO_BIN="${GO_BIN:-go}"
PROVE_ATTEMPTS="${PROVE_ATTEMPTS:-12}"
PROVE_SLEEP="${PROVE_SLEEP:-30}"
EXPECT_FILE="${EXPECT_FILE:-}"
EXPECT_SHA256="${EXPECT_SHA256:-}"

fail() {
  printf 'release-prove: FAIL: %s\n' "$*" >&2
  exit 1
}

usage_error() {
  printf 'release-prove: %s\n' "$*" >&2
  exit 2
}

[ "$#" -eq 3 ] || usage_error "usage: release-prove.sh <module> <version> <commit-sha>"
module="$1"
version="$2"
sha="$3"
for check in module version sha; do
  value="${!check}"
  reason="$("release_check_$check" "$value" 2>&1)" || usage_error "$reason"
done

reason="$(release_check_expect "$module" "$EXPECT_FILE" "$EXPECT_SHA256" 2>&1)" || usage_error "$reason"

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

import_path="github.com/eshu-hq/eshu-sdk/${module}"
work="$(mktemp -d "${TMPDIR:-/tmp}/release-prove.XXXXXX")"
trap 'chmod -R u+w "$work" 2>/dev/null; rm -rf "$work"' EXIT

# A clean consumer environment: its own empty module and build caches, the
# public proxy with no direct fallback, the public checksum database, no go env
# file (GOENV=off), and none of the variables that would route around either.
unset GONOSUMDB GONOSUMCHECK GONOPROXY GOPRIVATE GOINSECURE GOFLAGS GOVCS
export GOENV=off
export GOWORK=off
export GOPROXY=https://proxy.golang.org
export GOSUMDB=sum.golang.org
export GOMODCACHE="$work/modcache"
export GOCACHE="$work/gocache"
export GOPATH="$work/gopath"

printf 'release-prove: %s@%s expecting origin %s\n' "$import_path" "$version" "$sha"
mkdir "$work/consumer"
cd "$work/consumer"
"$GO_BIN" mod init example.com/release-proof >/dev/null 2>&1 || fail "go mod init failed"
cat >main.go <<EOF2
package main

import (
	"fmt"

	_ "${import_path}"
)

func main() { fmt.Println("release-proof: imported ${import_path}") }
EOF2

attempt=1
until get_out="$("$GO_BIN" get "${import_path}@${version}" 2>&1)"; do
  printf 'release-prove: attempt %d/%d: go get failed: %s\n' "$attempt" "$PROVE_ATTEMPTS" "$(printf '%s' "$get_out" | tail -n 3)"
  if [ "$attempt" -ge "$PROVE_ATTEMPTS" ]; then
    fail "${import_path}@${version} did not resolve through proxy.golang.org after $PROVE_ATTEMPTS attempts"
  fi
  attempt=$((attempt + 1))
  sleep "$PROVE_SLEEP"
done
printf 'release-prove: resolved on attempt %d\n' "$attempt"

info="$("$GO_BIN" mod download -json "${import_path}@${version}" 2>&1)" || fail "go mod download failed: $info"
field() { printf '%s' "$info" | jq -r "$1 // empty"; }
printf 'release-prove: module   %s\n' "$(field .Path)@$(field .Version)"
printf 'release-prove: zip sum  %s\n' "$(field .Sum)"
printf 'release-prove: mod sum  %s\n' "$(field .GoModSum)"
origin_hash="$(field .Origin.Hash)"
printf 'release-prove: origin   %s %s %s\n' "$(field .Origin.VCS)" "$(field .Origin.URL)" "$(field .Origin.Ref)"
[ -n "$origin_hash" ] || fail "the proxy reported no origin commit for ${import_path}@${version}"
[ "$origin_hash" = "$sha" ] || fail "the proxy reports origin commit $origin_hash, not the verified commit $sha"
printf 'release-prove: origin commit %s matches the verified commit\n' "$origin_hash"

if [ -n "$EXPECT_FILE" ]; then
  dir="$(field .Dir)"
  [ -n "$dir" ] || fail "go mod download reported no directory"
  file="${dir}/${EXPECT_FILE#"${module}/"}"
  [ -f "$file" ] || fail "${EXPECT_FILE} is not in the downloaded module"
  got="$(sha256_of "$file")"
  [ "$got" = "$EXPECT_SHA256" ] || fail "${EXPECT_FILE} has SHA-256 $got, expected $EXPECT_SHA256"
  printf 'release-prove: %s sha256 %s matches\n' "$EXPECT_FILE" "$got"
fi

printf 'release-prove: consumer go.sum lines\n'
grep -F "$import_path" go.sum | sed 's/^/release-prove:   /' || fail "go.sum has no line for ${import_path}"

vet_out="$("$GO_BIN" vet ./... 2>&1)" || fail "go vet failed for the consumer: $vet_out"
run_out="$("$GO_BIN" run . 2>&1)" || fail "the consumer program failed: $run_out"
printf 'release-prove: %s\n' "$run_out"
printf 'release-prove: proof complete for %s@%s\n' "$import_path" "$version"
