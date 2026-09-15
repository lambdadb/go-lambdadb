# BulkUpsertDocsRequestBody


## Fields

| Field                                          | Type                                           | Required                                       | Description                                    |
| ---------------------------------------------- | ---------------------------------------------- | ---------------------------------------------- | ---------------------------------------------- |
| `ObjectKey`                                    | *string*                                       | :heavy_check_mark:                             | Object key uploaded based on bulk upsert info. |
| `Type`                                         | [*operations.Type](../../models/operations/type.md) | :heavy_minus_sign:                        | Optional completion field. `BulkUpsert` retains its `application/json` default when nil. The server validates the uploaded object's Content-Type instead; the upload still requires `Content-Type: application/json`. |
| `Branch`                                       | **string*                                      | :heavy_minus_sign:                             | Write target branch. Defaults to `main` when omitted. |
