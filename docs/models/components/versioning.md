# Data Versioning models

## RefContext

`RefContext` selects a Branch, Tag, or Alias for Query, Fetch, or extended List
reads. Selecting a ref that does not exist returns `ResourceNotFoundError` (HTTP
404). Reading through an Alias whose target is dangling returns
`BadRequestError` (HTTP 400) until the Alias is retargeted to an existing Branch
or Tag.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Kind` | `components.RefKind` | Yes | `branch`, `tag`, or `alias`. |
| `Name` | `string` | Yes | Ref name. |

## RefSource

`RefSource` is shared by Branch and Tag creation. Both accept a Branch or Tag
source; aliases are not valid sources. Branch creation rejects Alias and
unknown kinds locally before sending an HTTP request.
Omitting `Source` selects `main`. `AsOf` is a Unix epoch millisecond cutoff and
is valid only with a Branch source; the server rejects it on a Tag source with
400. Existing source helper signatures are unchanged.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Kind` | `components.RefSourceKind` | Yes | `branch` or `tag`. |
| `Name` | `string` | Yes | Source ref name. |
| `AsOf` | `*int64` | No | Selects the latest retained committed snapshot at or before the cutoff for a Branch source. Inherited and pending writes are excluded; this is not an exact-time restore. |

## SnapshotDetails

| Field | Type | Description |
| --- | --- | --- |
| `SnapshotID` | `string` | Immutable snapshot identity. |
| `SnapshotCommittedAt` | `types.UnixMilliTime` | Snapshot commit time, independent of ref creation and Collection `dataUpdatedAt`. |

`GetSnapshotCommittedAt()` returns `time.Time`.

## ParentBranchDetails

| Field | Type | Description |
| --- | --- | --- |
| `BranchID` | `string` | Direct source Branch identity at creation time (`branchId` in JSON). |
| `Name` | `string` | Direct source Branch name at creation time. |

Available from both the root package and `models/components`, with nil-safe
`GetBranchID()` and `GetName()` getters.

## BranchDetails

Branch Create/Get/Update return `*BranchDetails`; List returns
`[]BranchDetails` and `ListPage` a `*ListBranchesResult`.

| Field | Type | Description |
| --- | --- | --- |
| `Name` | `string` | Branch name. |
| `ParentBranch` | `*ParentBranchDetails` | Fixed direct source Branch; nil for `main`, for a Branch created from a Tag, or when no parent was recorded. |
| `HeadSnapshot` | `*SnapshotDetails` | Current committed head; nil for an empty Branch. |
| `ParentSnapshot` | `*SnapshotDetails` | Fixed snapshot from which the Branch was created, not its previous head. |
| `NumDocs` | `int64` | Document count in the Branch's committed head; 0 for an empty Branch. |
| `DataUpdatedAt` | `*types.UnixMilliTime` | Last data update of the committed head; nil for an empty Branch. `GetDataUpdatedAt()` returns `time.Time`, zero when nil. |
| `IndexConfigs` | `map[string]components.IndexConfigsUnion` | The Branch's own index schema. |
| `CreatedAt` | `types.UnixMilliTime` | Branch creation time. `GetCreatedAt()` returns `time.Time`. |

`ParentBranch`, both snapshot fields and `DataUpdatedAt` encode nil as explicit
JSON null.
`GetParentBranch()` is nil-safe. An empty Branch can have a non-nil
`ParentBranch` while both snapshots are nil. The parent identifies the requested
source, even when `AsOf` selects a snapshot originating on an ancestor. It is
historical metadata: it does not prevent parent deletion, and recreating the
parent's name does not change its recorded identity.

Both snapshot fields retain their existing semantics. `ParentSnapshot` stays
nil for `main` and Branches created from an empty source, even after their head
advances. For a nonempty source, head and parent initially match; the head can
advance independently. Parent metadata does not extend snapshot retention.
Branch responses have no top-level `SnapshotID`.

## TagDetails

Tag Create/Get return `*TagDetails`; List returns `[]TagDetails` and `ListPage`
a `*ListTagsResult`.

| Field | Type | Description |
| --- | --- | --- |
| `Name` | `string` | Tag name. |
| `SnapshotID` | `*string` | Current published snapshot; nil before the first publication. It can change as the Tag's captured writes are published. `GetSnapshotID()` returns `""` when nil. |
| `SnapshotCommittedAt` | `*types.UnixMilliTime` | Current snapshot commit time; nil before the first publication. |
| `NumDocs` | `int64` | Document count in the current published snapshot; 0 before the first publication. |
| `DataUpdatedAt` | `*types.UnixMilliTime` | Last data update in the current snapshot; nil before the first publication. |
| `IndexConfigs` | `map[string]components.IndexConfigsUnion` | Index schema pinned by the Tag; with a source `AsOf`, the selected snapshot's schema. |
| `CreatedAt` | `types.UnixMilliTime` | Tag creation time, independent of snapshot commit time. |

`GetSnapshotCommittedAt()`, `GetDataUpdatedAt()` and `GetCreatedAt()` return
`time.Time`; the first two return the zero time when the field is nil. A Tag
can be created before its source's first snapshot when it captures pending
writes; `SnapshotID`, `SnapshotCommittedAt` and `DataUpdatedAt` then decode
from explicit JSON null to nil and encode back as null. Before contract
`docs@9a1f563`, these fields were `string` and `types.UnixMilliTime`.

## AliasTarget

`AliasTarget` selects a Branch or Tag when creating or retargeting an Alias.

## AliasDetails

`AliasDetails` includes the Alias ID and name, resolved target kind, name and
ID, revision, dangling status, `IndexConfigs`, and creation time. The returned
target kind is `BRANCH` or `TAG`. Alias Create/Get/Retarget return
`*AliasDetails`; List returns `[]AliasDetails` and `ListPage` a
`*ListAliasesResult`.

`IndexConfigs` (`map[string]components.IndexConfigsUnion`) is the schema reads
through the Alias see: the target Branch's current schema or the target Tag's
pinned schema. It is nil, and encodes as explicit JSON null, for a dangling
Alias, which `Aliases().Get` still returns.

## UpdateBranchInput

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `IndexConfigs` | `map[string]components.IndexConfigsUnion` | Yes | Nonempty complete schema of the Branch, preserving every existing field and its settings; new fields may be added at any depth. The server-managed `id` field may be sent back as returned or omitted. |

It is the whole request body of `Branches().Update`; description, tags and
snapshot retention stay on the Collection.

## RollbackBranchInput

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Tag` | `string` | Yes | Tag in the same Collection whose data and index schema the Branch is restored to. |

It is the whole request body of `Branches().Rollback`.

## RollbackBranchResponse

`Branches().Rollback` returns `*RollbackBranchResponse`, available from both
the root package and `models/components`, with nil-safe getters.

| Field | Type | Description |
| --- | --- | --- |
| `BranchID` | `string` | Identity of the rolled-back Branch, which is unchanged. |
| `HeadSnapshot` | `*SnapshotDetails` | New head committed by the rollback; nil, and encoded as explicit JSON null, when the Tag had no published snapshot, which leaves the Branch head empty. |

## ListRefsOpts

Optional parameters of `ListPage` and `ListIterator` on Branches, Tags and
Aliases. Nil uses the defaults.

| Field | Type | Description |
| --- | --- | --- |
| `Size` | `*int64` | Page size from 1 to 100; the server default is 10. Nil sends no `size`. |
| `PageToken` | `*string` | Opaque token from the previous page. Nil or empty sends no `pageToken`, which selects the first page. |
