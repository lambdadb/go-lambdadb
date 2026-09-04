# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Clarified and regression-tested ref read errors against
  `lambdadb/docs@a52ce19f5a1ce5ad3a30a55a5560e4591f0be9fa`: selecting a ref
  that does not exist returns `ResourceNotFoundError`, while reading through an
  Alias whose target is dangling returns `BadRequestError`. The SDK already
  decoded these HTTP statuses into the corresponding error types.

## [0.4.0-rc.1] - 2026-09-03

Implemented against the Data Versioning API contract in
`lambdadb/docs@63e07d6b2e281704aa3367fbeb94f40f519241b8` (OpenAPI `1.1.1`).
This source revision does not by itself indicate that the API is deployed.

### Breaking changes

- `CollectionResponse` no longer exposes `CollectionStatus`,
  `SourceProjectName`, `SourceCollectionName`, or
  `SourceCollectionVersionID`, and their corresponding getters have been
  removed. The current Collection response contract has no direct replacements
  for these fields.
- `CreateCollectionOptions` no longer exposes `SourceProjectName`,
  `SourceCollectionName`, `SourceDatetime`, or `SourceProjectAPIKey`, and their
  corresponding getters have been removed. Create Collections using the
  current metadata and retention options. Data Versioning Branches and Tags
  version data within an existing Collection and are not a direct replacement
  for the removed cross-collection source behavior.
- `CollectionResponse.CreatedAt`, `UpdatedAt`, and `DataUpdatedAt` now use
  `types.UnixMilliTime` instead of `types.UnixTime` because the current API
  returns Unix epoch milliseconds. Code that depends on the concrete field
  type must migrate; the `GetCreatedAt`, `GetUpdatedAt`, and `GetDataUpdatedAt`
  helpers continue to return `time.Time`.
- Collection creation now requires HTTP 201 instead of HTTP 202, and deletion
  requires HTTP 200 instead of HTTP 202, in accordance with the current API
  contract. Update test servers and custom transports that return the previous
  status codes.

### Added

- Collection-scoped Branch, Tag, and Alias lifecycle operations.
- Ref-scoped Query, Fetch, and extended List reads.
- Branch-scoped Upsert, Update, Delete, and Bulk Upsert writes.
- Collection descriptions, metadata tags, default branch, and snapshot
  retention fields.
- Signed bulk-upload header forwarding and Branch-scoped upload URL requests.
- Ref and source constructors for concise, safer Branch, Tag, and Alias usage.
- `WithTransferClient` for configuring presigned uploads and out-of-line result
  downloads independently from authenticated API requests.

### Changed

- Public Data Versioning method signatures use top-level SDK type names, and
  paginated document reads preserve their Ref across every page.

## [0.3.3] - 2026-05-28

### Added

- **List docs filters and vector inclusion**: `ListDocsOpts` now supports `IncludeVectors`, `Filter`, `PartitionFilter`, and `Fields`. The collection-scoped `Docs().List`, `ListIterator`, and `ListAll` helpers use the extended list endpoint automatically when filter, partition filter, or field selector options are set.

## [0.3.0] - 2025-03-01

### Added

- **List docs: `isDocsInline` and `docsUrl`**: When the list-docs API returns `isDocsInline=false` and `docsUrl`, the SDK fetches the document list from the presigned URL automatically (same behavior as Query and Fetch). Response body type [ListDocsResponseBody](docs/models/operations/listdocsresponsebody.md) includes `IsDocsInline` and `DocsURL`.

### Changed

- **Breaking**: `ListDocsResult.Docs` is now `[]operations.ListDocsDoc` instead of `[]map[string]any`. Each `ListDocsDoc` has `Collection` and `Doc` (document content). Use `item.Doc` for the document map. `ListAll` still returns `[]map[string]any` (document content only). This aligns List with Fetch and the API response shape.

## [0.2.1] - 2025-02-26

### Added

- **CollectionResponse timestamps**: `CreatedAt`, `UpdatedAt`, and `DataUpdatedAt` (API sends Unix epoch seconds; SDK exposes as `types.UnixTime` and `GetCreatedAt()` / `GetUpdatedAt()` / `GetDataUpdatedAt()` for `time.Time`). Documented in [CollectionResponse](docs/models/components/collectionresponse.md) and [Collections Get](docs/sdks/collections/README.md#get).

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
