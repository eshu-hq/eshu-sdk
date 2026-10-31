#!/usr/bin/env bash
# Shared input checks for the release scripts (release-verify.sh,
# release-tag.sh, release-prove.sh). Source it; it defines functions only.
#
# Each check returns 0 when the value is acceptable and prints a reason on
# stderr and returns 1 when it is not, so a caller decides whether that is a
# usage error or a failed check. Patterns are matched with [[ =~ ]] against the
# whole value, so a value with an embedded newline cannot pass line by line.

# The modules this repository tags. release.yml's module choice must list the
# same names; scripts/ci/test-release-verify.sh pins that.
RELEASE_ALLOWED_MODULES=(collector factschema wirecontract)

# The module is compared against the list and never interpolated into a path
# first, so "../x" and "a/b" cannot pass.
release_check_module() {
  local candidate
  for candidate in "${RELEASE_ALLOWED_MODULES[@]}"; do
    [ "$1" = "$candidate" ] && return 0
  done
  printf "module '%s' is not one of: %s\n" "$1" "${RELEASE_ALLOWED_MODULES[*]}" >&2
  return 1
}

# A plain semver tag: leading v, no build metadata, no leading zeros in the
# numeric parts or in numeric pre-release identifiers (Go rejects v0.1.0-rc.01).
# "latest", "v1", and "0.1.0" without the v all fail. The major version rule
# for a module path is release_check_major.
release_check_version() {
  local num='(0|[1-9][0-9]*)'
  local ident='(0|[1-9][0-9]*|[0-9]*[A-Za-z][0-9A-Za-z]*)'
  local semver="^v${num}\\.${num}\\.${num}(-${ident}(\\.${ident})*)?\$"
  if ! [[ $1 =~ $semver ]]; then
    printf "version '%s' is not vMAJOR.MINOR.PATCH[-prerelease]\n" "$1" >&2
    return 1
  fi
}

# Only major 0 and 1 are released for now. Go serves a module whose go.mod path
# has no /vN suffix at major 0 or 1 only, and release-prove.sh imports the
# unsuffixed path, so a v2 or later tag could be created but never proved. Until
# a /vN release path exists in prove, major >= 2 is refused outright, and the
# go.mod path must be exactly the unsuffixed one. $1 is the version, $2 the
# module path from go.mod, $3 the unsuffixed path the module is expected to have.
release_check_major() {
  local version="$1" path="$2" base="$3" major
  major="${version#v}"
  major="${major%%.*}"
  if [ "$major" != 0 ] && [ "$major" != 1 ]; then
    printf "version %s: major >= 2 needs a /vN module path release path that prove supports; not supported yet\n" "$version" >&2
    return 1
  fi
  if [ "$path" != "$base" ]; then
    printf "go.mod declares module %s, expected %s\n" "$path" "$base" >&2
    return 1
  fi
}

# A full lowercase SHA, so a branch or short name cannot resolve to a
# different commit between the checks and the tag.
release_check_sha() {
  if ! [[ $1 =~ ^[0-9a-f]{40}$ ]]; then
    printf "commit '%s' is not a full 40-hex SHA\n" "$1" >&2
    return 1
  fi
}

# The optional file-hash pair: both or neither, a clean repo-relative path
# inside the module, and a 64-hex lowercase hash. $1 module, $2 file, $3 hash.
release_check_expect() {
  local module="$1" file="$2" hash="$3"
  [ -z "$file" ] && [ -z "$hash" ] && return 0
  if [ -z "$file" ] || [ -z "$hash" ]; then
    printf "expect_file and expect_sha256 go together\n" >&2
    return 1
  fi
  case "$file" in
    "${module}/"*) ;;
    *)
      printf "expect_file '%s' must start with '%s/'\n" "$file" "$module" >&2
      return 1
      ;;
  esac
  if [[ /$file/ == *"/../"* || /$file/ == *"/./"* || /$file/ == *"//"* || $file == *$'\n'* ]]; then
    printf "expect_file '%s' must be a clean relative path\n" "$file" >&2
    return 1
  fi
  if ! [[ $hash =~ ^[0-9a-f]{64}$ ]]; then
    printf "expect_sha256 is not 64 lowercase hex\n" >&2
    return 1
  fi
}
