#!/usr/bin/env bash

# Runs every deployed LambdaDB live smoke test one at a time against the
# environment in LAMBDADB_BASE_URL, LAMBDADB_PROJECT_NAME, and
# LAMBDADB_PROJECT_API_KEY. Every test must run and pass: missing configuration,
# a skipped test, or a test that did not run fails the suite.

set -euo pipefail

live_tests=(
  TestIntegrationAnalyzerPresetsSmoke
  TestIntegrationBayesianNativeEmbeddingSmoke
  TestIntegrationDataVersioningSmoke
  TestIntegrationNativeRerankingSmoke
  TestIntegrationReadLimitsSmoke
)

for name in LAMBDADB_BASE_URL LAMBDADB_PROJECT_NAME LAMBDADB_PROJECT_API_KEY; do
  value="${!name:-}"
  if [[ -z "${value//[[:space:]]/}" ]]; then
    echo "Missing live test configuration: ${name}" >&2
    exit 1
  fi
done
unset value

export LAMBDADB_RUN_ANALYZER_SMOKE=1
export LAMBDADB_RUN_BAYESIAN_SMOKE=1
export LAMBDADB_RUN_VERSIONING_SMOKE=1
export LAMBDADB_RUN_RERANK_SMOKE=1
export LAMBDADB_RUN_READ_LIMITS_SMOKE=1

pattern="^($(IFS='|'; echo "${live_tests[*]}"))\$"
cd "$(dirname "${BASH_SOURCE[0]}")/.."

# -p 1 and no t.Parallel keep the tests sequential; concurrent bulk writes can
# exceed the project's rate limit. Output is streamed and kept in memory only.
status=0
output="$(go test -p 1 -count=1 -timeout 45m -v -run "${pattern}" . 2>&1 | tee /dev/stderr)" || status=$?

if grep -Eq '^[[:space:]]*--- SKIP: ' <<<"${output}"; then
  echo "A live smoke test was skipped; skipped tests are not validation" >&2
  status=1
fi
for test in "${live_tests[@]}"; do
  if ! grep -Eq "^--- PASS: ${test} " <<<"${output}"; then
    echo "Live smoke test ${test} did not pass" >&2
    status=1
  fi
done
exit "${status}"
