package lambdadb_test

import (
	"encoding/json"
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

func ExampleTagSource() {
	input := lambdadb.CreateTagInput{
		TagName: "validated-copy",
		Source:  lambdadb.TagSource("validated"),
	}
	fmt.Println(input.TagName, input.Source.Kind, input.Source.Name)
	// Output: validated-copy tag validated
}

func ExampleBranchDetails_parentBranch() {
	// A branch created from an empty source still records its direct parent.
	var branch lambdadb.BranchDetails
	if err := json.Unmarshal([]byte(`{"name":"candidate","parentBranch":{"branchId":"main-id","name":"main"},"headSnapshot":null,"parentSnapshot":null,"indexConfigs":{"id":{"type":"keyword"}},"numDocs":0,"dataUpdatedAt":null,"createdAt":1788336000123}`), &branch); err != nil {
		panic(err)
	}
	if parent := branch.GetParentBranch(); parent != nil {
		fmt.Println(parent.BranchID, parent.Name)
	}
	fmt.Println(branch.HeadSnapshot == nil, branch.ParentSnapshot == nil)
	// Output:
	// main-id main
	// true true
}

func ExampleRollbackBranchResponse() {
	// A tag without a published snapshot leaves the branch head empty.
	var restored lambdadb.RollbackBranchResponse
	if err := json.Unmarshal([]byte(`{"branchId":"main-branch-id","headSnapshot":null}`), &restored); err != nil {
		panic(err)
	}
	if head := restored.GetHeadSnapshot(); head != nil {
		fmt.Println(restored.BranchID, "restored to", head.SnapshotID)
	} else {
		fmt.Println(restored.BranchID, "has an empty head")
	}
	// Output: main-branch-id has an empty head
}

func ExampleTagDetails_beforeFirstSnapshot() {
	// A tag created before the first snapshot has no snapshot fields yet.
	var tag lambdadb.TagDetails
	if err := json.Unmarshal([]byte(`{"name":"pending","snapshotId":null,"snapshotCommittedAt":null,"numDocs":0,"dataUpdatedAt":null,"indexConfigs":{"id":{"type":"keyword"}},"createdAt":1788336000123}`), &tag); err != nil {
		panic(err)
	}
	fmt.Println(tag.SnapshotID == nil, tag.GetSnapshotID() == "", tag.GetSnapshotCommittedAt().IsZero(), tag.NumDocs)
	// Output: true true true 0
}

func ExampleListDocsOpts_sort() {
	opts := lambdadb.ListDocsOpts{
		Size:           lambdadb.Int64(10),
		Sort:           []map[string]any{{"category": "asc"}, {"publishedAt": "desc"}},
		ConsistentRead: lambdadb.Bool(true),
	}

	fmt.Println(opts.Sort[0]["category"], opts.Sort[1]["publishedAt"], *opts.ConsistentRead)
	// Output: asc desc true
}
