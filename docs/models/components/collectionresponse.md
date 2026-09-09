# CollectionResponse

## Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `ProjectName` | `string` | Yes | Project name. |
| `CollectionName` | `string` | Yes | Collection name. |
| `IndexConfigs` | `map[string]components.IndexConfigsUnion` | Yes | Collection index configuration. |
| `Description` | `string` | Yes | Collection description. |
| `Tags` | `map[string]string` | Yes | Collection metadata tags. |
| `PartitionConfig` | `*components.PartitionConfig` | No | Partition configuration. |
| `NumPartitions` | `int64` | Yes | Total partitions including the default partition. |
| `NumDocs` | `int64` | Yes | Document count in the default `main` Branch's committed head, not a sum across Branches or a selected Ref count. |
| `DefaultBranchName` | `string` | Yes | Default writable branch, currently `main`. |
| `SnapshotRetentionInDays` | `int64` | Yes | Committed snapshot retention period. |
| `CreatedAt` | `types.UnixMilliTime` | Yes | Collection creation time as Unix epoch milliseconds. |
| `UpdatedAt` | `types.UnixMilliTime` | Yes | Collection update time as Unix epoch milliseconds. |
| `DataUpdatedAt` | `types.UnixMilliTime` | No | Last data update recorded in the default `main` Branch's committed head, in epoch milliseconds. Commits without data mutations retain the previous value; this is not necessarily the head commit time. |

The timestamp fields expose the embedded `time.Time`. The `GetCreatedAt`,
`GetUpdatedAt`, and `GetDataUpdatedAt` methods return `time.Time` values.

Before a committed head exists, `dataUpdatedAt` may be absent;
`GetDataUpdatedAt().IsZero()` is then true.
