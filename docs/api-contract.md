# Current API contract

SDK alignment is pinned to this immutable revision:

- [OpenAPI: docs@7be299af9cac4195b75feb6e38b5cd7f8989f5a0 `reference/api/openapi.json`](https://github.com/lambdadb/docs/blob/7be299af9cac4195b75feb6e38b5cd7f8989f5a0/reference/api/openapi.json),
  SHA-256 `ec499bb6ad4fd7359c1854359e9ecfd4cd8609420f440f7dcbf75566ccda8f32`,
  `info.version` `1.3.0`.
- Backend source of that copy:
  [lambdadb@f9aa63c79cbac2ac477a76ec45e5015ce665afd1 `api/openapi.json`](https://github.com/lambdadb/lambdadb/blob/f9aa63c79cbac2ac477a76ec45e5015ce665afd1/api/openapi.json),
  as recorded in `reference/api/openapi-source.json` of the same docs commit.

The backend owns the API contract; the public reference copy is
`reference/api/openapi.json` in [lambdadb/docs](https://github.com/lambdadb/docs),
pinned by commit and SHA-256. The previous pin (docs@1714009a,
`info.version` `1.1.1`, SHA-256
`604e6fdcf0e0a5af2c9ac8714e60d282d0ebfb9847a49371ff0452fac48be070`) differs
from this revision by Branch-scoped index schemas with the individual Branch,
Tag, and Alias routes, paginated ref listings, Branches created from Tags, the
full Collection shape in the create response, and the Branch, Tag, and listing
semantics of pending writes described below. Use the commit and SHA-256 to
identify this contract. SDK source alignment is not a live environment test or
release approval.

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
| Branch, Tag, and Alias list `size` | 100 | 10 |

Each KNN query uses exactly one non-null `queryText`, `queryInputs` or
`queryVector`. Query text and query inputs require native embeddings; query
vectors work on any vector field.
KNN `k`, final size, and candidate budgets remain independent. The SDK does not
clamp values or add local validation: invalid requests retain server validation
order and HTTP error classification. Empty schemas, document lists/objects and
explicit empty deletion IDs are rejected by the server. Nullable extended-List
size, pageToken and filter can be omitted by using nil SDK fields.

## Server-managed `id` field

Collection responses include the server-managed top-level `id` field as
`{"type":"keyword"}` in `IndexConfigs`. Create and Update requests may send it
back exactly as returned or omit it, so a schema read with Get can be reused as
the base of a full-schema update. A schema containing only `id` is accepted.
Any other `id` configuration, `id` inside an object field, or another reserved
internal field name is rejected by the server. The SDK forwards `IndexConfigs`
unchanged and does not validate field names. Deployments before backend #515
return `id` but reject it in requests; omit it there.

## Index field descriptions

Every `IndexConfigs` member, including nested fields under an object's
`objectIndexConfigs`, accepts an optional `description` string of at most 255
characters. It is a free-form explanation of what the field holds, returned in
Collection reads only when set, and never affects indexing. On create, an
omitted, null, or empty description means none. On Update, an omitted or null
description keeps the stored one, an empty string clears it, and any other
value replaces it; this is the only setting of an existing field that may
change, and it needs no reindexing. The server-managed `id` field rejects a
description. The SDK models it as `Description *string` on each member:
nil omits the key and `lambdadb.String("")` sends `""`. The length limit and
the `id` rule remain server checks.

## Consistent reads through an Alias

`consistentRead: true` is allowed when `ref` selects a Branch directly
(including the omitted-ref default `main`) or an Alias that currently targets
a Branch; the server resolves the Alias when the request is admitted. Tag
reads and an Alias that targets a Tag reject true with HTTP 400. The SDK
performs no local ref validation and forwards the request unchanged.

## Branch-scoped index schemas and ref reads

A Branch owns its index schema. `GET` and `PATCH
/collections/{collectionName}/branches/{branchName}` map to
`Branches().Get` and `Branches().Update`; `GET .../tags/{tagName}` and
`GET .../aliases/{aliasName}` map to `Tags().Get` and `Aliases().Get`. The
Branch update body is `indexConfigs` only (`UpdateBranchInput`), required and
nonempty, with the same additive rules as the Collection schema update.
Naming a Tag or an Alias returns 400.

Every Branch, Tag, and Alias response carries `indexConfigs`; Branches and
Tags also carry `numDocs` and `dataUpdatedAt`. `BranchDetails.DataUpdatedAt`
is a nullable pointer (null for an empty Branch) and
`AliasDetails.IndexConfigs` is nil for a dangling Alias, which the single-Alias
read still returns with 200. There is no schema version identifier in the
contract or the SDK.

Collection responses keep `IndexConfigs`, which mirror `main`, and
`POST /collections` now returns the same shape as Get.
`UpdateCollectionRequestBody.IndexConfigs` is deprecated: it still updates
`main` only; use `Branches().Update`.

## Ref and Collection list pagination

`GET .../branches`, `.../tags` and `.../aliases` return one page in name order
and take `size` (1 to 100, default 10) and `pageToken`; `nextPageToken` is
absent on the last page. These listings and `GET /collections` reject any other
query parameter and an empty or malformed `pageToken` with 400. The SDK therefore sends `size` and `pageToken` only when set and treats a
nil or empty `PageToken` as the first page, for `ListRefsOpts` and
`ListCollectionsOpts` alike. `GET /collections` declares the 400 response,
which the SDK returns as `BadRequestError`.

`List` on Branches, Tags and Aliases keeps its signature and still returns
every ref: it follows `nextPageToken` with the server's default page size.
`ListPage` and `ListIterator` expose single pages.

## Ref creation and pending writes

Branch creation accepts a Branch or Tag source; `asOf` is valid only for a
Branch source, and a Branch created from a Tag has a null `parentBranch`.
The SDK rejects Alias and unknown source kinds locally and leaves the
Tag-with-`asOf` rejection to the server.

Without `asOf`, a new Branch starts from the source's committed snapshot and
takes over the writes still pending on the source at creation; a new Tag pins
the snapshot together with those writes. Ordinary Branch reads use only the
snapshot, and `consistentRead: true` adds eligible inherited and own pending
writes. Query and Fetch through a Tag or an Alias targeting a Tag include the
captured writes but still reject `consistentRead: true`. Pending bulk imports
become visible only after a background build materializes them. Document
listing uses only the published snapshot for every ref. With `asOf`, only the
latest retained committed snapshot at or before the cutoff is selected. These
are server semantics; request bodies are unchanged.

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

Native embedding fields accept `knn.queryText`, ordered `knn.queryInputs`
parts (`{"type":"text"|"image","value":"..."}`), or a raw `knn.queryVector`.
Cohere accepts mixed text and image parts; OpenAI accepts one text part. Query
inputs use the same part, image and byte limits as document sources. A direct
query vector skips inference and its quota checks; it must come from a compatible
embedding space, which matching dimensions alone do not establish. Queries remain
free-form maps, and the SDK delegates these checks to the server.

## Cohere native embeddings and reranking

`EmbeddingConfig.Provider` accepts `openai` and `cohere`. Cohere supports
`embed-v5.0-pro` and `embed-v5.0-fast`, with dimensions 256, 512, 768, 1024,
1536 or 2048 (default 2048). Configure exactly one of the legacy `SourceField`
or ordered `Sources` (`EmbeddingSource{Field, Type}`); an empty `SourceField` is
omitted on the wire. OpenAI requires one text source. Cohere accepts up to 16
sources; image sources take a public HTTPS URL or supported data URI, or a flat
array of them, combined with text into one vector. Limits per input are 16
effective parts, 8 images, 1 MiB decoded image bytes, and 4,194,304 pixels per
image. URL, format, byte and pixel validation remain server checks.

Upsert and Update accept a caller-supplied vector on a native field. It takes
precedence over inference for that request; later source changes without a new
vector regenerate it. Bulk upload now accepts native embedding collections with
caller-supplied vectors only: a present source without its vector fails before
any document is committed (empty image arrays are skipped).

`RerankConfig` accepts `cohere` with `rerank-v5.0-pro` or `rerank-v5.0-fast`.
Cohere requires omitted or null `Criteria`; applied results report
`CriteriaVersion: "cohere-relevance-v1"`. TypeSafe/`jev-1.13.0` is unchanged.
See [native embeddings](bayesian-native-embeddings.md#cohere-multimodal-sources)
and [native reranking](native-reranking.md#cohere-reranking).

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
Bulk upload never generates native embeddings; include each vector.
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
400/503 error preservation. `cohere_contract_test.go` and the component tests
check query-input, native query-vector and Cohere rerank pass-through, Cohere
rerank metadata, `Sources` round trips without an empty `sourceField`, and
unchanged pass-through of a returned `id` field and an `id`-only schema.
`field_description_contract_test.go` checks `description` on every member and
on nested object fields, omission when nil, `""` when cleared, and read
round trips; `contract_alignment_test.go` checks that Alias refs forward
`consistentRead: true` unchanged and preserve the server's 400 for Tag refs
and Tag-targeting Aliases. `versioning_contract_test.go` checks the single
Branch, Tag and Alias reads and their 404, the Branch update body and its 400,
the ref fields on every response, the null `dataUpdatedAt` of an empty Branch
and null `indexConfigs` of a dangling Alias, ref list query parameters
(nothing sent when unset or empty), `List` following every page, the 400 and
omitted empty `pageToken` on the Collection listing, and the full create
response; `versioning_test.go` checks Tag sources for Branch creation.
`download_test.go` checks real HTTP gzip transfers,
all high-level download callers, metadata, empty/null arrays and corrupted or
interrupted downloads. These are local SDK tests, not server acceptance tests.

The opt-in `TestIntegrationReadLimitsSmoke` checks real query/fetch/GET and POST
List boundaries with 1000 small documents and no model inference calls.
`TestIntegrationBayesianNativeEmbeddingSmoke` checks unindexed nested source
writes, sibling/empty-object vector reuse, changed-source regeneration, invalid
source rejection, native-field direct query vectors, Bayesian budgets and
unindexed reranking. `TestIntegrationCohereNativeSmoke` re-sends a returned
`Sources` schema unchanged, including the server-managed `id` field, writes text, mixed, image-only and direct-vector documents,
runs `queryText`, mixed `queryInputs` and direct `queryVector` KNN, applies
Cohere reranking, rejects explicit Cohere criteria, and checks bulk vector
acceptance and source-without-vector rejection.
All create temporary collections and verify deletion through GET 404. Run only
in an authorized development or staging project after verifying deployment;
see [RELEASING.md](../RELEASING.md). Large-result capacity, quota accounting,
provider failure injection and production performance require separate validation.
