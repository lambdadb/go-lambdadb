# ListDocsRequest


## Fields

| Field                                      | Type                                       | Required                                   | Description                                |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `CollectionName`                           | *string*                                   | :heavy_check_mark:                         | Collection name.                           |
| `Size`                                     | **int64*                                   | :heavy_minus_sign:                         | Documents per page, from 1 to 1000; defaults to 100. |
| `PageToken`                                | **string*                                  | :heavy_minus_sign:                         | Opaque position returned as `nextPageToken`. Repeat the other parameters with it; a malformed token or changed request condition returns 400. |
| `IncludeVectors`                           | **bool*                                    | :heavy_minus_sign:                         | Set to true to include vector values in the response. Defaults to false. |
