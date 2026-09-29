# Keyword facets

The SDK supports the keyword facet contract in:

- `lambdadb/lambdadb` commit `8da50bcd0b5a3c781ffccd7f01fb07ed0510dd30`,
  `api/src/main/java/ai/lambdadb/dto/QueryRequest.java`,
  `api/src/main/java/ai/lambdadb/model/query/FacetRequest.java`, and
  `api/src/main/java/ai/lambdadb/model/query/FacetResult.java`.
- `lambdadb/docs` commit `899092420ff801cfcb3b693b1ba273be7ac1f1ef`,
  `reference/api/openapi.json` and `guides/search/facets.mdx`.

These source commits do not establish deployment or package publication. Use a
server build containing the feature and rebuild existing data into a new collection.
Old keyword indexes and old Tags are unsupported; partial updates or segment merging
do not migrate their format.

Query-less document reads also require the server score-serialization fix in
[PR #426](https://github.com/lambdadb/lambdadb/pull/426), merged as
`335cb16fcef5d7b8d60f88c84f2ce2cf87f96939`. Older servers can return string
`"NaN"` scores that the Go SDK cannot decode as numbers. This applies to both
inline and downloaded document responses, including queries without facets.

## Semantics

Request up to five keyword fields by name, including dotted paths. A field's `size`
is 1–100 (default 10). `size: 0` at the query level requires at least one facet and
returns counts without documents. Omit `query` to match all documents.

Counts include all documents matching the query and partition filter within the
selected ref, independent of the returned document count. Each distinct indexed
array value contributes once per document. Missing or unindexed values produce no
bucket. Results are ordered by count descending, then Unicode code point order.
`total` still counts returned documents. Facets remain available after `docsUrl`
document downloads. Without a facet request, existing responses may omit facets.

Use queryString, bool combinations of supported queries, or no query. Vector,
sparse-vector, hybrid, numeric/date range facets, and excluding a facet's own filter
are outside this contract. Existing ref and consistentRead constraints still apply.
At more than 10,000 distinct (field, value) buckets or 262,144 UTF-8 bytes of distinct
values across fields, the server returns HTTP 400 instead of partial counts.

Keyword array sorting uses the smallest indexed value ascending and largest
descending. Missing values sort last ascending and first descending.

## Example

```go
package main

import (
    "context"
    "fmt"
    lambdadb "github.com/lambdadb/go-lambdadb"
    "github.com/lambdadb/go-lambdadb/models/components"
)

func main() {
    client := lambdadb.New(lambdadb.WithBaseURL("YOUR_BASE_URL"),
        lambdadb.WithProjectName("YOUR_PROJECT_NAME"), lambdadb.WithAPIKey("YOUR_API_KEY"))
    result, err := client.Collection("items").Query(context.Background(), lambdadb.QueryInput{
        Size: lambdadb.Int64(0),
        Facets: map[string]components.FacetRequest{"tags": {Size: lambdadb.Int64(5)}},
    })
    if err != nil { panic(err) }
    for _, bucket := range result.Facets["tags"].Buckets { fmt.Println(bucket.Value, bucket.Count) }
}
```

A nil facet Size uses the server default. Counts use int64. Both the operation
response and high-level QueryResult preserve the facet map.

## Review and validation

- [Facet models](../models/components/facets.go), [operation contract](../models/operations/querycollection.go),
  [high-level result](../results.go), [hydration](../collection.go), and [wire tests](../facet_contract_test.go).
- `go test ./...` and `go vet ./...` passed. Tests cover size zero, request payloads,
  facet-only and downloaded document responses, Unicode values, and int64 counts.
- Development live validation on 2026-09-29 passed all 22 facet cases at SDK
  merged commit `02d0099fbff33feaf649655de31d955ad0f7008b`, including query-less
  document reads, partition filters, Branch/Tag/Alias scope, and real `docsUrl`
  downloads after the server score fix. This validates the tested development
  environment, not production availability or package publication.
