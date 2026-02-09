# Collections

## Overview

### Available Operations

* [List](#list) - List all collections in an existing project.
* [Create](#create) - Create a collection.
* [Delete](#delete) - Delete an existing collection.
* [Get](#get) - Get metadata of an existing collection.
* [Update](#update) - Configure a collection.
* [Query](#query) - Search a collection with a query and return the most similar documents.

## List

List all collections in an existing project.

### Example Usage: forkedCollection

<!-- UsageSnippet language="go" operationID="listCollections" method="get" path="/collections" example="forkedCollection" -->
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

    res, err := s.Collections.List(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```
### Example Usage: normalCollection

<!-- UsageSnippet language="go" operationID="listCollections" method="get" path="/collections" example="normalCollection" -->
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

    res, err := s.Collections.List(ctx)
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
| `opts`                                                   | [][operations.Option](../../models/operations/option.md) | :heavy_minus_sign:                                       | The options for this request.                            |

### Response

**[*operations.ListCollectionsResponse](../../models/operations/listcollectionsresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## Create

Create a collection.

### Example Usage: example

<!-- UsageSnippet language="go" operationID="createCollection" method="post" path="/collections" example="example" -->
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

    res, err := s.Collections.Create(ctx, operations.CreateCollectionRequest{
        CollectionName: "<value>",
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```
### Example Usage: forkedCollection

<!-- UsageSnippet language="go" operationID="createCollection" method="post" path="/collections" example="forkedCollection" -->
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

    res, err := s.Collections.Create(ctx, operations.CreateCollectionRequest{
        CollectionName: "example-collection-name",
        SourceProjectName: lambdadb.Pointer("example-source-project-name"),
        SourceCollectionName: lambdadb.Pointer("example-source-collection-name"),
        SourceDatetime: lambdadb.Pointer("2023-10-01T12:00:00Z"),
        SourceProjectAPIKey: lambdadb.Pointer("example-source-project-api-key"),
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```
### Example Usage: normalCollection

<!-- UsageSnippet language="go" operationID="createCollection" method="post" path="/collections" example="normalCollection" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Collections.Create(ctx, operations.CreateCollectionRequest{
        CollectionName: "example-collection-name",
        IndexConfigs: map[string]components.IndexConfigsUnion{
            "example-field1": components.CreateIndexConfigsUnionText(
                components.IndexConfigsText{
                    Type: components.TypeTextText,
                    Analyzers: []components.Analyzer{
                        components.AnalyzerEnglish,
                    },
                },
            ),
            "example-field2": components.CreateIndexConfigsUnionVector(
                components.IndexConfigsVector{
                    Type: components.TypeVectorVector,
                    Dimensions: 128,
                },
            ),
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

| Parameter                                                                                | Type                                                                                     | Required                                                                                 | Description                                                                              |
| ---------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `ctx`                                                                                    | [context.Context](https://pkg.go.dev/context#Context)                                    | :heavy_check_mark:                                                                       | The context to use for the request.                                                      |
| `request`                                                                                | [operations.CreateCollectionRequest](../../models/operations/createcollectionrequest.md) | :heavy_check_mark:                                                                       | The request object to use for the request.                                               |
| `opts`                                                                                   | [][operations.Option](../../models/operations/option.md)                                 | :heavy_minus_sign:                                                                       | The options for this request.                                                            |

### Response

**[*operations.CreateCollectionResponse](../../models/operations/createcollectionresponse.md), error**

### Errors

| Error Type                           | Status Code                          | Content Type                         |
| ------------------------------------ | ------------------------------------ | ------------------------------------ |
| apierrors.BadRequestError            | 400                                  | application/json                     |
| apierrors.UnauthenticatedError       | 401                                  | application/json                     |
| apierrors.ResourceAlreadyExistsError | 409                                  | application/json                     |
| apierrors.TooManyRequestsError       | 429                                  | application/json                     |
| apierrors.InternalServerError        | 500                                  | application/json                     |
| apierrors.APIError                   | 4XX, 5XX                             | \*/\*                                |

## Delete

Delete an existing collection.

### Example Usage

<!-- UsageSnippet language="go" operationID="deleteCollection" method="delete" path="/collections/{collectionName}" example="example" -->
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

    res, err := s.Collections.Delete(ctx, "<value>")
    if err != nil {
        log.Fatal(err)
    }
    if res.MessageResponse != nil {
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

**[*operations.DeleteCollectionResponse](../../models/operations/deletecollectionresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## Get

Get metadata of an existing collection.

### Example Usage

<!-- UsageSnippet language="go" operationID="getCollection" method="get" path="/collections/{collectionName}" example="normalCollection" -->
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

    res, err := s.Collections.Get(ctx, "<value>")
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

**[*operations.GetCollectionResponse](../../models/operations/getcollectionresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## Update

Configure a collection.

### Example Usage: example

<!-- UsageSnippet language="go" operationID="updateCollection" method="patch" path="/collections/{collectionName}" example="example" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Collections.Update(ctx, "<value>", operations.UpdateCollectionRequestBody{
        IndexConfigs: map[string]components.IndexConfigsUnion{
            "example-field1": components.CreateIndexConfigsUnionText(
                components.IndexConfigsText{
                    Type: components.TypeTextText,
                    Analyzers: []components.Analyzer{
                        components.AnalyzerEnglish,
                    },
                },
            ),
            "example-field2": components.CreateIndexConfigsUnionVector(
                components.IndexConfigsVector{
                    Type: components.TypeVectorVector,
                    Dimensions: 128,
                },
            ),
            "example-field3": components.CreateIndexConfigsUnionKeyword(
                components.IndexConfigs{
                    Type: components.TypeKeyword,
                },
            ),
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
### Example Usage: normalCollection

<!-- UsageSnippet language="go" operationID="updateCollection" method="patch" path="/collections/{collectionName}" example="normalCollection" -->
```go
package main

import(
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := lambdadb.New(
        lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
    )

    res, err := s.Collections.Update(ctx, "<value>", operations.UpdateCollectionRequestBody{
        IndexConfigs: map[string]components.IndexConfigsUnion{
            "key": components.CreateIndexConfigsUnionObject(
                components.IndexConfigsObject{
                    Type: components.TypeObjectObject,
                    ObjectIndexConfigs: map[string]any{

                    },
                },
            ),
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

| Parameter                                                                                        | Type                                                                                             | Required                                                                                         | Description                                                                                      |
| ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| `ctx`                                                                                            | [context.Context](https://pkg.go.dev/context#Context)                                            | :heavy_check_mark:                                                                               | The context to use for the request.                                                              |
| `collectionName`                                                                                 | *string*                                                                                         | :heavy_check_mark:                                                                               | Collection name.                                                                                 |
| `body`                                                                                           | [operations.UpdateCollectionRequestBody](../../models/operations/updatecollectionrequestbody.md) | :heavy_check_mark:                                                                               | N/A                                                                                              |
| `opts`                                                                                           | [][operations.Option](../../models/operations/option.md)                                         | :heavy_minus_sign:                                                                               | The options for this request.                                                                    |

### Response

**[*operations.UpdateCollectionResponse](../../models/operations/updatecollectionresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.BadRequestError       | 400                             | application/json                |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |

## Query

Search a collection with a query and return the most similar documents.

### Example Usage

<!-- UsageSnippet language="go" operationID="queryCollection" method="post" path="/collections/{collectionName}/query" example="example" -->
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

    res, err := s.Collections.Query(ctx, "<value>", operations.QueryCollectionRequestBody{
        Size: lambdadb.Pointer[int64](2),
        Query: map[string]any{
            "queryString": map[string]any{
                "query": "example-field1:example-value",
            },
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

| Parameter                                                                                      | Type                                                                                           | Required                                                                                       | Description                                                                                    |
| ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `ctx`                                                                                          | [context.Context](https://pkg.go.dev/context#Context)                                          | :heavy_check_mark:                                                                             | The context to use for the request.                                                            |
| `collectionName`                                                                               | *string*                                                                                       | :heavy_check_mark:                                                                             | Collection name.                                                                               |
| `body`                                                                                         | [operations.QueryCollectionRequestBody](../../models/operations/querycollectionrequestbody.md) | :heavy_check_mark:                                                                             | N/A                                                                                            |
| `opts`                                                                                         | [][operations.Option](../../models/operations/option.md)                                       | :heavy_minus_sign:                                                                             | The options for this request.                                                                  |

### Response

**[*operations.QueryCollectionResponse](../../models/operations/querycollectionresponse.md), error**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| apierrors.BadRequestError       | 400                             | application/json                |
| apierrors.UnauthenticatedError  | 401                             | application/json                |
| apierrors.ResourceNotFoundError | 404                             | application/json                |
| apierrors.TooManyRequestsError  | 429                             | application/json                |
| apierrors.InternalServerError   | 500                             | application/json                |
| apierrors.APIError              | 4XX, 5XX                        | \*/\*                           |