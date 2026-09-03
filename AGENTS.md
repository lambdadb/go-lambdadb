# Repository agent instructions

These instructions apply to the entire repository.

<!-- CLONE:ARTIFACT-REVIEW-CONTRACT:START -->
## Artifact review contract

When a response creates, changes, or relies on an artifact that needs human
review, the final response must list every such artifact with a directly
reviewable URI. Use an absolute local path for a local artifact and a direct URL
or deep link for a remote artifact. Include enough context to identify what
should be reviewed, and do not claim that an artifact was reviewed unless it
was actually opened and checked.
<!-- CLONE:ARTIFACT-REVIEW-CONTRACT:END -->

## Release policy

For any task that changes the SDK implementation, public API, API contract,
dependencies, version string, changelog, Git tags, GitHub Releases, or publish
workflow, read and follow [RELEASING.md](RELEASING.md) in full before making
changes. `RELEASING.md` is the source of truth for the release process.

The following rules are mandatory:

- Pin the exact API contract revision used for implementation. Do not treat a
  source branch, source version, or documentation change as deployment or
  general-availability evidence.
- Use an exact commit SHA for development validation. Do not create routine
  development tags.
- Use `vX.Y.Z-rc.N` tags for release candidates. The prerelease suffix in the
  Git tag, not the GitHub Release setting, keeps the version out of Go's
  `@latest` selection while a stable version exists.
- Before publishing an RC, ensure its GitHub Release is marked as a prerelease
  and is not marked as the latest GitHub Release.
- Never create a stable-looking `vX.Y.Z` tag and rely on GitHub's prerelease
  flag. Go treats that tag as a stable module version.
- Treat every pushed tag as immutable. Never move, replace, or delete a
  published tag; publish the next RC or patch version instead.
- Never create or push a tag, publish a GitHub Release, or promote an RC to a
  stable release without explicit user approval.
- Before any RC or stable release, verify that the SDK `Version`, changelog,
  Git tag, and reviewed commit agree; run all checks required by
  `RELEASING.md`; and complete the applicable environment smoke tests.
