# IndexConfigsManagedEmbeddingVector


This legacy type/helper always serializes `managedEmbedding: true`. For preferred
embedding-only input, use `IndexConfigsNativeEmbeddingVector`; see
[native embeddings](../../bayesian-native-embeddings.md#native-embedding-configuration).

## Fields

| Field                                                               | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| `Type`                                                              | [components.TypeVector](../../models/components/typevector.md)      | :heavy_check_mark:                                                  | N/A                                                                 |
| `ManagedEmbedding`                                                  | *bool*                                                              | :heavy_check_mark:                                                  | Legacy flag for native embedding vector fields.                                     |
| `Embedding`                                                         | [components.EmbeddingConfig](../../models/components/embeddingconfig.md) | :heavy_check_mark:                                                  | N/A                                                                 |
| `Description` | **string* | :heavy_minus_sign: | Optional explanation of the field, at most 255 characters. Nil omits it; `lambdadb.String("")` clears it on update. Returned only when set. See [field descriptions](indexconfigsunion.md#field-descriptions). |
