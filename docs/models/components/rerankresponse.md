# RerankResponse

Whole-stage metadata on `QueryResult.Rerank` and the low-level query response.
Absent for requests without reranking. See [managed reranking](../../managed-reranking.md).

| Field | Go type | Required | Meaning |
| --- | --- | --- | --- |
| `Status` | `string` | Yes | applied, skipped, or fallback. |
| `Provider` | `string` | Yes | Requested provider. |
| `Model` | `string` | Yes | Requested model. |
| `ResolvedModel` | `*string` | No | Provider-reported model. |
| `CandidateCount` | `int64` | Yes | Actual unique candidate count. |
| `ScoredCount` | `int64` | Yes | Equal to candidate count when applied; zero on skip/fallback. |
| `Took` | `int64` | Yes | Rerank-stage milliseconds. |
| `CriteriaVersion` | `*string` | No | Applied only: default-relevance-v1 or custom. custom is not a content hash/unique version. |
| `Reason` | `*string` | No | skipped: noCandidates; fallback: timeout, rateLimit, unavailable, invalidResponse, or credentials. |

Zero counts and zero milliseconds are present values. Empty candidates return
skipped/noCandidates with zero counts. Fallback discards all partial scores.
These counts are diagnostic and are not billing units.
