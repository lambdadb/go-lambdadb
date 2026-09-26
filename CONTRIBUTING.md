# Contributing to This Repository

The SDK has been maintained manually since v0.2.0. Speakeasy generation was
removed; there is no checked-in OpenAPI snapshot or active code-generation
workflow. The Go models in `models/components` and `models/operations`, their
JSON helpers, and the Markdown reference under `docs/` are maintained directly.
Changes and pull requests are welcome.

## How to Report Issues

If you encounter any bugs or have suggestions for improvements, please open an issue on GitHub. When reporting an issue, please provide as much detail as possible to help us reproduce the problem. This includes:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected and actual behavior
- Any relevant logs, screenshots, or error messages
- Information about your environment (e.g., operating system, software versions)
    - For example can be collected using the `npx envinfo` command from your terminal if you have Node.js installed

## API Contract Changes

Use `reference/api/openapi.json` in [lambdadb/docs](https://github.com/lambdadb/docs)
as the upstream contract. Pin the exact reviewed source commit in the change's
Unreleased changelog entry and tests; a branch name or OpenAPI version alone is
not a reproducible reference. Record the relevant backend revision as well.
Source changes do not establish deployment or general availability.

Update the affected Go models, request/response behavior, and model or SDK
documentation together. For extensible string enums such as `Analyzer`, update
both the constants and `IsExact()` known-value list without adding closed-enum
validation. Preserve omission, explicit empty values, and unknown strings where
already supported. Avoid unrelated regeneration or formatting changes.

Add tests at the public SDK wire boundary for the changed behavior, then run:

```bash
go test ./...
go vet ./...
go build ./...
git diff --check
```

## Releasing

Maintainers must follow the development, release-candidate, and stable release
procedure in [RELEASING.md](RELEASING.md). In particular, development validation
uses an exact commit SHA, release candidates use `vX.Y.Z-rc.N`, and published
tags must never be moved or replaced.

## Contact

If you have any questions or need further assistance, please feel free to reach out by opening an issue.

Thank you for your understanding and cooperation!

The Maintainers
