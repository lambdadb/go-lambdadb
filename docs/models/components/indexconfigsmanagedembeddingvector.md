# IndexConfigsManagedEmbeddingVector


This legacy type/helper always serializes `managedEmbedding: true`. For preferred
embedding-only input, use `IndexConfigsNativeEmbeddingVector`; see
[native embeddings](../../bayesian-native-embeddings.md#native-embedding-configuration).

## Fields

| Field                                                               | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `Type`                                                              | [components.TypeVector](../../models/components/typevector.md)      | :heavy_check_mark:                                                  | N/A                                                                 |
| `ManagedEmbedding`                                                  | *bool*                                                              | :heavy_check_mark:                                                  | Managed embedding vector field.                                     |
| `Embedding`                                                         | [components.EmbeddingConfig](../../models/components/embeddingconfig.md) | :heavy_check_mark:                                                  | N/A                                                                 |
