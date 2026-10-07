# Bayesian search and native embeddings

This contract is pinned to backend
[`9072a1bc8925954369a887f558f1eaf387b7ea0e`](https://github.com/lambdadb/lambdadb/commit/9072a1bc8925954369a887f558f1eaf387b7ea0e).
The relevant contracts are
[QueryRequest](https://github.com/lambdadb/lambdadb/blob/9072a1bc8925954369a887f558f1eaf387b7ea0e/api/src/main/java/ai/lambdadb/dto/QueryRequest.java),
[BayesianQuery](https://github.com/lambdadb/lambdadb/blob/9072a1bc8925954369a887f558f1eaf387b7ea0e/api/src/main/java/ai/lambdadb/model/query/BayesianQuery.java),
and [FieldConfig](https://github.com/lambdadb/lambdadb/blob/9072a1bc8925954369a887f558f1eaf387b7ea0e/api/src/main/java/ai/lambdadb/model/collection/FieldConfig.java).
Source compatibility does not establish deployment or production availability.

## Bayesian hybrid search

Go queries remain free-form `map[string]any`; the SDK passes them through without
new typed query helpers or client-side fusion validation. The server requires:

- A top-level `bayesian` array containing exactly two signals. Use `bool` arrays
  to combine clauses within a signal.
- No explicit `boost`, including `1`, on either signal or its Boolean descendants.
- No nested `bayesian`, `rrf`, `mm`, or `l2`. Rank fusion is top-level only.
- Without rerank (including `rerank: null`), explicit top-level `candidateSize`
  with `1 <= size <= candidateSize <= 100`. Omitting it returns HTTP 400.
- With rerank, omit top-level `candidateSize`. Use `rerank.candidateSize`, or
  leave it unset for the existing server default `max(50, size)`.

No fusion weights are required. Existing text/KNN/RRF/Min-Max/L2 requests must
omit top-level `candidateSize`; they retain their existing defaults. The SDK
never inserts a candidate budget, output size, KNN `k`, or fusion weight.
Invalid requests retain the server's validation order and error classification,
including `apierrors.BadRequestError` for HTTP 400.

```go
query := map[string]any{"bayesian": []any{
    map[string]any{"queryString": map[string]any{"query": "body:restore"}},
    map[string]any{"knn": map[string]any{
        "field": "vector", "queryText": "restore a previous version", "k": 30,
    }},
}}
retrieved, err := collection.Query(ctx, lambdadb.QueryInput{
    Query: query, Size: lambdadb.Int64(10), CandidateSize: lambdadb.Int64(30),
})
```

For caller-provided vector fields, use `queryVector` with the configured dimensions
instead of `queryText`. Bayesian candidate calibration and final output size are
separate budgets. Scores are fusion scores, not relevance probabilities.

```go
config := components.RerankConfig{
    Provider: "typesafe", Model: "jev-1.13.0",
    QueryText: "How do I restore a previous version?", Fields: []string{"body"},
}
reranked, err := collection.Query(ctx, lambdadb.QueryInput{
    Query: query, Size: lambdadb.Int64(10),
    Rerank: optionalnullable.From(&config),
})
```

Bayesian fusion runs before native reranking. Applied reranking preserves the
original Bayesian score in `RetrievalScore`; `Score` is the final evaluation
score. Existing metadata, fallback, projection and downloaded response handling
remain unchanged. See [native reranking](native-reranking.md).

## Native embedding configuration

The preferred JSON input works for collection creation and schema updates:

```json
{
  "type": "vector",
  "embedding": {
    "provider": "openai",
    "model": "text-embedding-3-small",
    "sourceField": "body"
  }
}
```

Use the native union helper to preserve omission:

```go
vector := components.CreateIndexConfigsUnionNativeEmbeddingVector(
    components.IndexConfigsNativeEmbeddingVector{
        Embedding: components.EmbeddingConfig{
            Provider: components.EmbeddingConfigProviderOpenai,
            Model: "text-embedding-3-small", SourceField: "body",
        },
    },
)
schema := map[string]components.IndexConfigsUnion{
    "body": components.CreateIndexConfigsUnionText(components.IndexConfigsText{}),
    "vector": vector,
}
_, err := client.Collections.Create(ctx, lambdadb.CreateCollectionOptions{
    CollectionName: "articles", IndexConfigs: schema,
})
_, err = client.Collection("articles").Update(ctx, lambdadb.UpdateCollectionOptions{
    IndexConfigs: schema,
})
```

Update requests supply the full schema, preserving existing fields and settings.
Native `Dimensions` and `Similarity` belong inside `Embedding`. Their omission
survives JSON decode/encode and leaves model defaults to the server. The SDK no
longer inserts `embedding.similarity` while decoding an omitted value.

`IndexConfigsManagedEmbeddingVector` and
`CreateIndexConfigsUnionManagedEmbeddingVector` retain their existing behavior,
including serializing `managedEmbedding: true`. The native type also accepts
`ManagedEmbedding: lambdadb.Bool(true)` when explicit legacy input is needed;
explicit false with embedding remains invalid. Server responses normalized to
`managedEmbedding: true` continue to use the legacy union member and preserve
resolved dimensions/similarity. Caller-provided `IndexConfigsVector` behavior
and its defaults are unchanged.

This repository contains no CLI. The separately maintained TypeScript CLI uses
JSON/file input and its own SDK dependency; updating this Go module does not
update or validate that CLI.

## Opt-in live validation

First verify the actual deployed revision through deployment records and running
image digests. Use an authorized test project and load these only in memory:
`LAMBDADB_BASE_URL`, `LAMBDADB_PROJECT_NAME`, `LAMBDADB_PROJECT_API_KEY`.

```bash
LAMBDADB_RUN_BAYESIAN_SMOKE=1 \
  go test -run '^TestIntegrationBayesianNativeEmbeddingSmoke$' -count=1 -v .
```

The test creates unique temporary collections and verifies deletion by GET 404.
It covers native and legacy create/update, normalized metadata, actual OpenAI
embedding generation, ordinary native/caller-vector KNN, Bayesian retrieval,
null rerank, existing fusion methods, eighteen invalid requests, and two JEV
rerank calls with default/explicit candidate budgets. Client retries are disabled;
readiness polling precedes query embedding and rerank calls. It does not verify
relevance quality, throughput, usage accounting, or injected provider failures.
Revoke temporary keys, remove temporary projects, and verify absence separately;
never remove persistent CI resources.

Development validation follows [RELEASING.md](../RELEASING.md): consume exact
commit SHAs as Go pseudo-versions. There is no automatically published npm-style
`dev` package or dist-tag. Publishing automation, if needed, is a separate change.
