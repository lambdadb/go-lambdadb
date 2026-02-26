# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-02-26

### Added

- **Configuration**: `WithBaseURL`, `WithProjectName`, `WithAPIKey` options. Defaults follow OpenAPI spec (`https://api.lambdadb.ai`, `playground`).
- **Collection-scoped API**: `client.Collection(name)` returns a handle for a single collection. Use `coll.Get`, `coll.Update`, `coll.Delete`, `coll.Query` and `coll.Docs().List`, `coll.Docs().Upsert`, etc. without passing the collection name on every call.
- **Project-level collections**: `client.Collections` exposes only `List` and `Create`.
- **ListDocsOpts**: Optional parameters for listing documents are now passed via `*ListDocsOpts` (e.g. `List(ctx, nil)` or `List(ctx, &lambdadb.ListDocsOpts{Size: lambdadb.Int64(20)})`).
- **Public API type aliases**: `CreateCollectionOptions`, `UpdateCollectionOptions`, `QueryInput`, `UpsertDocsInput`, `UpdateDocsInput`, `DeleteDocsInput`, `FetchDocsInput`, `BulkUpsertInput` for a cleaner public API.

### Changed

- **Breaking**: Removed `WithServerURL`, `WithProjectHost`, `ServerList`, `WithServerIndex`. Use `WithBaseURL` and `WithProjectName` instead.
- **Breaking**: Removed top-level `client.Docs`. Use `client.Collection(name).Docs()` for document operations.
- **Breaking**: `Collection.Docs().List` signature is now `List(ctx, listOpts *ListDocsOpts, opts ...operations.Option)` instead of `List(ctx, size, pageToken, opts...)`.

### Removed

- Speakeasy-based code generation; SDK is now maintained manually.

## [0.1.x]

Initial releases (Speakeasy-generated). See git history for details.
