#!/usr/bin/env bash
set -euo pipefail
root=$(git rev-parse --show-toplevel)
git config --local core.hooksPath "$root/scripts/dev/hooks"
printf 'Installed SDK hooks from %s\n' "$root/scripts/dev/hooks"
