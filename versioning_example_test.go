package lambdadb_test

import (
	"fmt"
	"time"

	lambdadb "github.com/lambdadb/go-lambdadb"
)

func ExampleCreateBranchInput() {
	input := lambdadb.CreateBranchInput{
		BranchName: "candidate",
		Source:     lambdadb.BranchSource("main"),
	}

	fmt.Println(input.BranchName, input.Source.Kind, input.Source.Name)
	// Output: candidate branch main
}

func ExampleBranchSourceAt() {
	cutoff := time.Date(2026, time.September, 3, 12, 0, 0, 0, time.UTC)
	source := lambdadb.BranchSourceAt("main", cutoff)

	fmt.Println(source.Kind, source.Name, time.UnixMilli(*source.AsOf).UTC().Format(time.RFC3339))
	// Output: branch main 2026-09-03T12:00:00Z
}

func ExampleListDocsOpts_ref() {
	opts := lambdadb.ListDocsOpts{
		Size: lambdadb.Int64(100),
		Ref:  lambdadb.AliasRef("production"),
	}

	fmt.Println(opts.Ref.Kind, opts.Ref.Name)
	// Output: alias production
}
