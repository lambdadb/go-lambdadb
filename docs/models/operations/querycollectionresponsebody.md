# QueryCollectionResponseBody

Documents selected by query.


## Fields

| Field                                                                            | Type                                                                             | Required                                                                         | Description                                                                      |
| -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `Took`                                                                           | *int64*                                                                          | :heavy_check_mark:                                                               | Elapsed time in milliseconds.                                                    |
| `MaxScore`                                                                       | **float64*                                                                       | :heavy_minus_sign:                                                               | Maximum final returned score; absent for empty results. Numeric zero is valid.                                                                   |
| `Total`                                                                          | *int64*                                                                          | :heavy_check_mark:                                                               | Total number of documents returned.                                              |
| `Docs`                                                                           | [][operations.QueryCollectionDoc](../../models/operations/querycollectiondoc.md) | :heavy_check_mark:                                                               | List of documents.                                                               |
| `IsDocsInline`                                                                   | *bool*                                                                           | :heavy_check_mark:                                                               | Whether the list of documents is included.                                       |
| `DocsURL`                                                                        | **string*                                                                        | :heavy_minus_sign:                                                               | Optional download URL for the list of documents.                                 |
| `Rerank` | [*components.RerankResponse](../components/rerankresponse.md) | :heavy_minus_sign: | Whole-stage applied/skipped/fallback metadata; absent when unused. |

The high-level `QueryResult` preserves `Rerank`, `MaxScore`, and existing facets
when it downloads documents from `DocsURL`. See [managed reranking](../../managed-reranking.md).
