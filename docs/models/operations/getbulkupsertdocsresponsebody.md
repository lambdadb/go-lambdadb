# GetBulkUpsertDocsResponseBody

Required info to upload documents.


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `URL`                                                           | *string*                                                        | :heavy_check_mark:                                              | Presigned URL.                                                  |
| `Type`                                                          | [*operations.Type](../../models/operations/type.md)             | :heavy_minus_sign:                                              | Content type that must be specified when uploading documents.   |
| `HTTPMethod`                                                    | [*operations.HTTPMethod](../../models/operations/httpmethod.md) | :heavy_minus_sign:                                              | HTTP method that must be specified when uploading documents.    |
| `ObjectKey`                                                     | *string*                                                        | :heavy_check_mark:                                              | Object key that must be specified when uploading documents.     |
| `SizeLimitBytes`                                                | **int64*                                                        | :heavy_minus_sign:                                              | Object size limit in bytes.                                     |