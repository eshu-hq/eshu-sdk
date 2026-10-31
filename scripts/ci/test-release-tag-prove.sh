#!/usr/bin/env bash
# Self-test for scripts/ci/release-tag.sh and scripts/ci/release-prove.sh.
#
# Both scripts take their external tools from the environment (GH_BIN, GO_BIN),
# so this test drives them with stubs and seeds the failures the release
# workflow exists to stop: retargeting an existing tag, a tag that lands on the
# wrong object, a proxy origin that is not the verified commit, a file whose
# hash is not the expected one, and malformed inputs.
#
# Run: bash scripts/ci/test-release-tag-prove.sh
# Runs in CI in the `shape` job and locally via `make shape`.

set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TAG="$ROOT/scripts/ci/release-tag.sh"
PROVE="$ROOT/scripts/ci/release-prove.sh"
failures=0

fail() { printf 'FAIL: %s\n' "$*" >&2; failures=$((failures + 1)); }
pass() { printf 'PASS: %s\n' "$*"; }

tmp="$(mktemp -d "${TMPDIR:-/tmp}/release-tag-prove-test.XXXXXX")"
trap 'rm -rf "$tmp"' EXIT

sha=0123456789abcdef0123456789abcdef01234567
other=fedcba9876543210fedcba9876543210fedcba98

# run <expect-exit> <name> <reason-substring> <command...>
run() {
  local want_rc="$1" name="$2" want="$3"
  shift 3
  local out rc
  out="$("$@" 2>&1)"
  rc=$?
  if [ "$rc" -ne "$want_rc" ]; then
    fail "$name: expected exit $want_rc, got $rc: $out"
  elif [ -n "$want" ] && ! printf '%s' "$out" | grep -Fq -- "$want"; then
    fail "$name: exit $rc without '$want' in: $out"
  else
    pass "$name"
  fi
}

