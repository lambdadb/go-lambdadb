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
| `NumDocs` | `int64` | Yes | Total documents. |
| `DefaultBranchName` | `string` | Yes | Default writable branch, currently `main`. |
| `SnapshotRetentionInDays` | `int64` | Yes | Committed snapshot retention period. |
| `CreatedAt` | `types.UnixMilliTime` | Yes | Collection creation time as Unix epoch milliseconds. |
| `UpdatedAt` | `types.UnixMilliTime` | Yes | Collection update time as Unix epoch milliseconds. |
| `DataUpdatedAt` | `types.UnixMilliTime` | No | Data update time as Unix epoch milliseconds. |

The timestamp fields expose the embedded `time.Time`. The `GetCreatedAt`,
`GetUpdatedAt`, and `GetDataUpdatedAt` methods return `time.Time` values.
