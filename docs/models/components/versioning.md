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

`RefSource` selects a Branch or Tag when creating a Branch or Tag. `AsOf` is a
Unix epoch millisecond cutoff and is valid only with a Branch source.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Kind` | `components.RefSourceKind` | Yes | `branch` or `tag`. |
| `Name` | `string` | Yes | Source ref name. |
| `AsOf` | `*int64` | No | Latest committed snapshot cutoff for a Branch source. |

## SnapshotDetails

| Field | Type | Description |
| --- | --- | --- |
| `SnapshotID` | `string` | Immutable snapshot identity. |
| `SnapshotCommittedAt` | `types.UnixMilliTime` | Snapshot commit time, independent of ref creation and Collection `dataUpdatedAt`. |

`GetSnapshotCommittedAt()` returns `time.Time`.

## BranchDetails

Branch Create/List return `*BranchDetails`/`[]BranchDetails`.

| Field | Type | Description |
| --- | --- | --- |
| `Name` | `string` | Branch name. |
| `HeadSnapshot` | `*SnapshotDetails` | Current committed head; nil for an empty Branch. |
| `ParentSnapshot` | `*SnapshotDetails` | Fixed snapshot from which the Branch was created, not its previous head. |
| `CreatedAt` | `types.UnixMilliTime` | Branch creation time. `GetCreatedAt()` returns `time.Time`. |

Both snapshot fields encode nil as explicit JSON null. `ParentSnapshot` stays
nil for `main` and Branches created from an empty source, even after their head
advances. For a nonempty source, head and parent initially match; the head can
advance independently. Parent metadata does not extend snapshot retention.
Branch responses have no top-level `SnapshotID`.

## TagDetails

Tag Create/List return `*TagDetails`/`[]TagDetails`.

| Field | Type | Description |
| --- | --- | --- |
| `Name` | `string` | Tag name. |
| `SnapshotID` | `string` | Immutable pinned snapshot; a Tag cannot pin an empty head. |
| `SnapshotCommittedAt` | `types.UnixMilliTime` | Pinned snapshot commit time. |
| `CreatedAt` | `types.UnixMilliTime` | Tag creation time, independent of snapshot commit time. |

`GetSnapshotCommittedAt()` and `GetCreatedAt()` return `time.Time`.

## AliasTarget

`AliasTarget` selects a Branch or Tag when creating or retargeting an Alias.

## AliasDetails

`AliasDetails` includes the Alias ID and name, resolved target kind, name and
ID, revision, dangling status, and creation time. The returned target kind is
`BRANCH` or `TAG`.
