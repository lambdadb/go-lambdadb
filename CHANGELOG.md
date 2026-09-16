# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-09-16

The first stable Data Versioning release, including the RC1–RC3 changes and the
Branch source/parent metadata update from server [PR #405](https://github.com/lambdadb/lambdadb/pull/405)
(merge commit `d1a76659884a9ed09283a0b2e2989897dc799247`). The final source
contract is [lambdadb/docs@c44180406c05b1a9043d8516e7c7f60df91fc9a7](https://github.com/lambdadb/docs/blob/c44180406c05b1a9043d8516e7c7f60df91fc9a7/reference/api/openapi.json).
Source revisions do not establish deployment or general availability of the API.

### Breaking changes

- Removed `CollectionStatus`, `SourceProjectName`, `SourceCollectionName`, and
  `SourceCollectionVersionID` from `CollectionResponse`, along with their
  getters. The current API has no direct replacements for these fields.
- Removed `SourceProjectName`, `SourceCollectionName`, `SourceDatetime`, and
  `SourceProjectAPIKey` from `CreateCollectionOptions`, along with their getters.
  Branches and Tags version data within one Collection and do not replace the
  removed cross-collection source behavior.
- Collection `CreatedAt`, `UpdatedAt`, and `DataUpdatedAt` now use
  `types.UnixMilliTime` instead of `types.UnixTime`. Their getters still return
  `time.Time`. Collection creation expects HTTP 201 and deletion HTTP 200;
  update custom transports or test servers that return the previous HTTP 202.
- For RC1/RC2 consumers, `BranchDetails` and `TagDetails` replace `RefDetails`.
  Branch Create/List return `*BranchDetails`/`[]BranchDetails`; Tag Create/List
  return `*TagDetails`/`[]TagDetails`. Read a Branch snapshot ID through
  `HeadSnapshot.SnapshotID` after a nil check. Tag `SnapshotID` is a non-null
  string. Snapshot commit times are independent of ref creation time.
- Branch creation now rejects non-Branch sources locally before sending an HTTP
  request. Replace `TagSource` with `BranchSource` or `BranchSourceAt` when
  creating a Branch. Omitting `Source` still selects `main`; Branch `AsOf` and
  Tag creation from either a Branch or Tag remain supported. The shared
  `RefSource` model and existing source helpers retain their Go signatures.

### Added

- Collection-scoped Branch, Tag, and Alias lifecycle operations; ref-scoped
  Query, Fetch, and extended List reads; Branch-scoped document mutations.
- Collection descriptions, metadata tags, default Branch, and snapshot
  retention settings.
- Branch Create/List expose nullable `ParentBranch` with `BranchID` and `Name`,
  through `ParentBranchDetails` in the root and components packages. It records
  the direct source even for an empty Branch or an ancestor snapshot selected
  through `AsOf`; it is nil for `main` or when no parent was recorded. Nil
  serializes as explicit JSON null. Parent deletion or name reuse does not
  change this historical identity. `ParentSnapshot` remains the fixed creation
  snapshot, not the previous head, and does not extend snapshot retention.
- Ref/source helpers, signed bulk-upload header forwarding, and
  `WithTransferClient` for presigned uploads and out-of-line result downloads.

### Fixed

- Paginated document reads preserve the selected Ref on every page.
- Bulk completion explicitly sends `type: "application/json"` when `Type` is
  omitted. All signed upload headers remain required independently.

### Changed

- Documented missing-ref 404 and dangling-Alias 400 errors, alias-referenced
  Branch/Tag deletion conflicts (409), Collection metadata updates, and
  committed-read semantics. Only `consistentRead: true` requires a direct Branch.
- Added contract regression tests, runnable examples, and an opt-in live smoke
  covering ref lifecycle, nullable parent/snapshot metadata, Tag-to-Tag creation,
  pagination, committed reads, and both bulk-upload paths.

## [0.4.0-rc.3] - 2026-09-15

Aligned with [lambdadb/docs@c8495bf47cd8918cfd546b4742823fd4cf3d0814](https://github.com/lambdadb/docs/blob/c8495bf47cd8918cfd546b4742823fd4cf3d0814/reference/api/openapi.json),
reviewing `b171ff0..c8495bf`. This pins the source contract; deployment and
live environment behavior require separate validation.

### Breaking changes

- Replaced `RefDetails` with `BranchDetails` and `TagDetails` in both the root
  package and `models/components`. Branch Create/List now return
  `*BranchDetails`/`[]BranchDetails`; Tag Create/List return
  `*TagDetails`/`[]TagDetails`.
- Read Branch snapshot IDs through `HeadSnapshot.SnapshotID` after checking
  `HeadSnapshot != nil`. `ParentSnapshot` is the fixed creation source, not the
  previous head. Both snapshot pointers preserve explicit JSON nulls.
- Tag `SnapshotID` is now a non-null `string`. Tags and nested Branch snapshots
  expose `SnapshotCommittedAt` as `types.UnixMilliTime`, with getters returning
  `time.Time`, independently of ref creation time.

### Changed

- Corrected schema-update guidance: new fields may be added under existing
  objects at any depth while preserving all existing fields and settings.
- Documented alias-referenced Branch/Tag deletion conflicts (409), preserving
  the existing `ResourceAlreadyExistsError` mapping and non-retry behavior.
  Updated the opt-in smoke lifecycle to release aliases before target deletion.
- Clarified that only `consistentRead: true` requires a direct Branch (including
  omitted-ref `main`); false or omitted remains valid for Tag and Alias reads.
- Clarified that bulk completion `type` is optional. The SDK retains its
  `application/json` default; upload `Content-Type: application/json` and all
  signed headers remain required independently of the completion field.

## [0.4.0-rc.2] - 2026-09-09

Aligned with `lambdadb/docs@b171ff0a408bbeb024535941b83b861d205a829f`
(`reference/api/openapi.json`), reviewing `a52ce19..b171ff0`. This pins the
source contract and does not establish deployment or general availability.

### Compatibility

No additional public API breaking changes relative to `0.4.0-rc.1`. The
breaking changes documented under RC1 still apply when upgrading from `0.3.x`.

### Fixed

- Bulk completion now explicitly sends `type: "application/json"` when callers
  omit `Type`, preserving the existing optional Go field without relying on a
  server default.

### Changed

- Clarified Collection PATCH, default-Branch statistics, committed-data reads,
  pagination, bulk upload retries, and Gateway error handling. Corrected empty
  schema and empty deletion examples. Regression tests cover PATCH clearing
  versus omission, Gateway response context, and storage 412 handling.
- Extended the live smoke test to verify metadata clearing and omitted-Type
  bulk completion, wait for committed data, and allow an eight-minute total
  budget for the additional indexing commits.
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
