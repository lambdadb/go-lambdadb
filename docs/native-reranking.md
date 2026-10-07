# Native reranking

Select reranking separately for each scoring search request through `QueryInput.Rerank`.
It is not a collection setting. LambdaDB manages provider credentials; use only your
normal LambdaDB project API key. No Jev API key is required from the caller.
An unset or explicitly null `rerank` retains existing search behavior.

## Default criteria

This example assumes a native embedding vector field named `bodyEmbedding` and stored scalar
text fields `title` and `body`:

```go
config := components.RerankConfig{
    Provider: "typesafe",
    Model: "jev-1.13.0",
    QueryText: "How do I restore a previous collection version?",
    Fields: []string{"title", "body"},
}
result, err := collection.Query(ctx, lambdadb.QueryInput{
    Size: lambdadb.Int64(10),
    Query: map[string]any{"knn": map[string]any{
        "field": "bodyEmbedding",
        "queryText": config.QueryText,
        "k": 50,
    }},
    Rerank: optionalnullable.From(&config),
})
```

Import `github.com/lambdadb/go-lambdadb/optionalnullable` alongside the SDK and
`models/components`. Leave `Criteria` unset to use the default criteria, or set
`config.Criteria = optionalnullable.From[[]string](nil)` to send explicit null.
To send a null rerank object, use
`Rerank: optionalnullable.From[components.RerankConfig](nil)`.
The zero value of either wrapper omits that property entirely.

## Custom criteria

Use ordered descriptions from lowest to highest relevance. For example:

```go
criteria := []string{
    "Does not explain how to restore a collection version.",
    "Explains part of the restore procedure but leaves required steps missing.",
    "Explains all steps needed to restore the requested collection version.",
}
config.Criteria = optionalnullable.From(&criteria)
config.CandidateSize = lambdadb.Int64(50)
config.OnFailure = lambdadb.String("returnOriginal")
```

Use the updated `config` in the query's `Rerank` wrapper. These descriptions are
an illustrative example, not a validated quality rubric. Custom criteria must
contain 2–10 distinct nonblank strings, at most 2 KiB UTF-8 each and 8 KiB total.
Text and ordering are preserved. An empty list is invalid and is sent unchanged
for server validation; it does not select the default. Custom criteria replace
the default descriptions for that request. Supplying ten custom descriptions
does not reproduce the default scoring configuration.

## Request constraints and candidate counts

| Input | Meaning |
| --- | --- |
| `provider`, `model` | Required: `typesafe`, `jev-1.13.0`. |
| `queryText` | Required nonblank text, at most 8 KiB UTF-8, independent of retrieval query text or vectors. |
| `fields` | 1–8 unique stored scalar text paths; schema types `text` or `keyword`. Nested object paths are allowed; arrays and non-string values are not. |
| `size` | Final number of returned documents, positive and at most 100 when reranking. |
| `knn.k` | Number of candidates from that dense vector retrieval leg. |
| `candidateSize` | Cap after global candidate merge and deduplication; defaults to `max(50, size)`, with `size <= candidateSize <= 100`. |
| `onFailure` | `error` (default) or `returnOriginal`. |

The SDK does not automatically change `knn.k`. A vector leg with `k=20` and a
lexical leg contributing 50 results can supply a merged pool capped at 50.
For a vector-only query, increase `k` explicitly when a deeper pool is wanted.
The cap does not guarantee that many candidates; metadata reports actual counts.
Sparse vector and lexical queries have no separate dense `knn.k` control.

[Bayesian search](bayesian-native-embeddings.md) supports native reranking. Omit
top-level `candidateSize` and use `rerank.candidateSize`; its default remains
`max(50, size)`. Applied reranking preserves Bayesian scores in `RetrievalScore`.

Reranking requires a scoring retrieval query. It cannot be combined with `sort`
or used for query-less/filter-only requests. Existing facet restrictions remain:
reranking does not add facet support to vector/hybrid queries. Supported lexical
facet counts survive reranking and fallback, and are not recomputed from the
returned top documents. Facet-only `size: 0` remains valid without reranking.

The server validates input before empty-result handling and provider calls.
Selected fields are provider inputs; response `Fields` projection remains
independent. Missing/null fields contribute no text, but a candidate with no
nonblank selected text or a non-string selected value fails validation.
Rendered candidate text is limited to 16 KiB and combined query/candidate/custom
criteria input to 256 KiB UTF-8. Inputs are not silently truncated.
The SDK delegates these checks and defaults to the server. Unsupported options
such as `weights` and `threshold` are not exposed; decoding them into
`RerankConfig` returns an error instead of silently discarding them.

