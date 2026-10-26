# Options

API methods accept optional `operations.Option` values. Client-level base URL,
project name and API key are configured with `lambdadb.WithBaseURL`,
`WithProjectName` and `WithAPIKey`. Options below affect API calls; presigned
uploads and document downloads use the separate transfer client.

## WithRetries

`WithRetries` overrides the retry policy for one API operation. All three
`retry.BackoffStrategy` interval fields are integer **milliseconds**;
`retry.Config.Backoff` is a pointer. Use `Strategy: "none"` to disable retries.

```go
package main

import (
    "context"
    "time"

    lambdadb "github.com/lambdadb/go-lambdadb"
    "github.com/lambdadb/go-lambdadb/models/operations"
    "github.com/lambdadb/go-lambdadb/retry"
)

func main() {
    client := lambdadb.New(lambdadb.WithAPIKey("<YOUR_PROJECT_API_KEY>"))
    _, _ = client.Collections.List(context.Background(), nil,
        operations.WithRetries(retry.Config{
            Strategy: "backoff",
            Backoff: &retry.BackoffStrategy{
                InitialInterval: 500,
                MaxInterval: 60000,
                Exponent: 1.5,
                MaxElapsedTime: 300000,
            },
            RetryConnectionErrors: true,
        }),
        operations.WithOperationTimeout(30*time.Second),
    )
}
```

## WithOperationTimeout

`WithOperationTimeout(time.Duration)` bounds a single API operation, including
its retries. This option uses Go durations, unlike the integer millisecond
backoff fields. It overrides `lambdadb.WithTimeout` for that operation. An earlier
caller context deadline still takes precedence; automatic storage transfers
use the caller context and transfer-client timeout.

## WithSetHeaders

`WithSetHeaders(map[string]string)` adds or overrides API request headers.
It does not configure headers on presigned storage transfers.
