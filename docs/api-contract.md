# Current API contract

SDK alignment is pinned to these immutable revisions:

- [OpenAPI: docs@94a7b71613a3bf217a370d5117b3ff6ddfbae851](https://github.com/lambdadb/docs/blob/94a7b71613a3bf217a370d5117b3ff6ddfbae851/reference/api/openapi.json).
- [Server: lambdadb@0a54b6ecb483ce7df9677ba2071a491f9f9879c9](https://github.com/lambdadb/lambdadb/commit/0a54b6ecb483ce7df9677ba2071a491f9f9879c9).

The OpenAPI `info.version` remains `1.1.1`; use the source SHA to identify this
contract. SDK source alignment is not a live environment test or release approval.

## Read limits

| Operation | Maximum | Default |
| --- | --- | --- |
| Query `size` | 1000 | 10; zero requires at least one facet |
| Dense KNN `k` | 1000 | Existing server behavior |
| Bayesian `candidateSize` without rerank | 1000 | Required; `1 <= size <= candidateSize` |
| Rerank `candidateSize` | 1000 | `max(50, size)`; size must be positive and no greater than the candidate budget |
| Fetch IDs | 1000 nonblank IDs | Required, nonempty list |
| GET/POST document List `size` | 1000 | 100; POST null also uses the default |
| Collection List and facet bucket size | 100 | Existing defaults unchanged |

Each KNN query uses exactly one non-null `queryText` or `queryVector`. Query text
requires native embeddings; query vectors are for caller-provided vector fields.
KNN `k`, final size, and candidate budgets remain independent. The SDK does not
clamp values or add local validation: invalid requests retain server validation
order and HTTP error classification. Empty schemas, document lists/objects and
explicit empty deletion IDs are rejected by the server. Nullable extended-List
size, pageToken and filter can be omitted by using nil SDK fields.

## Vector dimensions and similarity

The OpenAPI contract specifies 64–4096 dimensions for caller-provided vectors.
Native `Embedding.Dimensions` uses the same range and must not exceed the selected
model's supported maximum. Omit native dimensions to let the server resolve them.
Stored and query vectors must match the field's configured dimension count.

Both vector configurations advertise `cosine` and `euclidean`, with `cosine` as
the documented default. Native dimensions and similarity belong inside
`Embedding`; omitted values remain omitted on the wire so the server can apply
model defaults and restrictions.

The SDK retains its existing local checks and legacy enum constants for
compatibility. It does not enforce the 64–4096 range or native model limits
locally. `Similarity.IsExact()` recognizes legacy constants as well as current
contract values; it is not a current-contract validator. `dot_product` and
`max_inner_product` are outside the pinned OpenAPI contract and should not be
used for new configurations. See [Similarity](models/components/similarity.md).

## Native query input

The pinned OpenAPI and server both require `knn.queryText` on native embedding
fields; direct `knn.queryVector` returns HTTP 400. Use `knn.queryVector` on
caller-provided vector fields. The SDK delegates this validation to the server.

## Stored source paths

Native embedding sources and rerank input paths can select unindexed stored
scalar strings, including nested object paths. Declared leaves must be text or
keyword and declared ancestors must be objects. Array traversal and empty path
segments are invalid. See [native embedding](bayesian-native-embeddings.md#unindexed-stored-source-paths)
for write/partial-update behavior and [native reranking](native-reranking.md)
for candidate validation and projection. No new provider or SDK API is needed.

## Write payload budgets

Ordinary Upsert and Update batches share the Gateway request-body budget,
including the JSON envelope and every document. The pinned Gateway default is
6,029,312 bytes; oversized requests return HTTP 413. Independently, each document
must fit the existing 5 MiB limit, including generated native embeddings. Use
bulk upload for larger batches; it does not remove the per-document limit.

For presigned bulk uploads, honor `SizeLimitBytes` returned by the server.
`MaxBulkUpsertPayloadBytes` is a 200 MiB reference constant, not a separately
enforced client maximum. `BulkUpsertDocuments` validates a positive returned
limit before upload and does not substitute that constant when a limit is absent.
See the [bulk upload guide](sdks/docs/README.md#bulkupsert) for signed headers,
Branch binding and retry behavior.

## Result budgets and downloads

Count limits do not guarantee that all requested document bodies fit. Fetch
rejects projected bodies above its 64 MiB per-node JSON budget with HTTP 400 and
no partial result. Request fewer IDs/fields or disable `IncludeVectors` to reduce
that payload. Document List can return a shorter page; iterators continue using
`NextPageToken`. Result-storage admission failures return HTTP 503, distinct from
invalid input; retries follow the configured retry policy.

Query has no aggregate 64 MiB JSON result cap. Large results retain the existing
`isDocsInline: false` and `docsUrl` envelope. Query, Fetch, List, and the List
iterator download a complete document array, decoding one document at a time.
They preserve order, scores, facets, rerank metadata and page tokens. The final
slice still contains all decoded documents: this is not a streaming public API
or a fixed client-memory bound. With the default transfer client, Go's HTTP
transport automatically decodes gzip; custom transports must provide equivalent
content decoding. Malformed, truncated, trailing, or failed transfers return an
error without exposing a partial result. Downloads use the separate transfer
client and do not receive LambdaDB authentication.

## Validation

`current_contract_test.go` checks wire pass-through at 101/1000/1001 and server
400/503 error preservation. `download_test.go` checks real HTTP gzip transfers,
all high-level download callers, metadata, empty/null arrays and corrupted or
interrupted downloads. These are local SDK tests, not server acceptance tests.

The opt-in `TestIntegrationReadLimitsSmoke` checks real query/fetch/GET and POST
List boundaries with 1000 small documents and no model inference calls.
`TestIntegrationBayesianNativeEmbeddingSmoke` checks unindexed nested source
writes, sibling/empty-object vector reuse, changed-source regeneration, invalid
source rejection, native-field query-vector rejection, Bayesian budgets and
unindexed reranking.
Both create temporary collections and verify deletion through GET 404. Run only
in an authorized development or staging project after verifying deployment;
see [RELEASING.md](../RELEASING.md). Large-result capacity, quota accounting,
provider failure injection and production performance require separate validation.