# --- release-tag.sh with a gh stub ---
cat >"$tmp/gh" <<'EOF'
#!/usr/bin/env bash
case "$*" in
  api\ repos/*/git/matching-refs/tags/*) cat "$STUB_MATCHING" ;;
  api\ -X\ POST\ repos/*/git/refs*)
    printf '%s\n' "$*" >>"$STUB_LOG"
    if [ -n "${STUB_POST_FAIL:-}" ]; then echo "gh: HTTP 422: Reference already exists" >&2; exit 1; fi
    echo '{}' ;;
  api\ repos/*/git/ref/tags/*) cat "$STUB_READBACK" ;;
  *) echo "gh stub: unexpected call: $*" >&2; exit 99 ;;
esac
EOF
chmod +x "$tmp/gh"

tag_run() { # tag_run <module> <version> <sha>
  env GH_BIN="$tmp/gh" REPO=eshu-hq/eshu-sdk STUB_MATCHING="$tmp/matching.json" STUB_READBACK="$tmp/readback.json" STUB_LOG="$tmp/post.log" \
    bash "$TAG" "$@"
}

echo '[]' >"$tmp/matching.json"
printf '{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"commit"}}\n' "$sha" >"$tmp/readback.json"
: >"$tmp/post.log"
run 0 "tag: creates the tag and confirms it" "confirmed refs/tags/factschema/v0.1.0-rc.11" tag_run factschema v0.1.0-rc.11 "$sha"
if grep -Fq -- "-f ref=refs/tags/factschema/v0.1.0-rc.11 -f sha=$sha" "$tmp/post.log"; then pass "tag: POSTs the ref at the verified SHA"; else fail "tag: POST not as expected: $(cat "$tmp/post.log")"; fi

: >"$tmp/post.log"
printf '[{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"commit"}}]\n' "$other" >"$tmp/matching.json"
run 1 "tag: refuses an existing tag (never retargets)" "already exists" tag_run factschema v0.1.0-rc.11 "$sha"
if [ ! -s "$tmp/post.log" ]; then pass "tag: no POST after the refusal"; else fail "tag: POSTed despite an existing tag"; fi

# A ref that already points at the verified commit (a rerun after a failed read-back) is accepted without a POST.
: >"$tmp/post.log"
printf '[{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"commit"}}]\n' "$sha" >"$tmp/matching.json"
printf '{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"commit"}}\n' "$sha" >"$tmp/readback.json"
run 0 "tag: an existing tag at the verified commit is accepted" "nothing to create" tag_run factschema v0.1.0-rc.11 "$sha"
if [ ! -s "$tmp/post.log" ]; then pass "tag: no POST for a tag already at the verified commit"; else fail "tag: POSTed for an existing tag"; fi
# The read-back still runs for an existing tag: a ref that moved in between is caught.
printf '{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"commit"}}\n' "$other" >"$tmp/readback.json"
run 1 "tag: an existing tag is still read back and checked" "not the verified commit" tag_run factschema v0.1.0-rc.11 "$sha"
printf '{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"commit"}}\n' "$sha" >"$tmp/readback.json"
printf '[{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"tag"}}]\n' "$sha" >"$tmp/matching.json"
run 1 "tag: an existing annotated tag is refused even at the same SHA" "already exists at" tag_run factschema v0.1.0-rc.11 "$sha"
echo '[]' >"$tmp/matching.json"
: >"$tmp/post.log"
run 1 "tag: a version with an embedded newline" "is not vMAJOR" tag_run factschema $'v0.1.0-rc.11\nevil' "$sha"
run 1 "tag: a commit with an embedded newline" "40-hex" tag_run factschema v0.1.0-rc.11 "$sha"$'\nzz'
run 1 "tag: a leading-zero pre-release identifier" "is not vMAJOR" tag_run factschema v0.1.0-rc.01 "$sha"
if [ ! -s "$tmp/post.log" ]; then pass "tag: no POST for any rejected input"; else fail "tag: POSTed for a rejected input: $(cat "$tmp/post.log")"; fi
# matching-refs matches by prefix: an existing rc.10 must not block rc.1's absence check.
printf '[{"ref":"refs/tags/factschema/v0.1.0-rc.10","object":{"sha":"%s","type":"commit"}}]\n' "$other" >"$tmp/matching.json"
printf '{"ref":"refs/tags/factschema/v0.1.0-rc.1","object":{"sha":"%s","type":"commit"}}\n' "$sha" >"$tmp/readback.json"
run 0 "tag: a longer tag sharing the prefix is not a clash" "confirmed" tag_run factschema v0.1.0-rc.1 "$sha"

echo '[]' >"$tmp/matching.json"
printf '{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"commit"}}\n' "$other" >"$tmp/readback.json"
run 1 "tag: read-back at another commit fails" "not the verified commit" tag_run factschema v0.1.0-rc.11 "$sha"
printf '{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"tag"}}\n' "$sha" >"$tmp/readback.json"
run 1 "tag: an annotated tag object fails" "lightweight" tag_run factschema v0.1.0-rc.11 "$sha"
printf '{"ref":"refs/tags/factschema/v0.1.0-rc.11","object":{"sha":"%s","type":"commit"}}\n' "$sha" >"$tmp/readback.json"
run 1 "tag: a rejected POST fails" "creating refs/tags/factschema/v0.1.0-rc.11 failed" env STUB_POST_FAIL=1 GH_BIN="$tmp/gh" REPO=eshu-hq/eshu-sdk STUB_MATCHING="$tmp/matching.json" STUB_READBACK="$tmp/readback.json" STUB_LOG="$tmp/post.log" bash "$TAG" factschema v0.1.0-rc.11 "$sha"
run 1 "tag: bad module" "is not one of" tag_run evil v0.1.0-rc.11 "$sha"
run 1 "tag: bad version" "is not vMAJOR" tag_run factschema latest "$sha"
run 1 "tag: short commit" "40-hex" tag_run factschema v0.1.0-rc.11 "${sha:0:10}"
run 2 "tag: wrong argument count" "usage" tag_run factschema v0.1.0-rc.11
run 1 "tag: no repository" "REPO" env -u GITHUB_REPOSITORY GH_BIN="$tmp/gh" REPO= bash "$TAG" factschema v0.1.0-rc.11 "$sha"

# --- release-prove.sh with a go stub ---
mkdir "$tmp/moddir"
mkdir -p "$tmp/moddir/fixturepack/payloads"
printf 'golden bytes\n' >"$tmp/moddir/fixturepack/payloads/g.json"
if command -v sha256sum >/dev/null 2>&1; then good_hash="$(sha256sum "$tmp/moddir/fixturepack/payloads/g.json" | cut -d' ' -f1)"; else good_hash="$(shasum -a 256 "$tmp/moddir/fixturepack/payloads/g.json" | cut -d' ' -f1)"; fi
cat >"$tmp/go" <<'EOF'
#!/usr/bin/env bash
# Records the clean-consumer environment on `get`, so the test can assert it.
case "$1 $2" in
  "mod init") printf 'module example.com/release-proof\n' >go.mod ;;
  "get "*)
    count="$(cat "$STUB_COUNT" 2>/dev/null || echo 0)"
    echo $((count + 1)) >"$STUB_COUNT"
    printf 'GOWORK=%s GOPROXY=%s GOSUMDB=%s GONOSUMDB=%s GOFLAGS=%s GOPRIVATE=%s GOINSECURE=%s GONOPROXY=%s GOENV=%s GOMODCACHE=%s\n' \
      "${GOWORK-unset}" "${GOPROXY-unset}" "${GOSUMDB-unset}" "${GONOSUMDB-unset}" "${GOFLAGS-unset}" "${GOPRIVATE-unset}" "${GOINSECURE-unset}" "${GONOPROXY-unset}" "${GOENV-unset}" "${GOMODCACHE-unset}" >>"$STUB_ENVLOG"
    if [ "$count" -lt "${STUB_GET_FAILS:-0}" ]; then echo "go: module not found yet" >&2; exit 1; fi
    path="${2%@*}"
    printf '%s v0.1.0-rc.11 h1:zip\n%s v0.1.0-rc.11/go.mod h1:mod\n' "$path" "$path" >go.sum ;;
  "mod download")
    origin=',"Origin":{"VCS":"git","URL":"https://github.com/eshu-hq/eshu-sdk","Ref":"refs/tags/factschema/v0.1.0-rc.11","Hash":"0123456789abcdef0123456789abcdef01234567"}'
    printf '{"Path":"github.com/eshu-hq/eshu-sdk/factschema","Version":"v0.1.0-rc.11","Sum":"h1:zip","GoModSum":"h1:mod","Dir":"%s"%s}\n' \
      "$STUB_DIR" "${STUB_ORIGIN_JSON-$origin}" ;;
  "vet ./...") [ "${STUB_VET_RC:-0}" = 0 ] || echo "vet complaint" >&2; exit "${STUB_VET_RC:-0}" ;;
  "run .") echo "release-proof: imported github.com/eshu-hq/eshu-sdk/factschema" ;;
  *) echo "go stub: unexpected call: $*" >&2; exit 99 ;;
esac
EOF
chmod +x "$tmp/go"

prove_run() { # prove_run <extra env...> -- <args...>
  env GO_BIN="$tmp/go" STUB_DIR="$tmp/moddir" STUB_COUNT="$tmp/count" STUB_ENVLOG="$tmp/env.log" PROVE_SLEEP=0 PROVE_ATTEMPTS=4 \
    GONOSUMDB=poison GOFLAGS=-mod=mod GOPRIVATE=poison GOINSECURE=poison GONOPROXY=poison "$@"
}

rm -f "$tmp/count" "$tmp/env.log"
run 0 "prove: resolves, matches the origin, builds, runs" "proof complete" prove_run bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
env_line="$(cat "$tmp/env.log" 2>/dev/null)"
clean=1
# Whole-token matches, so GOPROXY=https://proxy.golang.org,direct cannot pass.
for want in GOWORK=off GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org GOENV=off GONOSUMDB=unset GOFLAGS=unset GOPRIVATE=unset GOINSECURE=unset GONOPROXY=unset; do
  printf '%s\n' "$env_line" | tr ' ' '\n' | grep -Fxq -- "$want" || { clean=0; fail "prove: consumer environment lacks exactly $want: $env_line"; }
done
printf '%s' "$env_line" | grep -Eq 'GOMODCACHE=.*/release-prove\.[A-Za-z0-9]+/modcache' || { clean=0; fail "prove: consumer does not use its own module cache: $env_line"; }
if [ "$clean" -eq 1 ]; then pass "prove: runs in a clean consumer environment (workspace off, exact public proxy and sumdb, no go env file, poison variables cleared, own cache)"; fi
out="$(prove_run bash "$PROVE" factschema v0.1.0-rc.11 "$sha" 2>&1)"
for needle in "release-proof: imported github.com/eshu-hq/eshu-sdk/factschema" "zip sum  h1:zip" "mod sum  h1:mod" "origin   git https://github.com/eshu-hq/eshu-sdk refs/tags/factschema/v0.1.0-rc.11" "consumer go.sum lines"; do
  if printf '%s' "$out" | grep -Fq "$needle"; then pass "prove: log records '$needle'"; else fail "prove: log lacks '$needle'"; fi
