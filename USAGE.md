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

	// Collection-scoped: no need to pass collection name every time
	coll := client.Collection("my-collection")
	meta, _ := coll.Get(ctx)
	coll.Docs().List(ctx, nil)
}
```
<!-- End SDK Example Usage [usage] -->
