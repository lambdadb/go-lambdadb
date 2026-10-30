# Releasing the Go SDK

This document defines the required release process for
`github.com/lambdadb/go-lambdadb`.

Go modules are published through immutable Git tags. A GitHub Release is useful
release metadata, but it does not determine which version the Go toolchain
selects. The Go toolchain uses semantic versions from Git tags and ignores
GitHub's `prerelease` flag.

## Branches

Releases are trunk-based: `main` is the only long-lived branch.

- Branch from `main`, open a pull request into `main`, and merge it with
  **Squash and merge**. The organization ruleset `protect-main` requires a pull
  request with one approval (the `dev` team may bypass the approval) and blocks
  deletion and force pushes.
- There are no `develop`, release, or hotfix branches and no back-merges.
  Maintenance releases are cut from `main`; anything merged into `main` ships in
  the next release.
- Every RC and stable tag must point to a reviewed commit on `main`.

`Version` in `lambdadb.go` is the only committed version. Between releases it
keeps the most recently released version; a release pull request sets it to the
next RC or stable version. There is no committed development placeholder and no
automatic development publication.

## Release channels

### Development validation

Use an untagged commit on `main` for internal development and integration
testing. Consumers must pin the exact commit:

```bash
go get github.com/lambdadb/go-lambdadb@<main-commit-sha>
```

The Go toolchain converts the commit to a pseudo-version in the consumer's
`go.mod`, derived from the newest release tag reachable from that commit on
`main` (for example, `v0.7.3-0.<timestamp>-<sha>` after `v0.7.2`). Prefer an
exact commit over `@main` so that test results are reproducible.

Do not create a `dev` tag for routine development validation. A tag in the
public repository is a publicly retrievable Go module version even when its
GitHub Release is marked as a prerelease.

### Release candidate

Publish a release candidate only after the implementation is reviewed, merged
to `main`, and validated against the intended LambdaDB environment. Tag the
merge commit of the release pull request with a SemVer prerelease tag:

```text
vX.Y.Z-rc.1
vX.Y.Z-rc.2
```

Consumers must opt in by requesting the exact version:

```bash
go get github.com/lambdadb/go-lambdadb@vX.Y.Z-rc.1
```

If a stable version already exists, `@latest` continues to select the highest
stable version instead of the release candidate. This behavior comes from the
`-rc.N` suffix in the Git tag, not from GitHub's `prerelease` setting.

Never create a stable-looking tag such as `vX.Y.Z` and rely on
`prerelease: true` in GitHub. Go will treat that tag as a stable release.

### Stable release

Promote the validated release candidate by publishing the matching stable
version without a prerelease suffix:

```text
vX.Y.Z
```

Do not publish the stable tag until release-candidate feedback is resolved and
the final validation checklist passes.

For `0.6.0`, the maintainer explicitly selected a direct stable release on
2026-10-03 after exact-commit consumer validation and shared-development analyzer
and native reranking smoke tests passed. This release skips RC publication and
RC feedback steps only. All other validation, review, branch synchronization,
immutable-tag and explicit publication-approval requirements still apply.
This decision does not establish production availability or authorize publication
as part of release preparation.

For `0.7.0`, the maintainer explicitly authorized direct stable publication on
2026-10-06, without dev or RC publication. Skip RC publication and RC feedback
steps for this release only. Exact-commit consumer validation, applicable live
smoke tests, review, metadata checks, immutable tags, and branch synchronization
remain required. This authorization does not authorize server deployment or
establish production availability.

For `0.7.1`, the maintainer selected direct stable release preparation on
2026-10-07, without RC publication or RC feedback. This patch only updates
terminology and release metadata; API contract revisions remain those pinned
for `0.7.0`. No runtime feature changed, so feature-specific live smoke tests
are not applicable to this patch. Exact-commit consumer validation, SDK checks,
review, metadata checks, immutable tags, and branch synchronization remain
required. Release preparation does not authorize merging, tag publication, or
GitHub Release publication; obtain explicit approval before those actions.

