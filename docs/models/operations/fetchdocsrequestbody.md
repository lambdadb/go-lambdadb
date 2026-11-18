# FetchDocsRequestBody

## Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Ids` | []*string* | :heavy_check_mark: | 1–1000 nonblank document IDs. Projected bodies must also fit the server's 64 MiB per-node JSON budget; size errors return no partial result. |
| `ConsistentRead` | **bool* | :heavy_minus_sign: | True overlays eligible inherited and own pending writes on the selected Branch's committed snapshot: a direct Branch, omitted-ref `main`, or an Alias that currently targets a Branch (resolved by the server when the request is admitted). False or omitted reads only the snapshot of a Branch. Tag refs and Aliases targeting Tags always include the writes the Tag captured at creation and reject true with 400. Pending bulk imports are excluded until a background build materializes them. |
| `IncludeVectors` | **bool* | :heavy_minus_sign: | Includes vector values in the response when true. |
| `Fields` | [*components.FieldsSelectorUnion](../../models/components/fieldsselectorunion.md) | :heavy_minus_sign: | Fields to include or exclude. |
| `PartitionFilter` | [*components.PartitionFilter](../../models/components/partitionfilter.md) | :heavy_minus_sign: | Restricts the request to matching partition values. |
| `Ref` | [*components.RefContext](../../models/components/versioning.md#refcontext) | :heavy_minus_sign: | Branch, tag, or alias to read. A missing ref returns 404; a dangling alias returns 400. |
