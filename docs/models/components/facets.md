# Keyword facet models

Models in [facets.go](../../../models/components/facets.go) are used by
`QueryInput.Facets` and `QueryResult.Facets`.

## FacetRequest

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Size` | `*int64` | No | Number of keyword buckets, from 1 to 100; server default 10. |

Requests are keyed by keyword field path. A query can request up to five facets.

## FacetResult

| Field | Type | Description |
| --- | --- | --- |
| `Buckets` | `[]components.FacetBucket` | Buckets ordered by descending count, then Unicode code point value order. |

## FacetBucket

| Field | Type | Description |
| --- | --- | --- |
| `Value` | `string` | Indexed keyword value. |
| `Count` | `int64` | Number of matching documents containing that value, independently of returned query size. |

See the [keyword facets guide](../../keyword-facets.md) for supported queries,
ref/filter semantics, resource budgets and an example.
