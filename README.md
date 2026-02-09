# lambdadb

Developer-friendly & type-safe Go SDK specifically catered to leverage *lambdadb* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=lambdadb&utm_campaign=go)
[![License: MIT](https://img.shields.io/badge/LICENSE_//_MIT-3b5bdb?style=for-the-badge&labelColor=eff6ff)](https://opensource.org/licenses/MIT)


<br /><br />
> [!IMPORTANT]
> This SDK is not yet ready for production use. To complete setup please follow the steps outlined in your [workspace](https://app.speakeasy.com/org/lambdadb/go). Delete this section before > publishing to a package manager.

<!-- Start Summary [summary] -->
## Summary

LambdaDB API: LambdaDB Open API Spec
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [lambdadb](#lambdadb)
  * [SDK Installation](#sdk-installation)
  * [SDK Example Usage](#sdk-example-usage)
  * [Authentication](#authentication)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Server Selection](#server-selection)
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
<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports the following security scheme globally:

| Name            | Type   | Scheme  | Environment Variable       |
| --------------- | ------ | ------- | -------------------------- |
| `ProjectAPIKey` | apiKey | API key | `LAMBDADB_PROJECT_API_KEY` |

You can configure it using the `WithSecurity` option when initializing the SDK client instance. For example:
```go
package main

import (
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
<!-- End Authentication [security] -->

<!-- Start Available Resources and Operations [operations] -->
## Available Resources and Operations

<details open>
<summary>Available methods</summary>

### [Collections](docs/sdks/collections/README.md)

* [List](docs/sdks/collections/README.md#list) - List all collections in an existing project.
* [Create](docs/sdks/collections/README.md#create) - Create a collection.
* [Delete](docs/sdks/collections/README.md#delete) - Delete an existing collection.
* [Get](docs/sdks/collections/README.md#get) - Get metadata of an existing collection.
* [Update](docs/sdks/collections/README.md#update) - Configure a collection.
* [Query](docs/sdks/collections/README.md#query) - Search a collection with a query and return the most similar documents.

### [Docs](docs/sdks/docs/README.md)

* [List](docs/sdks/docs/README.md#list) - List documents in a collection.
* [Upsert](docs/sdks/docs/README.md#upsert) - Upsert documents into a collection. Note that the maximum supported payload size is 6MB.
* [GetBulkUpsertInfo](docs/sdks/docs/README.md#getbulkupsertinfo) - Request required info to upload documents.
* [BulkUpsert](docs/sdks/docs/README.md#bulkupsert) - Bulk upsert documents into a collection. Note that the maximum supported object size is 200MB.
* [Update](docs/sdks/docs/README.md#update) - Update documents in a collection. Note that the maximum supported payload size is 6MB.
* [Delete](docs/sdks/docs/README.md#delete) - Delete documents by document IDs or query filter from a collection.
* [Fetch](docs/sdks/docs/README.md#fetch) - Lookup and return documents by document IDs from a collection.

</details>
<!-- End Available Resources and Operations [operations] -->

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

	s := lambdadb.New(
		lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
	)

	res, err := s.Collections.List(ctx, operations.WithRetries(
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

	s := lambdadb.New(
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

	s := lambdadb.New(
		lambdadb.WithSecurity("<YOUR_PROJECT_API_KEY>"),
	)

	res, err := s.Collections.List(ctx)
	if err != nil {

		var e *apierrors.UnauthenticatedError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}

		var e *apierrors.ResourceNotFoundError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}

		var e *apierrors.TooManyRequestsError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}

		var e *apierrors.InternalServerError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}

		var e *apierrors.APIError
		if errors.As(err, &e) {
			// handle error
			log.Fatal(e.Error())
		}
	}
}

```
<!-- End Error Handling [errors] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Server Variables

The default server `https://{projectHost}` contains variables and is set to `https://api.lambdadb.com/projects/default` by default. To override default values, the following options are available when initializing the SDK client instance:

| Variable      | Option                                | Default                               | Description                |
| ------------- | ------------------------------------- | ------------------------------------- | -------------------------- |
| `projectHost` | `WithProjectHost(projectHost string)` | `"api.lambdadb.com/projects/default"` | The project URL of the API |

#### Example

```go
package main

import (
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"log"
)

func main() {
	ctx := context.Background()

	s := lambdadb.New(
		lambdadb.WithServerIndex(0),
		lambdadb.WithProjectHost("api.lambdadb.com/projects/default"),
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

### Override Server URL Per-Client

The default server can be overridden globally using the `WithServerURL(serverURL string)` option when initializing the SDK client instance. For example:
```go
package main

import (
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"log"
)

func main() {
	ctx := context.Background()

	s := lambdadb.New(
		lambdadb.WithServerURL("https://api.lambdadb.com/projects/default"),
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
<!-- End Server Selection [server] -->

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

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=lambdadb&utm_campaign=go)
