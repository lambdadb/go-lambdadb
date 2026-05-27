# ListDocsExtendedRequestBody

## Fields

| Field             | Type                                                                     | Required             | Description                                                                    |
| ----------------- | ------------------------------------------------------------------------ | -------------------- | ------------------------------------------------------------------------------ |
| `Size`            | **int64*                                                                 | :heavy_minus_sign:   | Max number of documents to return at once.                                     |
| `PageToken`       | **string*                                                                | :heavy_minus_sign:   | Next page token.                                                               |
| `Filter`          | map[string]*any*                                                         | :heavy_minus_sign:   | Filter applied before pagination.                                              |
| `PartitionFilter` | [*components.PartitionFilter](../../models/components/partitionfilter.md) | :heavy_minus_sign:   | Restricts the request to matching partition values.                            |
| `Fields`          | [*components.FieldsSelectorUnion](../../models/components/fieldsselectorunion.md) | :heavy_minus_sign: | An object to specify a list of field names to include and/or exclude in the result. |
| `IncludeVectors`  | **bool*                                                                  | :heavy_minus_sign:   | Set to true to include vector values in the response. Defaults to false.        |
