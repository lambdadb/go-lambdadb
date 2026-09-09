# FetchDocsRequestBody

## Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Ids` | []*string* | :heavy_check_mark: | Document IDs to fetch, up to 100. |
| `ConsistentRead` | **bool* | :heavy_minus_sign: | Requests a strongly consistent read. Valid only when `Ref` directly selects a branch. |
| `IncludeVectors` | **bool* | :heavy_minus_sign: | Includes vector values in the response when true. |
| `Fields` | [*components.FieldsSelectorUnion](../../models/components/fieldsselectorunion.md) | :heavy_minus_sign: | Fields to include or exclude. |
| `PartitionFilter` | [*components.PartitionFilter](../../models/components/partitionfilter.md) | :heavy_minus_sign: | Restricts the request to matching partition values. |
| `Ref` | [*components.RefContext](../../models/components/versioning.md#refcontext) | :heavy_minus_sign: | Branch, tag, or alias to read. A missing ref returns 404; a dangling alias returns 400. |
