#!/usr/bin/env bash

# Prints the repository Variable prefix of the LambdaDB environment that the
# publish workflow's live smoke tests target, as smoke_prefix=<prefix>.
#
# - A v* tag push is a release: production.
# - A main push whose committed Version has no v<Version> tag on origin yet is a
#   release commit awaiting its tag: production, because a pushed tag is public
#   immediately and this run is the last gate before it.
# - Any other main push is a development commit: the dev environment.

set -euo pipefail

ref_type="${1:-${GITHUB_REF_TYPE:-}}"
ref_name="${2:-${GITHUB_REF_NAME:-}}"

if [[ "${ref_type}" == "tag" ]]; then
  echo "Release tag ${ref_name}: production live smoke" >&2
  echo "smoke_prefix=LAMBDADB_PROD_"
  exit 0
fi
if [[ "${ref_type}" != "branch" || "${ref_name}" != "main" ]]; then
  echo "live smoke runs only for main pushes and release tags, not ${ref_type} ${ref_name}" >&2
  exit 1
fi

version="$(sed -n 's/^const Version = "\([^"]*\)"$/\1/p' lambdadb.go)"
if [[ -z "${version}" ]]; then
  echo "could not read Version from lambdadb.go" >&2
  exit 1
fi

# Query origin rather than local tags so a shallow or stale checkout cannot
# misclassify the commit. Exit status 2 means the tag does not exist.
status=0
git ls-remote --exit-code --tags origin "refs/tags/v${version}" >/dev/null || status=$?
case "${status}" in
  0)
    echo "v${version} is already tagged; development commit: dev live smoke" >&2
    echo "smoke_prefix=LAMBDADB_DEV_"
    ;;
  2)
    echo "v${version} is not tagged yet; release commit: production live smoke" >&2
    echo "smoke_prefix=LAMBDADB_PROD_"
    ;;
  *)
    echo "could not query origin for tag v${version}" >&2
    exit 1
    ;;
esac
