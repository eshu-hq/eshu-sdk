#!/usr/bin/env bash
# Fail when `go mod tidy` would change a module's go.mod/go.sum.
#
# Usage: tidy-check.sh <module-dir>
# Compares post-tidy state against a pre-tidy snapshot instead of the git
# tree, so a pull request that legitimately edits go.mod is not punished
# for its own diff.
set -euo pipefail

if [ "$#" -ne 1 ]; then
    printf 'usage: tidy-check.sh <module-dir>\n' >&2
    exit 2
fi
cd "$1"

tmp=$(mktemp -d /tmp/sdk-tidy-XXXXXX)
trap 'rm -rf "$tmp"' EXIT
cp go.mod go.sum "$tmp/"
GOWORK=off go mod tidy

status=0
diff "$tmp/go.mod" go.mod || status=$?
diff "$tmp/go.sum" go.sum || status=$?
if [ "$status" -ne 0 ]; then
    printf 'tidy-check: go mod tidy changes %s (run it and commit the result)\n' "$1" >&2
fi
exit "$status"
