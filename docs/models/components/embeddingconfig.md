# EmbeddingConfig

Native embedding configuration for vector fields.


## Fields

| Field                                                                      | Type                                                                       | Required                                                                   | Description                                                                |
| -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `Provider`                                                                 | [components.EmbeddingConfigProvider](../../models/components/embeddingconfigprovider.md) | :heavy_check_mark:                                                         | Embedding provider.                                                        |
| `Model` | *string* | :heavy_check_mark: | Embedding model name. OpenAI: `text-embedding-3-small`, `text-embedding-3-large`, `text-embedding-ada-002`. Cohere: `embed-v5.0-pro`, `embed-v5.0-fast`. |
| `SourceField` | *string* | :heavy_minus_sign: | Legacy single text source: stored scalar string path through objects; no source index is required. Set exactly one of `SourceField` or `Sources`; empty is omitted. Missing text skips generation; invalid values fail when inference is needed. |
| `Sources` | [][components.EmbeddingSource](../../models/components/embeddingsource.md) | :heavy_minus_sign: | 1–16 ordered sources with unique paths, combined into one vector. OpenAI requires exactly one text source; Cohere accepts text and image sources. |
| `Dimensions` | **int64* | :heavy_minus_sign: | 64–4096, and no greater than the model maximum. Cohere accepts 256, 512, 768, 1024, 1536 or 2048 (default 2048). Omit to use the model default; resolved in stored collection metadata. |
| `Similarity`                                                               | [*components.Similarity](../../models/components/similarity.md)            | :heavy_minus_sign:                                                         | `cosine` or `euclidean`; documented default: `cosine`. Omit to let the server apply model defaults and restrictions; resolved in stored collection metadata. |
