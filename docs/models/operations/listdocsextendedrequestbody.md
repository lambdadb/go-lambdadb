# ListDocsExtendedRequestBody

## Fields

| Field             | Type                                                                     | Required             | Description                                                                    |
| ----------------- | ------------------------------------------------------------------------ | -------------------- | ------------------------------------------------------------------------------ |
| `Size`            | **int64*                                                                 | :heavy_minus_sign:   | Documents per page, from 1 to 1000; defaults to 100 when omitted or null.                                     |
| `PageToken`       | **string*                                                                | :heavy_minus_sign:   | Opaque position in the requested order. Repeat the original body with it; a malformed token, incompatible sort value, or changed request condition returns 400. |
| `Sort`            | []map[string]*any*                                                       | :heavy_minus_sign:   | 1 to 4 objects, each mapping one keyword, long, double, or datetime field to `"asc"` or `"desc"`. Ascending document ID breaks ties and is the default order. Nil omits it. |
| `Filter`          | map[string]*any*                                                         | :heavy_minus_sign:   | Filter applied before pagination. Omitted or null means no filter.                                              |
| `PartitionFilter` | [*components.PartitionFilter](../../models/components/partitionfilter.md) | :heavy_minus_sign:   | Restricts the request to matching partition values.                            |
| `Fields`          | [*components.FieldsSelectorUnion](../../models/components/fieldsselectorunion.md) | :heavy_minus_sign: | An object to specify a list of field names to include and/or exclude in the result. |
| `IncludeVectors`  | **bool*                                                                  | :heavy_minus_sign:   | Set to true to include vector values in the response. Defaults to false.        |
| `Ref`             | [*components.RefContext](../../models/components/versioning.md#refcontext) | :heavy_minus_sign: | Branch, tag, or alias to read. A missing ref returns 404; a dangling alias returns 400. |
| `ConsistentRead`  | **bool*                                                                  | :heavy_minus_sign:   | Include pending ordinary writes of the branch and its inherited history. Requires a branch or an alias targeting a branch; tags and tag aliases reject true with 400. Nil omits it (server default false). |
