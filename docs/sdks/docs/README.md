# Docs

## Overview

### Available Operations

* [List](#list) - List documents in a collection.
* [Upsert](#upsert) - Upsert documents into a collection. Note that the maximum supported payload size is 6MB.
* [GetBulkUpsertInfo](#getbulkupsertinfo) - Request required info to upload documents.
* [BulkUpsert](#bulkupsert) - Bulk upsert documents into a collection. Note that the maximum supported object size is 200MB.
* [Update](#update) - Update documents in a collection. Note that the maximum supported payload size is 6MB.
* [Delete](#delete) - Delete documents by document IDs or query filter from a collection.
* [Fetch](#fetch) - Lookup and return documents by document IDs from a collection.

## List

List documents in a collection.

### Example Usage

<!-- UsageSnippet language="go" operationID="listDocs" method="get" path="/collections/{collectionName}/docs" example="example" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.List(ctx, "<value>", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                | Type                                                     | Required                                                 | Description                                              |
| -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- |
| `ctx`                                                    | [context.Context](https://pkg.go.dev/context#Context)    | :heavy_check_mark:                                       | The context to use for the request.                      |
| `collectionName`                                         | *string*                                                 | :heavy_check_mark:                                       | Collection name.                                         |
| `size`                                                   | **int64*                                                 | :heavy_minus_sign:                                       | Max number of documents to return at once.               |
| `pageToken`                                              | **string*                                                | :heavy_minus_sign:                                       | Next page token.                                         |
| `opts`                                                   | [][operations.Option](../../models/operations/option.md) | :heavy_minus_sign:                                       | The options for this request.                            |

### Response

**[*operations.ListDocsResponse](../../models/operations/listdocsresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.BadRequestError       | 400                             | application/json                |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## Upsert

Upsert documents into a collection. Note that the maximum supported payload size is 6MB.

### Example Usage

<!-- UsageSnippet language="go" operationID="upsertDocs" method="post" path="/collections/{collectionName}/docs/upsert" example="example" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.Upsert(ctx, "<value>", operations.UpsertDocsRequestBody{
        Docs: []map[string]any{
            map[string]any{
                "example-field1": "example-value1",
                "example-field2": []any{
                    0.1,
                    0.2,
                    0.3,
                },
            },
            map[string]any{
                "example-field1": "example-value2",
                "example-field2": []any{
                    0.4,
                    0.5,
                    0.6,
                },
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.MessageResponse != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                            | Type                                                                                 | Required                                                                             | Description                                                                          |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `ctx`                                                                                | [context.Context](https://pkg.go.dev/context#Context)                                | :heavy_check_mark:                                                                   | The context to use for the request.                                                  |
| `collectionName`                                                                     | *string*                                                                             | :heavy_check_mark:                                                                   | Collection name.                                                                     |
| `body`                                                                               | [operations.UpsertDocsRequestBody](../../models/operations/upsertdocsrequestbody.md) | :heavy_check_mark:                                                                   | N/A                                                                                  |
| `opts`                                                                               | [][operations.Option](../../models/operations/option.md)                             | :heavy_minus_sign:                                                                   | The options for this request.                                                        |

### Response

**[*operations.UpsertDocsResponse](../../models/operations/upsertdocsresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.BadRequestError       | 400                             | application/json                |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## GetBulkUpsertInfo

Request required info to upload documents.

### Example Usage

<!-- UsageSnippet language="go" operationID="getBulkUpsertDocs" method="get" path="/collections/{collectionName}/docs/bulk-upsert" example="example" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.GetBulkUpsertInfo(ctx, "<value>")
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                | Type                                                     | Required                                                 | Description                                              |
| -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- |
| `ctx`                                                    | [context.Context](https://pkg.go.dev/context#Context)    | :heavy_check_mark:                                       | The context to use for the request.                      |
| `collectionName`                                         | *string*                                                 | :heavy_check_mark:                                       | Collection name.                                         |
| `opts`                                                   | [][operations.Option](../../models/operations/option.md) | :heavy_minus_sign:                                       | The options for this request.                            |

### Response

**[*operations.GetBulkUpsertDocsResponse](../../models/operations/getbulkupsertdocsresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## BulkUpsert

Bulk upsert documents into a collection. Note that the maximum supported object size is 200MB.

### Example Usage

<!-- UsageSnippet language="go" operationID="bulkUpsertDocs" method="post" path="/collections/{collectionName}/docs/bulk-upsert" example="example" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.BulkUpsert(ctx, "<value>", operations.BulkUpsertDocsRequestBody{
        ObjectKey: "example-object-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.MessageResponse != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                                    | Type                                                                                         | Required                                                                                     | Description                                                                                  |
| -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `ctx`                                                                                        | [context.Context](https://pkg.go.dev/context#Context)                                        | :heavy_check_mark:                                                                           | The context to use for the request.                                                          |
| `collectionName`                                                                             | *string*                                                                                     | :heavy_check_mark:                                                                           | Collection name.                                                                             |
| `body`                                                                                       | [operations.BulkUpsertDocsRequestBody](../../models/operations/bulkupsertdocsrequestbody.md) | :heavy_check_mark:                                                                           | N/A                                                                                          |
| `opts`                                                                                       | [][operations.Option](../../models/operations/option.md)                                     | :heavy_minus_sign:                                                                           | The options for this request.                                                                |

### Response

**[*operations.BulkUpsertDocsResponse](../../models/operations/bulkupsertdocsresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.BadRequestError       | 400                             | application/json                |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## Update

Update documents in a collection. Note that the maximum supported payload size is 6MB.

### Example Usage

<!-- UsageSnippet language="go" operationID="updateDocs" method="post" path="/collections/{collectionName}/docs/update" example="example" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.Update(ctx, "<value>", operations.UpdateDocsRequestBody{
        Docs: []map[string]any{
            map[string]any{
                "id": "example-id1",
                "example-field1": "example-value1",
                "example-field2": []any{
                    0.1,
                    0.2,
                    0.3,
                },
            },
            map[string]any{
                "id": "example-id2",
                "example-field1": "example-value2",
                "example-field2": []any{
                    0.4,
                    0.5,
                    0.6,
                },
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.MessageResponse != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                            | Type                                                                                 | Required                                                                             | Description                                                                          |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `ctx`                                                                                | [context.Context](https://pkg.go.dev/context#Context)                                | :heavy_check_mark:                                                                   | The context to use for the request.                                                  |
| `collectionName`                                                                     | *string*                                                                             | :heavy_check_mark:                                                                   | Collection name.                                                                     |
| `body`                                                                               | [operations.UpdateDocsRequestBody](../../models/operations/updatedocsrequestbody.md) | :heavy_check_mark:                                                                   | N/A                                                                                  |
| `opts`                                                                               | [][operations.Option](../../models/operations/option.md)                             | :heavy_minus_sign:                                                                   | The options for this request.                                                        |

### Response

**[*operations.UpdateDocsResponse](../../models/operations/updatedocsresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.BadRequestError       | 400                             | application/json                |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## Delete

Delete documents by document IDs or query filter from a collection.

### Example Usage: deleteByIds

<!-- UsageSnippet language="go" operationID="deleteDocs" method="post" path="/collections/{collectionName}/docs/delete" example="deleteByIds" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.Delete(ctx, "<value>", operations.DeleteDocsRequestBody{
        Ids: []string{
            "example-doc-id-1",
            "example-doc-id-2",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.MessageResponse != nil {
        // handle response
    }
}
```
### Example Usage: deleteByQuery

<!-- UsageSnippet language="go" operationID="deleteDocs" method="post" path="/collections/{collectionName}/docs/delete" example="deleteByQuery" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.Delete(ctx, "<value>", operations.DeleteDocsRequestBody{
        Filter: map[string]any{
            "queryString": map[string]any{
                "query": "example-field1:example-value",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.MessageResponse != nil {
        // handle response
    }
}
```
### Example Usage: example

<!-- UsageSnippet language="go" operationID="deleteDocs" method="post" path="/collections/{collectionName}/docs/delete" example="example" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.Delete(ctx, "<value>", operations.DeleteDocsRequestBody{})
    if err != nil {
        log.Fatal(err)
    }
    if res.MessageResponse != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                            | Type                                                                                 | Required                                                                             | Description                                                                          |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `ctx`                                                                                | [context.Context](https://pkg.go.dev/context#Context)                                | :heavy_check_mark:                                                                   | The context to use for the request.                                                  |
| `collectionName`                                                                     | *string*                                                                             | :heavy_check_mark:                                                                   | Collection name.                                                                     |
| `body`                                                                               | [operations.DeleteDocsRequestBody](../../models/operations/deletedocsrequestbody.md) | :heavy_check_mark:                                                                   | N/A                                                                                  |
| `opts`                                                                               | [][operations.Option](../../models/operations/option.md)                             | :heavy_minus_sign:                                                                   | The options for this request.                                                        |

### Response

**[*operations.DeleteDocsResponse](../../models/operations/deletedocsresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.BadRequestError       | 400                             | application/json                |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## Fetch

Lookup and return documents by document IDs from a collection.

### Example Usage

<!-- UsageSnippet language="go" operationID="fetchDocs" method="post" path="/collections/{collectionName}/docs/fetch" example="example" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Docs.Fetch(ctx, "<value>", operations.FetchDocsRequestBody{
        Ids: []string{
            "example-doc-id-1",
            "example-doc-id-2",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                          | Type                                                                               | Required                                                                           | Description                                                                        |
| ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `ctx`                                                                              | [context.Context](https://pkg.go.dev/context#Context)                              | :heavy_check_mark:                                                                 | The context to use for the request.                                                |
| `collectionName`                                                                   | *string*                                                                           | :heavy_check_mark:                                                                 | Collection name.                                                                   |
| `body`                                                                             | [operations.FetchDocsRequestBody](../../models/operations/fetchdocsrequestbody.md) | :heavy_check_mark:                                                                 | N/A                                                                                |
| `opts`                                                                             | [][operations.Option](../../models/operations/option.md)                           | :heavy_minus_sign:                                                                 | The options for this request.                                                      |

### Response

**[*operations.FetchDocsResponse](../../models/operations/fetchdocsresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.BadRequestError       | 400                             | application/json                |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |