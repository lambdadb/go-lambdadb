# EmbeddingSource

One ordered source of a native embedding input. See
[Cohere multimodal sources](../../bayesian-native-embeddings.md#cohere-multimodal-sources).

## Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Field` | *string* | :heavy_check_mark: | Unique dotted path in the original stored JSON; no index declaration is required. |
| `Type` | `components.EmbeddingSourceType` | :heavy_check_mark: | `EmbeddingSourceTypeText` (`text`): a scalar string. `EmbeddingSourceTypeImage` (`image`, Cohere only): a public HTTPS URL or supported data URI, or a flat array of them. |
