# CreateCollectionRequest

## Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `CollectionName` | *string* | :heavy_check_mark: | Collection name must be unique within a project and may contain at most 52 characters. |
| `IndexConfigs` | map[string][components.IndexConfigsUnion](../../models/components/indexconfigsunion.md) | :heavy_check_mark: | Index configurations keyed by field name. |
| `Description` | **string* | :heavy_minus_sign: | Optional collection description. |
| `Tags` | map[string]string | :heavy_minus_sign: | Collection metadata tags. The API accepts up to five entries. |
| `PartitionConfig` | [*components.PartitionConfig](../../models/components/partitionconfig.md) | :heavy_minus_sign: | Optional partition configuration. |
| `SnapshotRetentionInDays` | **int64* | :heavy_minus_sign: | Number of days to retain committed snapshots. Defaults to 30. |
