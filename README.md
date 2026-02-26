# lambdadb

Developer-friendly & type-safe Go SDK specifically catered to leverage *lambdadb* API.

<!-- Start Summary [summary] -->
## Summary

LambdaDB API: LambdaDB Open API Spec
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [LambdaDB](#lambdadb)
  * [SDK Installation](#sdk-installation)
  * [SDK Example Usage](#sdk-example-usage)
  * [Authentication](#authentication)
  * [Configuration](#configuration)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Pagination](#pagination)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Custom HTTP Client](#custom-http-client)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

To add the SDK as a dependency to your project:
```bash
go get github.com/lambdadb/go-lambdadb
```
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example

```go
package main

import (
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"log"
)

func main() {
	ctx := context.Background()

	client := lambdadb.New(
		lambdadb.WithBaseURL("https://api.lambdadb.ai"),
		lambdadb.WithProjectName("playground"),
		lambdadb.WithAPIKey("<YOUR_PROJECT_API_KEY>"),
	)

	res, err := client.Collections.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		// handle response
	}
}

```

### Collection-scoped usage

Use `client.Collection(name)` to work with a single collection without passing the collection name on every call:

```go
	coll := client.Collection("my-collection")
	meta, err := coll.Get(ctx)
	docs, err := coll.Docs().List(ctx, nil)
	err = coll.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{Docs: myDocs})
```

<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports the following security scheme globally:

| Name            | Type   | Scheme  | Environment Variable       |
| --------------- | ------ | ------- | -------------------------- |
| `ProjectAPIKey` | apiKey | API key | `LAMBDADB_PROJECT_API_KEY` |

You can configure it using the `WithAPIKey` option when initializing the SDK client instance. For example:
```go
package main

import (
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"log"
)

func main() {
	ctx := context.Background()

	client := lambdadb.New(
		lambdadb.WithBaseURL("https://api.lambdadb.ai"),
		lambdadb.WithProjectName("playground"),
		lambdadb.WithAPIKey("<YOUR_PROJECT_API_KEY>"),
	)

	res, err := client.Collections.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		// handle response
	}
}

```
<!-- End Authentication [security] -->

<!-- Start Available Resources and Operations [operations] -->
## Available Resources and Operations

<details open>
<summary>Available methods</summary>

### Project-level: `client.Collections`

* [List](docs/sdks/collections/README.md#list) - List all collections in the project.
* [Create](docs/sdks/collections/README.md#create) - Create a collection.

### Collection-scoped: `client.Collection(name)`

Use `client.Collection("name")` for operations on a single collection (no need to pass the collection name on every call):

* **Collection**: Get, Update, Delete, Query (metadata and search).
* **Collection.Docs()**: List, Upsert, Fetch, Update, Delete, GetBulkUpsertInfo, BulkUpsert (document operations). See [docs API](docs/sdks/docs/README.md) for details.

</details>
<!-- End Available Resources and Operations [operations] -->

<!-- Start Pagination [pagination] -->
## Pagination

List endpoints return one page at a time. Use **iterators** to walk all pages without managing tokens, or **ListAll** to load every page into a single slice.

### Iterator (page-by-page)

```go
	// Documents
	iter := client.Collection("my-collection").Docs().ListIterator(ctx, &lambdadb.ListDocsOpts{Size: lambdadb.Int64(100)})
	for {
		page, err := iter.Next(ctx)
		if err != nil {
			return err
		}
		if page == nil {
			break
		}
		for _, doc := range page.Object.Docs {
			// process doc
		}
	}

	// Collections
	iter := client.Collections.ListIterator(ctx, &lambdadb.ListCollectionsOpts{Size: lambdadb.Int64(20)})
	for {
		page, err := iter.Next(ctx)
		if err != nil {
			return err
		}
		if page == nil {
			break
		}
		for _, c := range page.Object.Collections {
			// process collection
		}
	}
```

### ListAll (fetch all pages)

Use when you need the full list in memory; avoid on very large datasets.

```go
	docs, err := client.Collection("my-collection").Docs().ListAll(ctx, &lambdadb.ListDocsOpts{Size: lambdadb.Int64(100)})
	// ...
	colls, err := client.Collections.ListAll(ctx, &lambdadb.ListCollectionsOpts{Size: lambdadb.Int64(20)})
```

<!-- End Pagination [pagination] -->

<!-- Start Retries [retries] -->
## Retries

Some of the endpoints in this SDK support retries. If you use the SDK without any configuration, it will fall back to the default retry strategy provided by the API. However, the default retry strategy can be overridden on a per-operation basis, or across the entire SDK.

To change the default retry strategy for a single API call, simply provide a `retry.Config` object to the call by using the `WithRetries` option:
```go
package main

import (
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/retry"
	"log"
	"models/operations"
)

func main() {
	ctx := context.Background()

	client := lambdadb.New(
		lambdadb.WithAPIKey("<YOUR_PROJECT_API_KEY>"),
	)

	res, err := client.Collections.List(ctx, nil, operations.WithRetries(
		retry.Config{
			Strategy: "backoff",
			Backoff: &retry.BackoffStrategy{
				InitialInterval: 1,
				MaxInterval:     50,
				Exponent:        1.1,
				MaxElapsedTime:  100,
			},
			RetryConnectionErrors: false,
		}))
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		// handle response
	}
}

```

If you'd like to override the default retry strategy for all operations that support retries, you can use the `WithRetryConfig` option at SDK initialization:
```go
package main

import (
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/retry"
	"log"
)

func main() {
	ctx := context.Background()

	client := lambdadb.New(
		lambdadb.WithRetryConfig(
			retry.Config{
				Strategy: "backoff",
				Backoff: &retry.BackoffStrategy{
					InitialInterval: 1,
					MaxInterval:     50,
					Exponent:        1.1,
					MaxElapsedTime:  100,
				},
				RetryConnectionErrors: false,
			}),
		lambdadb.WithAPIKey("<YOUR_PROJECT_API_KEY>"),
	)

	res, err := client.Collections.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		// handle response
	}
}

```
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

Handling errors in this SDK should largely match your expectations. All operations return a response object or an error, they will never return both.

By Default, an API error will return `apierrors.APIError`. When custom error responses are specified for an operation, the SDK may also return their associated error. You can refer to respective *Errors* tables in SDK docs for more details on possible error types for each operation.

For example, the `List` function may return the following errors:

| Error Type                      | Status Code | Content Type     |
| ------------------------------- | ----------- | ---------------- |
| apierrors.UnauthenticatedError  | 401         | application/json |
| apierrors.ResourceNotFoundError | 404         | application/json |
| apierrors.TooManyRequestsError  | 429         | application/json |
| apierrors.InternalServerError   | 500         | application/json |
| apierrors.APIError              | 4XX, 5XX    | \*/\*            |

### Example

```go
package main

import (
	"context"
	"errors"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"log"
)

func main() {
	ctx := context.Background()

	client := lambdadb.New(
		lambdadb.WithBaseURL("https://api.lambdadb.ai"),
		lambdadb.WithProjectName("playground"),
		lambdadb.WithAPIKey("<YOUR_PROJECT_API_KEY>"),
	)

	res, err := client.Collections.List(ctx, nil)
	if err != nil {
		var authErr *apierrors.UnauthenticatedError
		if errors.As(err, &authErr) {
			log.Fatal(authErr.Error())
		}
		var notFoundErr *apierrors.ResourceNotFoundError
		if errors.As(err, &notFoundErr) {
			log.Fatal(notFoundErr.Error())
		}
		var rateLimitErr *apierrors.TooManyRequestsError
		if errors.As(err, &rateLimitErr) {
			log.Fatal(rateLimitErr.Error())
		}
		var serverErr *apierrors.InternalServerError
		if errors.As(err, &serverErr) {
			log.Fatal(serverErr.Error())
		}
		var apiErr *apierrors.APIError
		if errors.As(err, &apiErr) {
			log.Fatal(apiErr.Error())
		}
	}
}

```
<!-- End Error Handling [errors] -->

<!-- Start Configuration [config] -->
## Configuration

Configuration follows the REST API path structure. Defaults match the OpenAPI spec.

| Option | Default | Description |
| ------ | ------- | ----------- |
| `WithBaseURL(baseURL string)` | `https://api.lambdadb.ai` | API base URL (scheme + host). |
| `WithProjectName(projectName string)` | `playground` | Project name (path segment). |
| `WithAPIKey(apiKey string)` | (none) | Project API key. Can also use `LAMBDADB_PROJECT_API_KEY` env. |

The effective request base is `BaseURL + "/projects/" + ProjectName`. Example:

```go
	client := lambdadb.New(
		lambdadb.WithBaseURL("https://api.lambdadb.ai"),
		lambdadb.WithProjectName("my-project"),
		lambdadb.WithAPIKey("your-api-key"),
	)
```
<!-- End Configuration [config] -->

<!-- Start Custom HTTP Client [http-client] -->
## Custom HTTP Client

The Go SDK makes API calls that wrap an internal HTTP client. The requirements for the HTTP client are very simple. It must match this interface:

```go
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
```

The built-in `net/http` client satisfies this interface and a default client based on the built-in is provided by default. To replace this default with a client of your own, you can implement this interface yourself or provide your own client configured as desired. Here's a simple example, which adds a client with a 30 second timeout.

```go
import (
	"net/http"
	"time"

	"github.com/lambdadb/go-lambdadb"
)

var (
	httpClient = &http.Client{Timeout: 30 * time.Second}
	sdkClient  = lambdadb.New(lambdadb.WithClient(httpClient))
)
```

This can be a convenient way to configure timeouts, cookies, proxies, custom headers, and other low-level configuration.
<!-- End Custom HTTP Client [http-client] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release.
