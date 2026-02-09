<!-- Start SDK Example Usage [usage] -->
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