done

rm -f "$tmp/count"
run 0 "prove: retries while the proxy lags" "resolved on attempt 3" prove_run env STUB_GET_FAILS=2 bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
rm -f "$tmp/count"
run 1 "prove: gives up when the proxy never serves the tag" "did not resolve" prove_run env STUB_GET_FAILS=99 bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
rm -f "$tmp/count"
run 0 "prove: the last allowed attempt can succeed" "resolved on attempt 4" prove_run env STUB_GET_FAILS=3 bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
rm -f "$tmp/count"
run 1 "prove: one failure past the bound gives up" "after 4 attempts" prove_run env STUB_GET_FAILS=4 bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
if [ "$(cat "$tmp/count")" = "4" ]; then pass "prove: exactly 4 attempts were made"; else fail "prove: made $(cat "$tmp/count") attempts, want 4"; fi

run 1 "prove: origin is another commit" "not the verified commit" prove_run bash "$PROVE" factschema v0.1.0-rc.11 "$other"
rm -f "$tmp/count"
run 1 "prove: the proxy reports no origin" "reported no origin" prove_run env STUB_ORIGIN_JSON= bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
rm -f "$tmp/count"
run 1 "prove: go vet fails and says why" "go vet failed for the consumer: vet complaint" prove_run env STUB_VET_RC=1 bash "$PROVE" factschema v0.1.0-rc.11 "$sha"