For `0.7.2`, the maintainer selected direct stable release preparation on
2026-10-26, skipping RC publication and RC feedback for this release only.
Exact-commit consumer validation, applicable live smoke tests, SDK checks,
review, metadata verification, and branch synchronization remain required.
Release preparation does not authorize merging, tag publication, or GitHub
Release publication; obtain explicit approval before those actions.

## Required sequence

1. Pin the API contract revision used for the SDK implementation.
2. Implement the SDK changes on a branch from `main`, open a pull request into
   `main`, and squash-merge it after review.
3. Validate the exact `main` commit through a commit-SHA dependency.
4. Run smoke tests against the intended LambdaDB development or staging environment.
5. Open a release pull request into `main` that sets `Version` in `lambdadb.go`
   to `X.Y.Z-rc.1` and adds the matching `CHANGELOG.md` section. Run the
   [validation checklist](#validation-checklist).
6. Merge the release pull request. The resulting `main` push runs the publish
   workflow in verification mode: it validates the committed `Version` and
   changelog as the future tag, confirms the commit is on `main`, runs the SDK
   checks, and runs the [deployed live smoke tests](#deployed-live-smoke-gates)
   against production. It creates no tag or GitHub Release. Do not merge other
   pull requests until the release is tagged, or tag the release pull request's
   merge commit explicitly.
7. After verification passes and with explicit approval, tag that exact merge
   commit `vX.Y.Z-rc.1` and push the tag. The publish workflow repeats the
   checks and the production live smoke tests, confirms the tagged commit is on
   `main`, and only then creates a GitHub Release marked as a prerelease.
8. Address feedback in new pull requests into `main`, then release the next RC
   number through another release pull request. Never move an existing tag.
9. Publish `vX.Y.Z` only after the release candidate is accepted: a release
   pull request sets `Version` to `X.Y.Z` with its changelog section, and its
   verified merge commit is tagged as in steps 6 and 7.

## Validation checklist

Before publishing either an RC or stable tag:

- Confirm the working tree is clean and the tag target is the reviewed merge
  commit of the release pull request on `main`.
- Confirm `Version` in `lambdadb.go` matches the tag without the leading `v`.
- Update `CHANGELOG.md` with the same version and document breaking changes
  explicitly.
- Run `go test ./...`.
- Run `go vet ./...`.
- Run `go build ./...`.
- Complete the relevant integration and smoke tests against the intended
  LambdaDB environment.
- Confirm the production live smoke gate of the release pull request's `main`
  push passed. It is the last gate before the tag: a pushed tag is a public Go
  module version even when the tag run's checks fail.
- Confirm the publish workflow marks `v*-*` tags as GitHub prereleases and does
  not make them the latest GitHub Release.
- Confirm the publish workflow's verification run for the release pull request's
  `main` push passed.

Before pushing a tag, check out the tagged commit and validate the tag's
format, its agreement with the SDK version and changelog, and that the commit is
on `main`:

```bash
./scripts/validate-release-metadata.sh vX.Y.Z-rc.N
./scripts/verify-main-ancestry.sh vX.Y.Z-rc.N
```

The publish workflow runs the same checks and accepts only `vX.Y.Z` stable tags
or `vX.Y.Z-rc.N` release-candidate tags on `main`. RC numbering starts at 1. A
pushed tag is already a retrievable Go module version, so the workflow's checks
can only withhold the GitHub Release; run them locally before pushing.

After publishing an RC, verify both selections explicitly:

```bash
go list -m -json github.com/lambdadb/go-lambdadb@latest
go list -m -json github.com/lambdadb/go-lambdadb@vX.Y.Z-rc.1
```

The first command must resolve to the latest stable version when one exists.
The second command must resolve to the intended release candidate.

## Deployed live smoke gates

The publish workflow runs every LambdaDB live smoke test after the SDK checks:

| Event | Live target |
| :-- | :-- |
| Pull request | None. Ordinary CI has no credentials and runs no live tests. |
| `main` push whose `Version` is already tagged (development commit) | Dev environment |
| `main` push whose `Version` has no `v<Version>` tag on `origin` yet (release commit) | Production, verification only |
| `v*` tag push | Production, before the GitHub Release is created |

`scripts/select-smoke-environment.sh` makes this selection; it queries the tags
on `origin`, so a tag that exists only in a local checkout does not count. There
is no development publication: development commits are consumed as
pseudo-versions of `main` commits, and their gate is the dev live smoke result
of the `main` push.

`scripts/live-test.sh` is the live suite. It enables every opt-in flag
(`LAMBDADB_RUN_ANALYZER_SMOKE`, `LAMBDADB_RUN_BAYESIAN_SMOKE`,
`LAMBDADB_RUN_COHERE_SMOKE`, `LAMBDADB_RUN_VERSIONING_SMOKE`,
`LAMBDADB_RUN_RERANK_SMOKE`, `LAMBDADB_RUN_READ_LIMITS_SMOKE`) and runs all six
`TestIntegration*Smoke` tests one at a time (`go test -p 1`, no parallel tests); concurrent bulk writes
can exceed the project's rate limit and return HTTP 429. Missing configuration,
a failed test, a skipped test or subtest, and a test that did not run all fail
the suite. Each test creates uniquely named temporary Collections and verifies
their deletion. A new live test must be added to the script; the repository
tests fail until it is.

Every Cohere and JEV rerank query that is expected to succeed goes through
`queryIntegrationRerank` (`versioning_integration_test.go`), which sends
`onFailure: "returnOriginal"`. An `applied` result is checked as before. A
`fallback` must have the documented shape (`scoredCount` 0, no
`criteriaVersion` or `retrievalScore`, at most `size` documents) and is retried
only for reason `timeout`, `rateLimit` or `unavailable`, with at most 3 attempts
in total and a 5 s, then 10 s backoff. Any other reason or a third fallback fails
the test, so a fallback never counts as applied reranking. A request that sets
`onFailure` itself is sent once, unchanged.

In the workflow, `scripts/run-live-smoke.sh` runs the suite. The
`smoke-environment` job runs `scripts/select-smoke-environment.sh` and selects
the GitHub Environment: `dev` for development commits, `prod` for release
commits and tags. The `test` job runs in that Environment, whose secret
`LAMBDADB_PROJECT_API_KEY` holds the dedicated smoke project's API key. Actions
masks the key in logs; it is passed only in the smoke step's environment, never
in command arguments, files, or artifacts. A missing key or Variable fails the
run, which withholds the GitHub Release on a tag push.

The workflow reads these repository Variables. The `LAMBDADB_DEV_` Variables
serve development commits; the `LAMBDADB_PROD_` Variables serve release commits
and tags. A missing Variable fails the run; it never falls back to the other
environment.

| Variable | Value |
| :-- | :-- |
| `LAMBDADB_DEV_BASE_URL` | Dev environment base URL |
| `LAMBDADB_DEV_PROJECT_NAME` | Dedicated SDK smoke project in the dev environment (`go-sdk-ci`) |
| `LAMBDADB_PROD_BASE_URL` | Production base URL |
| `LAMBDADB_PROD_PROJECT_NAME` | Dedicated SDK smoke project in production (`go-sdk-ci`) |

Configure two GitHub Environments, each with the secret
`LAMBDADB_PROJECT_API_KEY` set to its project's API key:

| Environment | Deployment branches and tags | Key |
| :-- | :-- | :-- |
| `dev` | branch `main` | Dev smoke project |
| `prod` | branch `main`, tag `v*` | Production smoke project |

The deployment policy keeps the key away from other branches and tags. Use a
dedicated project where test Collections may be created and deleted, with a
`rateLimit` of at least 100 requests per second. The suite makes real
provider calls (OpenAI and Cohere embeddings, Cohere and JEV reranking), so each
environment's backend must have its Cohere Secret configured. Never use the
stack admin key. Rotate a key by issuing a new project key and replacing the
Environment secret (`gh secret set LAMBDADB_PROJECT_API_KEY --env <name>`).

For local live runs, keep the dev smoke project's settings outside the
repository in `~/.config/lambdadb/smoke/dev/go-sdk-ci.env` (mode `0600`) with
`LAMBDADB_BASE_URL`, `LAMBDADB_PROJECT_NAME`, and `LAMBDADB_PROJECT_API_KEY`,
load it into the shell without printing the key, and run the suite:

```bash
set -a; . ~/.config/lambdadb/smoke/dev/go-sdk-ci.env; set +a
./scripts/live-test.sh
```

Record the environment, SDK commit, duration, result, and cleanup evidence.
Production may lag the dev environment; report a failure caused by an
undeployed server change with its exact error rather than skipping the test.
The sections below describe each live test and the exact-commit consumer
validation that remains required before tagging a release.

## Data Versioning smoke test

Before an RC containing Data Versioning changes, load the target environment's
values without printing them:

```text
LAMBDADB_BASE_URL
LAMBDADB_PROJECT_NAME
LAMBDADB_PROJECT_API_KEY
```

Then run:

```bash
LAMBDADB_RUN_VERSIONING_SMOKE=1 \
  go test -run '^TestIntegrationDataVersioningSmoke$' -count=1 -v .
```

The test creates a uniquely named temporary Collection, exercises collection
metadata, Branch/Tag/Alias lifecycle operations, ref-scoped reads,
Branch-scoped writes, and the signed bulk-upload flow, and then deletes the
Collection. Run it only in an environment where creating and deleting that
temporary data is authorized. The API key must remain local and must not be
printed in logs or review artifacts.

## Bayesian and native embedding smoke test

Before an RC or stable release containing Bayesian or native embedding changes,
run `TestIntegrationBayesianNativeEmbeddingSmoke` against the intended authorized
LambdaDB environment after verifying the actual deployed backend revision.
Load `LAMBDADB_BASE_URL`, `LAMBDADB_PROJECT_NAME`, and `LAMBDADB_PROJECT_API_KEY`
without printing or saving credentials.

Before tagging a release, use a separate consumer module pinned to the exact SDK
`main` commit, with no local `replace` directive. Copy
`bayesian_integration_test.go` and `versioning_integration_test.go` (shared
helpers) into that module, using the same test package name in both files.
Confirm discovery and then execute:

```bash
go test -list '^TestIntegrationBayesianNativeEmbeddingSmoke$' .
LAMBDADB_RUN_BAYESIAN_SMOKE=1 \
  go test -run '^TestIntegrationBayesianNativeEmbeddingSmoke$' -count=1 -v .
```

The test must appear in the listing and pass without skipping; compilation,
"no tests to run", or a skipped test is not live validation. Record the SDK
source SHA and resolved pseudo-version, deployed backend revision evidence,
target environment, results, and cleanup evidence. The test exercises real
OpenAI document/query embeddings and JEV reranking, creates temporary
Collections, and verifies their deletion. Revoke temporary project keys and
remove temporary projects separately; never delete persistent CI resources.

## Cohere native embedding and reranking smoke test

For Cohere, multi-source or direct-vector native embedding changes, run this in
an authorized project whose deployed backend includes Cohere and has its
Cohere Secret configured, using the usual environment variables:

```bash
LAMBDADB_RUN_COHERE_SMOKE=1 \
  go test -run '^TestIntegrationCohereNativeSmoke$' -count=1 -v .
```

The test creates one uniquely named temporary Collection with a 256-dimension
`embed-v5.0-fast` field using text and image `Sources`, then re-sends the
returned schema unchanged, including the server-managed `id` field (backend
#515). It writes text, mixed, image-only and direct-vector documents,
runs `queryText`, mixed `queryInputs` and direct `queryVector` KNN, applies one
`rerank-v5.0-fast` stage, rejects explicit criteria, and checks bulk upload with
supplied vectors and rejection without them. Images are inline 32x32 PNG data
URIs, so no external URL is fetched. It deletes the Collection and verifies GET
404. It does not verify relevance quality, HTTPS image retrieval, usage
accounting, quotas or provider failure injection. For exact-commit consumer
validation, copy `cohere_integration_test.go` and `versioning_integration_test.go`
(shared helpers) into the same consumer test package.

## Read limit smoke test

For query/fetch/document-list contract changes, run the provider-free smoke in
an authorized development or staging project with the usual `LAMBDADB_BASE_URL`,
`LAMBDADB_PROJECT_NAME`, and `LAMBDADB_PROJECT_API_KEY` variables:

```bash
LAMBDADB_RUN_READ_LIMITS_SMOKE=1 \
  go test -run '^TestIntegrationReadLimitsSmoke$' -count=1 -v .
```

The test writes 1000 small documents in a uniquely named temporary Collection,
checks query/fetch and GET/POST document List at 101/1000, the List default of
100, and server rejection at 1001. It deletes the Collection and verifies GET
404. It makes no model inference calls and does not validate large-body limits,
load or production capacity. For exact-commit consumer validation, copy
`read_limits_integration_test.go` and `versioning_integration_test.go` (shared
helpers) into the same consumer test package, without a local replace directive.
Confirm the test is discovered and runs without skipping. The Bayesian smoke
also covers nested unindexed source paths, update reuse/regeneration and direct
query vectors on native fields. These live checks remain required before release;
local wire tests are not substitutes.

## Analyzer and native reranking smoke tests

Load the target environment's `LAMBDADB_BASE_URL`, `LAMBDADB_PROJECT_NAME`, and
`LAMBDADB_PROJECT_API_KEY` without printing the key. Run only in an authorized
development or staging project. Each test uses a unique temporary Collection
and deletes it afterward:

```bash
LAMBDADB_RUN_ANALYZER_SMOKE=1 \
  go test -run '^TestIntegrationAnalyzerPresetsSmoke$' -count=1 -v .

LAMBDADB_RUN_RERANK_SMOKE=1 \
  go test -run '^TestIntegrationNativeRerankingSmoke$' -count=1 -v .
```

The analyzer test checks all 49 preset names through schema read/write and a
small indexing probe. It does not establish language-quality equivalence.
The reranking test makes five provider-backed query requests (default, null
criteria, and custom 2/3/10 levels), plus legacy/null, empty-result and invalid-input
checks. It verifies final scores/order, retrieval scores, metadata and projection.
It waits for baseline query readiness before any paid stage and disables client
retries. A provider-backed query is repeated only on a transient `onFailure`
fallback, as described under "Deployed live smoke gates". It does not inject
fallback failures or verify usage accounting, quality, load or production
availability.

Before tagging a release, run these tests from a separate consumer module pinned
to the exact SDK `main` commit, with no local `replace` directive. Copy both
`rerank_integration_test.go` (the analyzer and reranking smoke-test definitions)
and `versioning_integration_test.go` (their shared helpers) into that module.
Use the same test package name in both copied files, matching the consumer module.
Before running, use `go test -list '^TestIntegration(AnalyzerPresets|NativeReranking)Smoke$' .`
and confirm that both test names appear. Record a passing, non-skipped execution
of each required smoke test; "no tests to run" or a skipped test is not validation.
Record the resolved pseudo-version, source commit, target environment,
results and cleanup evidence. A deployed server rejecting new analyzer names
is a release-validation gap; do not treat local enum/wire tests as a substitute.

## Tag safety

Treat every pushed tag as immutable. Go module proxies may cache a version after
it becomes available, so do not delete, replace, or move a published tag. If an
RC is incorrect, fix the issue and publish the next RC number. If a stable
release is incorrect, publish an appropriate patch release and use Go's
retraction mechanism when necessary.

## References

- [Go module version numbering](https://go.dev/doc/modules/version-numbers)
- [Go module version queries](https://go.dev/ref/mod#version-queries)
- [Go pseudo-versions](https://go.dev/ref/mod#pseudo-versions)
