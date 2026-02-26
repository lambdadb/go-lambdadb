## Options

Per-call options (e.g. retries, timeout) for API methods. Client-level configuration (base URL, project name, API key) is set when creating the client via `lambdadb.WithBaseURL`, `lambdadb.WithProjectName`, `lambdadb.WithAPIKey`.

### WithRetries

WithRetries allows customizing the default retry configuration. Only usable with methods that support retries.

```go
operations.WithRetries(retry.Config{
    Strategy: "backoff",
    Backoff: retry.BackoffStrategy{
        InitialInterval: 500 * time.Millisecond,
        MaxInterval: 60 * time.Second,
        Exponent: 1.5,
        MaxElapsedTime: 5 * time.Minute,
    },
    RetryConnectionErrors: true,
})
```

### WithOperationTimeout

WithOperationTimeout sets the request timeout for a single operation.

### WithSetHeaders

WithSetHeaders sets additional or override headers for the request.
