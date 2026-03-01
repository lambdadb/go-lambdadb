# ListDocsResponseBody

Documents list.


## Fields

| Field                | Type                 | Required             | Description                                            |
| -------------------- | -------------------- | -------------------- | ------------------------------------------------------ |
| `Total`              | *int64*              | :heavy_check_mark:   | N/A                                                    |
| `Docs`               | [][ListDocsDoc](./listdocsdoc.md)  | :heavy_check_mark:   | A list of documents.                                   |
| `NextPageToken`      | **string*            | :heavy_minus_sign:   | N/A                                                    |
| `IsDocsInline`       | *bool*               | :heavy_check_mark:   | Whether the list of documents is included in the response. |
| `DocsURL`            | **string*            | :heavy_minus_sign:   | Download URL for the list of documents when not inline. |