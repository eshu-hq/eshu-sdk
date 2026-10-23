#!/usr/bin/env bash
# Check every repository source file, including tracked files hidden by ignore rules.
set -euo pipefail

root=$(git rev-parse --show-toplevel)
cd "$root"
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

# Git covers tracked files (including ignored ones) and new non-ignored files.
# Ripgrep independently discovers hidden source files. Their union avoids an
# ignore-rule census gap while keeping ignored build caches out of the scan.
git ls-files --cached --others --exclude-standard -z > "$tmp_dir/git-files"
rg --files --hidden -0 > "$tmp_dir/rg-files"
cat "$tmp_dir/git-files" "$tmp_dir/rg-files" | LC_ALL=C sort -z -u > "$tmp_dir/files"

bad=0
while IFS= read -r -d '' path; do
    [[ -f "$path" ]] || continue
    if ! lines=$(awk 'END { print NR }' "$path"); then
        printf 'file-cap: cannot read %q\n' "$path" >&2
        exit 1
    fi
    if (( lines >= 500 )); then
        printf 'file-cap: %q has %s lines (limit: 499)\n' "$path" "$lines" >&2
        bad=1
    fi
done < "$tmp_dir/files"
exit "$bad"
