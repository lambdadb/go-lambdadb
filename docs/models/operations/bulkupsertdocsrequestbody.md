# BulkUpsertDocsRequestBody


## Fields

| Field                                          | Type                                           | Required                                       | Description                                    |
| ---------------------------------------------- | ---------------------------------------------- | ---------------------------------------------- | ---------------------------------------------- |
| `ObjectKey`                                    | *string*                                       | :heavy_check_mark:                             | Object key uploaded based on bulk upsert info. |
| `Type`                                         | [*operations.Type](../../models/operations/type.md) | :heavy_minus_sign:                        | Content type used for the uploaded object. |
| `Branch`                                       | **string*                                      | :heavy_minus_sign:                             | Write target branch. Defaults to `main` when omitted. |
