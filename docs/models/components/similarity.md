# Similarity

Vector similarity constants retained by the SDK. A constant's presence does
not make it valid for every field or embedding model; the server applies
field/model-specific restrictions. Omit native embedding similarity to use the
model default.


## Values

| Name                        | Value                       |
| --------------------------- | --------------------------- |
| `SimilarityCosine`          | cosine                      |
| `SimilarityEuclidean`       | euclidean                   |
| `SimilarityDotProduct`      | dot_product                 |
| `SimilarityMaxInnerProduct` | max_inner_product           |