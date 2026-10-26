# Similarity

The [current OpenAPI contract](../../api-contract.md#vector-dimensions-and-similarity)
supports the following values for caller-provided and native embedding vectors.
The documented default is `cosine`. Omit native embedding similarity to let the
server apply model defaults and restrictions.

## Current contract values

| Name | Value |
| --- | --- |
| `SimilarityCosine` | `cosine` |
| `SimilarityEuclidean` | `euclidean` |

## Legacy constants

These constants remain available for source and decoding compatibility, but are
outside the current OpenAPI contract. Do not use them for new configurations.

| Name | Value |
| --- | --- |
| `SimilarityDotProduct` | `dot_product` |
| `SimilarityMaxInnerProduct` | `max_inner_product` |

`IsExact()` recognizes all four historical SDK enum values. It does not validate
current API support or model-specific restrictions; the server validates those.