rm -f "$tmp/count"
run 0 "prove: the expected file hash matches" "sha256 $good_hash matches" prove_run env EXPECT_FILE=factschema/fixturepack/payloads/g.json EXPECT_SHA256="$good_hash" bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
rm -f "$tmp/count"
run 1 "prove: a different file hash fails" "expected" prove_run env EXPECT_FILE=factschema/fixturepack/payloads/g.json EXPECT_SHA256="$(printf '%064d' 0)" bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
rm -f "$tmp/count"
run 1 "prove: a missing file fails" "is not in the downloaded module" prove_run env EXPECT_FILE=factschema/nope.json EXPECT_SHA256="$good_hash" bash "$PROVE" factschema v0.1.0-rc.11 "$sha"

run 2 "prove: file without a hash" "go together" prove_run env EXPECT_FILE=factschema/fixturepack/payloads/g.json bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
run 2 "prove: hash without a file" "go together" prove_run env EXPECT_SHA256="$good_hash" bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
run 2 "prove: file outside the module" "must start with" prove_run env EXPECT_FILE=collector/go.mod EXPECT_SHA256="$good_hash" bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
run 2 "prove: path traversal" "clean relative path" prove_run env EXPECT_FILE=factschema/../collector/go.mod EXPECT_SHA256="$good_hash" bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
run 2 "prove: uppercase hash" "64 lowercase hex" prove_run env EXPECT_FILE=factschema/a.json EXPECT_SHA256="$(printf '%s' "$good_hash" | tr a-f A-F)" bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
run 2 "prove: 65-hex hash" "64 lowercase hex" prove_run env EXPECT_FILE=factschema/a.json EXPECT_SHA256="${good_hash}0" bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
run 2 "prove: hash with a trailing newline" "64 lowercase hex" prove_run env EXPECT_FILE=factschema/a.json EXPECT_SHA256="$good_hash"$'\n' bash "$PROVE" factschema v0.1.0-rc.11 "$sha"
run 2 "prove: bad module" "is not one of" prove_run bash "$PROVE" evil v0.1.0-rc.11 "$sha"
run 2 "prove: bad version" "is not vMAJOR" prove_run bash "$PROVE" factschema latest "$sha"
run 2 "prove: bad commit" "40-hex" prove_run bash "$PROVE" factschema v0.1.0-rc.11 abc
run 2 "prove: wrong argument count" "usage" prove_run bash "$PROVE" factschema v0.1.0-rc.11

if [ "$failures" -ne 0 ]; then
  printf 'test-release-tag-prove: %d failure(s)\n' "$failures" >&2
  exit 1
fi
printf 'test-release-tag-prove: ok\n'
