# QueryCollectionDoc


## Fields

| Field                      | Type                       | Required                   | Description                |
| -------------------------- | -------------------------- | -------------------------- | -------------------------- |
| `Collection`               | *string*                   | :heavy_check_mark:         | Collection name.           |
| `Score`                    | **float64*                 | :heavy_minus_sign:         | Final sorting score; applied reranking uses a 0–1 evaluation score, not a relevance probability. |
| `Doc`                      | map[string]*any*           | :heavy_check_mark:         | N/A                        |
| `RetrievalScore` | `*float64` | :heavy_minus_sign: | Original search/fusion score in the envelope, outside Doc; present only when reranking is applied. |

Score pointers distinguish numeric zero from absence and preserve double precision.
Without reranking or on fallback, `Score` retains the original search/fusion value
and `RetrievalScore` is absent. The SDK preserves server order.
