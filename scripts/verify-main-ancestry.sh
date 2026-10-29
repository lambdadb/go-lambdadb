#!/usr/bin/env bash

# Fails unless the given commit or tag (HEAD by default) is contained in the
# remote main branch. Releases are tagged only on reviewed main commits.

set -euo pipefail

target="${1:-HEAD}"
if ! commit="$(git rev-parse --verify --quiet "${target}^{commit}")"; then
  echo "unknown commit: ${target}" >&2
  exit 1
fi

git fetch --quiet --no-tags origin main
if ! git merge-base --is-ancestor "${commit}" FETCH_HEAD; then
  echo "commit ${commit} (${target}) is not contained in origin/main" >&2
  exit 1
fi

echo "commit ${commit} (${target}) is contained in origin/main"
