# QueryCollectionResponseBody

Documents selected by query.


## Fields

| Field                                                                            | Type                                                                             | Required                                                                         | Description                                                                      |
| -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `Took`                                                                           | *int64*                                                                          | :heavy_check_mark:                                                               | Elapsed time in milliseconds.                                                    |
| `MaxScore`                                                                       | **float64*                                                                       | :heavy_minus_sign:                                                               | Maximum score.                                                                   |
| `Total`                                                                          | *int64*                                                                          | :heavy_check_mark:                                                               | Total number of documents returned.                                              |
| `Docs`                                                                           | [][operations.QueryCollectionDoc](../../models/operations/querycollectiondoc.md) | :heavy_check_mark:                                                               | List of documents.                                                               |
| `IsDocsInline`                                                                   | *bool*                                                                           | :heavy_check_mark:                                                               | Whether the list of documents is included.                                       |
| `DocsURL`                                                                        | **string*                                                                        | :heavy_minus_sign:                                                               | Optional download URL for the list of documents.                                 |