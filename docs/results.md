# High-level results

Public read methods and bulk-upload info helpers return the flattened result
types below. Access their fields directly; they do not have an `Object` property.
Collection Create/Get/Update return `components.CollectionResponse` directly,
while Collection Delete and document mutation methods retain their
`models/operations` response wrappers. Types below are defined in
[results.go](../results.go).

## QueryResult

Returned by `collection.Query`.

| Field | Type | Meaning |
| --- | --- | --- |
| `Docs` | `[]operations.QueryCollectionDoc` | Document envelopes in server order, with optional scores. |
| `Took` | `int64` | Server-reported elapsed milliseconds. |
| `Total` | `int64` | Returned document count, not the full matching population. |
| `MaxScore` | `*float64` | Maximum final returned score; nil for no results, with zero a valid score. |
| `Facets` | `map[string]components.FacetResult` | Requested keyword facet results. |
| `Rerank` | `*components.RerankResponse` | Optional reranking stage metadata. |

See [native reranking](native-reranking.md) for final and retrieval score semantics.

## FetchResult

Returned by `collection.Docs().Fetch`.

| Field | Type | Meaning |
| --- | --- | --- |
| `Docs` | `[]operations.FetchDocsDoc` | Envelopes containing `Collection` and document content in `Doc`. |
| `Total` | `int64` | Returned document count. |
| `Took` | `int64` | Server-reported elapsed milliseconds. |

## ListDocsResult

Returned by `collection.Docs().List` and each `ListIterator.Next` call.

| Field | Type | Meaning |
| --- | --- | --- |
| `Docs` | `[]operations.ListDocsDoc` | Envelopes containing `Collection` and document content in `Doc`. |
| `Total` | `int64` | Document count in this page. |
| `NextPageToken` | `*string` | Opaque continuation token; absent/empty means no more pages. |

`ListAll` returns only document content as `[]map[string]any`. A short page is
not the end of pagination when it has a continuation token.

## ListCollectionsResult

Returned by `client.Collections.List` and each collection `ListIterator.Next` call.

| Field | Type | Meaning |
| --- | --- | --- |
| `Collections` | `[]components.CollectionResponse` | Collection metadata in the current page. |
| `NextPageToken` | `*string` | Opaque continuation token. |

Collection `ListAll` returns `[]components.CollectionResponse`.

## GetBulkUpsertInfoResult

Returned by `collection.Docs().GetBulkUpsertInfo` and `GetBulkUpsertInfoForBranch`.

| Field | Type | Meaning |
| --- | --- | --- |
| `URL` | `string` | Presigned upload URL. |
| `ObjectKey` | `string` | Key to send in the bulk-completion request. |
| `Type` | `*operations.Type` | Upload content type; defaults to `application/json` in the convenience helper. |
| `HTTPMethod` | `*operations.HTTPMethod` | Upload method; defaults to PUT in the convenience helper. |
| `SizeLimitBytes` | `*int64` | Actual upload budget when supplied by the server. |
| `Headers` | `map[string]string` | Signed headers to forward unchanged. |

Query, Fetch, and document List methods automatically download out-of-line
documents before returning. Downloads preserve response metadata; a failed
transfer returns an error instead of a partial result. The returned documents
remain fully resident in memory. See [result budgets](api-contract.md#result-budgets-and-downloads).
