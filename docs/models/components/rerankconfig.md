# RerankConfig

Query-level server-managed reranking; see [managed reranking](../../managed-reranking.md)
for default/custom examples, nullable wrappers, validation, and fallback behavior.

| Field | Go type | Required | Contract |
| --- | --- | --- | --- |
| `Provider` | `string` | Yes | `typesafe`. |
| `Model` | `string` | Yes | `jev-1.13.0`. |
| `QueryText` | `string` | Yes | Nonblank, at most 8 KiB UTF-8. |
| `Fields` | `[]string` | Yes | 1–8 unique stored scalar text paths, schema type text or keyword. |
| `CandidateSize` | `*int64` | No | Default `max(50, size)`; positive size <= candidateSize <= 100. |
| `OnFailure` | `*string` | No | `error` (default) or `returnOriginal`. |
| `Criteria` | `optionalnullable.OptionalNullable[[]string]` | No | Omitted/null selects default; otherwise 2–10 distinct nonblank descriptions, low to high relevance, at most 2 KiB each and 8 KiB total UTF-8. |

The server applies validation and defaults. Required fields are serialized even
when invalid; optional explicit empty criteria remain `[]` for server rejection.
Unsupported JSON properties are rejected by this model's `UnmarshalJSON`.
No provider credentials or custom weights/thresholds are accepted.
