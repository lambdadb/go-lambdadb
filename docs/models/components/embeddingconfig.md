# EmbeddingConfig

Native embedding configuration for vector fields.


## Fields

| Field                                                                      | Type                                                                       | Required                                                                   | Description                                                                |
| -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `Provider`                                                                 | [components.EmbeddingConfigProvider](../../models/components/embeddingconfigprovider.md) | :heavy_check_mark:                                                         | Embedding provider.                                                        |
| `Model`                                                                    | *string*                                                                   | :heavy_check_mark:                                                         | Embedding model name.                                                      |
| `SourceField`                                                              | *string*                                                                   | :heavy_check_mark:                                                         | Stored scalar string path through objects; no source index is required. Declared leaves must be text or keyword. Missing source skips generation; unindexed null/non-string values are rejected.                        |
| `Dimensions`                                                               | **int64*                                                                   | :heavy_minus_sign:                                                         | 64–4096, and no greater than the model maximum. Omit to use the model default; resolved in stored collection metadata. |
| `Similarity`                                                               | [*components.Similarity](../../models/components/similarity.md)            | :heavy_minus_sign:                                                         | `cosine` or `euclidean`; documented default: `cosine`. Omit to let the server apply model defaults and restrictions; resolved in stored collection metadata. |
