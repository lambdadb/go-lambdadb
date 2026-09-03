#!/usr/bin/env bash

set -euo pipefail

release_tag="${1:-${GITHUB_REF_NAME:-}}"
if [[ -z "${release_tag}" ]]; then
  echo "release tag is required" >&2
  exit 1
fi

if [[ ! "${release_tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[1-9][0-9]*)?$ ]]; then
  echo "invalid release tag: ${release_tag}" >&2
  echo "expected vX.Y.Z or vX.Y.Z-rc.N where N starts at 1" >&2
  exit 1
fi

release_version="${release_tag#v}"
sdk_version="$(sed -n 's/^const Version = "\([^"]*\)"$/\1/p' lambdadb.go)"
if [[ -z "${sdk_version}" ]]; then
  echo "could not read Version from lambdadb.go" >&2
  exit 1
fi
if [[ "${sdk_version}" != "${release_version}" ]]; then
  echo "release tag ${release_tag} does not match SDK Version ${sdk_version}" >&2
  exit 1
fi

if ! grep -Fq "## [${release_version}]" CHANGELOG.md; then
  echo "CHANGELOG.md has no release section for ${release_version}" >&2
  exit 1
fi

if [[ "${release_tag}" == *-rc.* ]]; then
  echo "prerelease=true"
  echo "make_latest=false"
else
  echo "prerelease=false"
  echo "make_latest=true"
fi
