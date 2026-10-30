#!/usr/bin/env bash

# Runs the deployed live smoke tests in the publish workflow. The job's GitHub
# Environment supplies the dedicated smoke project's API key in
# LAMBDADB_PROJECT_API_KEY; Actions masks it in logs. The key never appears in
# command arguments or files.

set -euo pipefail

for name in LAMBDADB_BASE_URL LAMBDADB_PROJECT_NAME LAMBDADB_PROJECT_API_KEY; do
  value="${!name:-}"
  if [[ -z "${value//[[:space:]]/}" ]]; then
    echo "Missing smoke configuration: ${name}" >&2
    exit 1
  fi
done
unset value

if [[ "${LAMBDADB_PROJECT_API_KEY}" == *$'\n'* || "${LAMBDADB_PROJECT_API_KEY}" == *$'\r'* ]]; then
  echo "Expected a single project API key" >&2
  exit 1
fi

if ! "$(dirname "${BASH_SOURCE[0]}")/live-test.sh"; then
  echo "Deployed live smoke failed" >&2
  exit 1
fi
