# UpdateCollectionRequestBody

## Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `IndexConfigs` | map[string][components.IndexConfigsUnion](../../models/components/indexconfigsunion.md) | :heavy_minus_sign: | Index configurations keyed by field name. |
| `Description` | **string* | :heavy_minus_sign: | Optional collection description. |
| `Tags` | map[string]string | :heavy_minus_sign: | Collection metadata tags. The API accepts up to five entries. |
| `SnapshotRetentionInDays` | **int64* | :heavy_minus_sign: | Number of days to retain committed snapshots. |
