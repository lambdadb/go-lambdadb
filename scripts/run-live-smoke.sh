#!/usr/bin/env bash

# Runs the deployed live smoke tests in the publish workflow. It reads the
# dedicated smoke project's API key from AWS Secrets Manager with the
# credentials of the job's OIDC role, masks it in GitHub Actions logs, and passes
# it only in the environment of scripts/live-test.sh. The key never appears in
# command arguments, files, or this script's own environment.

set -euo pipefail

for name in LAMBDADB_BASE_URL LAMBDADB_PROJECT_NAME LAMBDADB_SMOKE_SECRET_ARN; do
  value="${!name:-}"
  if [[ -z "${value//[[:space:]]/}" ]]; then
    echo "Missing smoke configuration: ${name}" >&2
    exit 1
  fi
done
unset value

# Do not forward AWS output: it can contain request details or credentials.
if ! key="$(aws secretsmanager get-secret-value --region ap-northeast-2 \
  --secret-id "${LAMBDADB_SMOKE_SECRET_ARN}" --query SecretString --output text 2>/dev/null)"; then
  echo "Unable to read the dedicated smoke project key from Secrets Manager" >&2
  exit 1
fi
if [[ -z "${key}" || "${key}" == "None" || "${key}" == *$'\n'* || "${key}" == *$'\r'* ]]; then
  echo "Expected a single project API key in SecretString" >&2
  exit 1
fi

if [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
  # Workflow commands escape % in their data; the mask applies to the raw key.
  echo "::add-mask::${key//%/%25}"
fi

if ! LAMBDADB_PROJECT_API_KEY="${key}" "$(dirname "${BASH_SOURCE[0]}")/live-test.sh"; then
  echo "Deployed live smoke failed" >&2
  exit 1
fi
