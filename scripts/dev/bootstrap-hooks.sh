#!/usr/bin/env bash
set -euo pipefail
git config --local core.hooksPath scripts/dev/hooks
printf 'Installed SDK hooks from %s\n' scripts/dev/hooks
