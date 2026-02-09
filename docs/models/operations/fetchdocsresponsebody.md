# FetchDocsResponseBody

Fetched documents.


## Fields

| Field                                                                | Type                                                                 | Required                                                             | Description                                                          |
| -------------------------------------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------- |
| `Total`                                                              | *int64*                                                              | :heavy_check_mark:                                                   | Total number of documents returned.                                  |
| `Took`                                                               | *int64*                                                              | :heavy_check_mark:                                                   | Elapsed time in milliseconds.                                        |
| `Docs`                                                               | [][operations.FetchDocsDoc](../../models/operations/fetchdocsdoc.md) | :heavy_check_mark:                                                   | N/A                                                                  |
| `IsDocsInline`                                                       | *bool*                                                               | :heavy_check_mark:                                                   | Whether the list of documents is included.                           |
| `DocsURL`                                                            | **string*                                                            | :heavy_minus_sign:                                                   | Download URL for the list of documents.                              |