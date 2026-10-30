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
cp go.mod "$tmp/"
had_sum=0
if [ -f go.sum ]; then
    cp go.sum "$tmp/"
    had_sum=1
fi
GOWORK=off go mod tidy

status=0
diff "$tmp/go.mod" go.mod || status=$?
if [ "$had_sum" -eq 1 ] && [ -f go.sum ]; then
    diff "$tmp/go.sum" go.sum || status=$?
elif [ "$had_sum" -eq 1 ] && [ ! -f go.sum ]; then
    printf 'tidy-check: go mod tidy removed %s/go.sum (run it and commit the result)\n' "$1" >&2
    status=1
elif [ "$had_sum" -eq 0 ] && [ -f go.sum ]; then
    printf 'tidy-check: go mod tidy created %s/go.sum (run it and commit the result)\n' "$1" >&2
    diff /dev/null go.sum || status=$?
fi
if [ "$status" -ne 0 ]; then
    printf 'tidy-check: go mod tidy changes %s (run it and commit the result)\n' "$1" >&2
fi
exit "$status"
