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

## RefDetails

`RefDetails` describes a Branch or Tag. `SnapshotID` is nil for an empty Branch
head. `CreatedAt` uses `types.UnixMilliTime`.

## AliasTarget

`AliasTarget` selects a Branch or Tag when creating or retargeting an Alias.

## AliasDetails

`AliasDetails` includes the Alias ID and name, resolved target kind, name and
ID, revision, dangling status, and creation time. The returned target kind is
`BRANCH` or `TAG`.
