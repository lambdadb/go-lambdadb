# UpdateCollectionRequestBody

## Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `IndexConfigs` | map[string][components.IndexConfigsUnion](../../models/components/indexconfigsunion.md) | :heavy_minus_sign: | Nonempty full schema preserving every existing field definition, including nested fields. Only new top-level fields can be added. |
| `Description` | **string* | :heavy_minus_sign: | Nil leaves it unchanged; an empty string clears it. Maximum 255 characters. |
| `Tags` | map[string]string | :heavy_minus_sign: | Nil leaves metadata unchanged; a supplied map replaces it; an empty map clears it. Up to five entries. |
| `SnapshotRetentionInDays` | **int64* | :heavy_minus_sign: | Committed snapshot retention, from 1 through 31 days. Nil leaves it unchanged. |

Supply at least one non-nil field. The server treats omitted and null PATCH
fields identically; use nil to omit them in Go. Metadata tag keys match
`[A-Za-z0-9_.-]{1,63}`. Values must contain a non-whitespace character, have at
most 127 characters, and exclude `:`, `#`, and `,`. Validation errors are
returned by the server as `BadRequestError`.
