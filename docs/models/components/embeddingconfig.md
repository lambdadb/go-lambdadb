# EmbeddingConfig

Managed embedding configuration for vector fields.


## Fields

| Field                                                                      | Type                                                                       | Required                                                                   | Description                                                                |
| -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `Provider`                                                                 | [components.EmbeddingConfigProvider](../../models/components/embeddingconfigprovider.md) | :heavy_check_mark:                                                         | Embedding provider.                                                        |
| `Model`                                                                    | *string*                                                                   | :heavy_check_mark:                                                         | Embedding model name.                                                      |
| `SourceField`                                                              | *string*                                                                   | :heavy_check_mark:                                                         | Source text field name used to generate embeddings.                        |
| `Dimensions`                                                               | **int64*                                                                   | :heavy_minus_sign:                                                         | Resolved embedding dimensions. Optional in requests and resolved in stored collection metadata. |
| `Similarity`                                                               | [*components.Similarity](../../models/components/similarity.md)            | :heavy_minus_sign:                                                         | Resolved vector similarity metric. Optional in requests and resolved in stored collection metadata. |