## Scores and metadata

For applied reranking, `result.Docs[i].Score` is the final 0–1 evaluation score
used for ordering. It is not a relevance probability. `RetrievalScore` is the
original search/fusion score in the same document envelope, outside `Doc`.
Both use `*float64`; zero is a valid present score, distinct from nil. The SDK
preserves server order and precision, including exact ties, without re-sorting.
Do not interpret equal numbers under different criteria or models as equal
relevance. Low-scoring results are not removed by an implicit threshold.

`MaxScore` is the maximum final returned score, including fallback search scores;
it is nil for empty results. `Total` is the returned document count, not a corpus
match count. The top-level `Took` includes the complete query.
`result.Rerank` preserves the stage metadata even when documents are downloaded
from `docsUrl`:

| Field | Presence and meaning |
| --- | --- |
| `Status` | Required: `applied`, `skipped`, or `fallback`. |
| `Provider`, `Model` | Required requested provider/model. |
| `ResolvedModel` | Optional model reported by the provider. |
| `CandidateCount` | Required actual unique candidate count. |
| `ScoredCount` | Required; equals candidate count when applied, zero when skipped/fallback. Partial scores are never exposed. |
| `Took` | Required rerank-stage milliseconds. |
| `CriteriaVersion` | Applied only: `default-relevance-v1` or `custom`. `custom` identifies caller-supplied criteria, not a hash or unique version. |
| `Reason` | Skipped/fallback diagnostic code. |

For example, an applied custom result can contain a valid zero score:

```json
{
  "took": 20,
  "total": 1,
  "maxScore": 0,
  "isDocsInline": true,
  "docs": [{
    "collection": "articles",
    "score": 0,
    "retrievalScore": 3.5,
    "doc": {"id": "article-1", "title": "Restore guide"}
  }],
  "rerank": {
    "status": "applied",
    "provider": "typesafe",
    "model": "jev-1.13.0",
    "candidateCount": 1,
    "scoredCount": 1,
    "took": 12,
    "criteriaVersion": "custom"
  }
}
```

No candidates produces `skipped`/`noCandidates`, zero counts, and no
`CriteriaVersion` or `MaxScore`. When reranking is unused, metadata and
`RetrievalScore` are absent. Fallback retains original search/fusion scores and
omits `RetrievalScore` and `CriteriaVersion`. The API does not expose
`rerankScore`, `rubricVersion`, user weights, or thresholds.

## Failure behavior

`returnOriginal` applies to the entire rerank stage for eligible provider failures:
`timeout`, `rateLimit`, `unavailable`, `invalidResponse` (malformed/incomplete
scores), or `credentials` (credential rejection without credential details).
It returns the first `size` documents in retained pre-rerank candidate order,
with `Status: "fallback"`, a reason, and `ScoredCount: 0`.
For a hybrid query, this expanded fused pool can differ from a separate legacy
query using smaller candidate depth. No partial rerank scores are combined with
original scores. The default `error` policy returns a server error.

Invalid input or selected text, disabled models, inference quota/configuration
errors, retrieval/hydration/version-routing failures, and authorization errors
are not provider fallback cases. Cancellation or an exhausted overall request
deadline does not start fallback work beyond that deadline. The SDK propagates
server errors and does not implement its own fallback or retry-specific ranking.

## Contract source and availability

The implementation follows backend develop revision
[lambdadb/lambdadb@55d888299fee44466326a9db8016af9811ade13b](https://github.com/lambdadb/lambdadb/blob/55d888299fee44466326a9db8016af9811ade13b/docs/design/managed-reranking.md).
Its relevant request/response DTOs, service, contract tests, and design document
match the supplied local checkout at `a5e06d49be06d95dc5f4046aeecaf51f8a7733c0`. The upstream
[OpenAPI at 961561c379acb079aec20191e13b89809ef096e9](https://github.com/lambdadb/docs/blob/961561c379acb079aec20191e13b89809ef096e9/reference/api/openapi.json)
has no rerank schema yet; that is a separate upstream follow-up.
This manually maintained SDK has no active generator. Source commits and SDK
types do not establish availability in a target environment.
