# Bayesian search and native embeddings

The current contract is pinned to
[OpenAPI docs@94a7b71613a3bf217a370d5117b3ff6ddfbae851](https://github.com/lambdadb/docs/blob/94a7b71613a3bf217a370d5117b3ff6ddfbae851/reference/api/openapi.json)
and backend
[lambdadb@0a54b6ecb483ce7df9677ba2071a491f9f9879c9](https://github.com/lambdadb/lambdadb/commit/0a54b6ecb483ce7df9677ba2071a491f9f9879c9).
Source compatibility does not verify a target environment's deployment.

## Bayesian hybrid search

Go queries remain free-form `map[string]any`; the SDK passes them through without
new typed query helpers or client-side fusion validation. The server requires:

- A top-level `bayesian` array containing exactly two signals. Use `bool` arrays
  to combine clauses within a signal.
- No explicit `boost`, including `1`, on either signal or its Boolean descendants.
- No nested `bayesian`, `rrf`, `mm`, or `l2`. Rank fusion is top-level only.
- Without rerank (including `rerank: null`), explicit top-level `candidateSize`
  with `1 <= size <= candidateSize <= 1000`. Omitting it returns HTTP 400.
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

Use exactly one non-null `queryText` or `queryVector` per KNN leg, with `k`
from 1 to 1000. Use query text for native embedding fields and query vectors
for caller-provided vector fields, with the configured dimensions. The pinned
server rejects direct query vectors on native embedding fields with HTTP 400;
see [native query input](api-contract.md#native-query-input).
Bayesian candidate calibration and final output size are separate budgets. Scores are fusion scores, not relevance probabilities.

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
longer inserts `embedding.similarity` while decoding an omitted value. Explicit
dimensions must be 64–4096 and no greater than the model maximum. The current
contract supports `cosine` and `euclidean`, with `cosine` as the documented
default. The server applies model-specific restrictions; see
[vector dimensions and similarity](api-contract.md#vector-dimensions-and-similarity).

### Unindexed stored source paths

`Embedding.SourceField` selects a scalar string in the original document, for
example `article.body`. Neither the source nor its object ancestors needs an
index declaration. A collection can declare only the generated vector; write
`{"id":"1","article":{"body":"Text to embed","title":"Stored sibling"}}`
through the ordinary `Docs` envelope. No wrapper or write-time source mapping is
needed. The same stored path can be selected by `RerankConfig.Fields`.

Declared source leaves must be `text` or `keyword`, and declared ancestors must
be objects. Empty path segments and array traversal are invalid. Newly supported
source paths reject null, arrays, objects, and non-string values before inference;
a missing source on a full upsert skips the vector. Existing indexed text source
behavior is preserved. Partial updates merge with stored JSON: unchanged source
text reuses an existing vector, changed text regenerates it, and an empty ancestor
object does not remove existing source text. Full upserts still replace documents
and do not preserve a previous vector when the source is omitted. Adding an
appropriate source index later uses the existing additive schema API.

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
embedding generation, nested unindexed sources, source update reuse/regeneration,
ordinary native/caller-vector KNN, Bayesian retrieval at candidate budgets
101/1000, null rerank, existing fusion methods, invalid requests, and two JEV
rerank calls on unindexed stored text with default/explicit candidate budgets. Client retries are disabled;
readiness polling precedes query embedding and rerank calls. It does not verify
relevance quality, throughput, usage accounting, or injected provider failures.
Revoke temporary keys, remove temporary projects, and verify absence separately;
never remove persistent CI resources.

Development validation follows [RELEASING.md](../RELEASING.md): consume exact
commit SHAs as Go pseudo-versions. There is no automatically published npm-style
`dev` package or dist-tag. Publishing automation, if needed, is a separate change.
