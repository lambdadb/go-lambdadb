# ListDocsRequest


## Fields

| Field                                      | Type                                       | Required                                   | Description                                |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `CollectionName`                           | *string*                                   | :heavy_check_mark:                         | Collection name.                           |
| `Size`                                     | **int64*                                   | :heavy_minus_sign:                         | Max number of documents to return at once. |
| `PageToken`                                | **string*                                  | :heavy_minus_sign:                         | Next page token.                           |
| `IncludeVectors`                           | **bool*                                    | :heavy_minus_sign:                         | Set to true to include vector values in the response. Defaults to false. |
