# ListDocsExtendedRequestBody

## Fields

| Field             | Type                                                                     | Required             | Description                                                                    |
| ----------------- | ------------------------------------------------------------------------ | -------------------- | ------------------------------------------------------------------------------ |
| `Size`            | **int64*                                                                 | :heavy_minus_sign:   | Documents per page, from 1 to 1000; defaults to 100 when omitted or null.                                     |
| `PageToken`       | **string*                                                                | :heavy_minus_sign:   | Next page token.                                                               |
| `Filter`          | map[string]*any*                                                         | :heavy_minus_sign:   | Filter applied before pagination. Omitted or null means no filter.                                              |
| `PartitionFilter` | [*components.PartitionFilter](../../models/components/partitionfilter.md) | :heavy_minus_sign:   | Restricts the request to matching partition values.                            |
| `Fields`          | [*components.FieldsSelectorUnion](../../models/components/fieldsselectorunion.md) | :heavy_minus_sign: | An object to specify a list of field names to include and/or exclude in the result. |
| `IncludeVectors`  | **bool*                                                                  | :heavy_minus_sign:   | Set to true to include vector values in the response. Defaults to false.        |
| `Ref`             | [*components.RefContext](../../models/components/versioning.md#refcontext) | :heavy_minus_sign: | Branch, tag, or alias to read. A missing ref returns 404; a dangling alias returns 400. |
