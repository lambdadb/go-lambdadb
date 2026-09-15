# QueryCollectionRequestBody

## Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Size` | **int64* | :heavy_minus_sign: | Number of documents to return, up to 100. |
| `Query` | map[string]*any* | :heavy_check_mark: | Query object. For managed embeddings use `knn.queryText`; for unmanaged vector fields use `knn.queryVector`. |
| `ConsistentRead` | **bool* | :heavy_minus_sign: | True overlays eligible pending writes only for a direct Branch or omitted-ref `main`. Tag and Alias refs reject true, including Aliases targeting Branches. False or omitted reads committed data. |
| `IncludeVectors` | **bool* | :heavy_minus_sign: | Includes vector values in the response when true. |
| `Sort` | []map[string]*any* | :heavy_minus_sign: | Field name and sort direction pairs. |
| `Fields` | [*components.FieldsSelectorUnion](../../models/components/fieldsselectorunion.md) | :heavy_minus_sign: | Fields to include or exclude. |
| `PartitionFilter` | [*components.PartitionFilter](../../models/components/partitionfilter.md) | :heavy_minus_sign: | Restricts the request to matching partition values. |
| `Ref` | [*components.RefContext](../../models/components/versioning.md#refcontext) | :heavy_minus_sign: | Branch, tag, or alias to read. A missing ref returns 404; a dangling alias returns 400. |
