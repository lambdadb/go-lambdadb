# Releasing the Go SDK

This document defines the required release process for
`github.com/lambdadb/go-lambdadb`.

Go modules are published through immutable Git tags. A GitHub Release is useful
release metadata, but it does not determine which version the Go toolchain
selects. The Go toolchain uses semantic versions from Git tags and ignores
GitHub's `prerelease` flag.

## Release channels

### Development validation

Use an untagged commit on `develop` for internal development and integration
testing. Consumers must pin the exact commit:

```bash
go get github.com/lambdadb/go-lambdadb@<commit-sha>
```

The Go toolchain converts the commit to a pseudo-version in the consumer's
`go.mod`. Prefer an exact commit over `@develop` so that test results are
reproducible.

Do not create a `dev` tag for routine development validation. A tag in the
public repository is a publicly retrievable Go module version even when its
GitHub Release is marked as a prerelease.

### Release candidate

Publish a release candidate only after the implementation is reviewed, merged
to `main`, and validated against the intended LambdaDB environment. Use a
SemVer prerelease tag:

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

## Required sequence

1. Pin the API contract revision used for the SDK implementation.
2. Implement and review the SDK changes on a feature branch.
3. Merge the reviewed changes into `develop`.
4. Validate the exact `develop` commit through a commit-SHA dependency.
5. Run smoke tests against the intended LambdaDB development or staging environment.
6. Merge the validated SDK commit into `main`.
7. Publish `vX.Y.Z-rc.1` and mark its GitHub Release as a prerelease.
8. Address feedback with a new commit and increment the RC number. Never move
   an existing tag.
9. Publish `vX.Y.Z` only after the release candidate is accepted.

## Validation checklist

Before publishing either an RC or stable tag:

- Confirm the working tree is clean and the tag target is the reviewed `main` commit.
- Confirm `Version` in `lambdadb.go` matches the tag without the leading `v`.
- Update `CHANGELOG.md` with the same version and document breaking changes
  explicitly.
- Run `go test ./...`.
- Run `go vet ./...`.
- Run `go build ./...`.
- Complete the relevant integration and smoke tests against the intended
  LambdaDB environment.
- Confirm the publish workflow marks `v*-*` tags as GitHub prereleases and does
  not make them the latest GitHub Release.

After publishing an RC, verify both selections explicitly:

```bash
go list -m -json github.com/lambdadb/go-lambdadb@latest
go list -m -json github.com/lambdadb/go-lambdadb@vX.Y.Z-rc.1
```

The first command must resolve to the latest stable version when one exists.
The second command must resolve to the intended release candidate.

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
