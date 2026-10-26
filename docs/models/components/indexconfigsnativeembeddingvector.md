# IndexConfigsNativeEmbeddingVector

Preferred embedding-only input for collection creation or schema updates.
Use `CreateIndexConfigsUnionNativeEmbeddingVector` to construct the union member.

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `Type` | `components.TypeVector` | Yes | `vector`; the union helper supplies it. |
| `Embedding` | [components.EmbeddingConfig](embeddingconfig.md) | Yes | Provider, model, source path, and optional model dimensions/similarity. |
| `ManagedEmbedding` | `*bool` | No | Omit or set true. Explicit false with an embedding config is invalid. |

Dimensions and similarity belong inside `Embedding`, not at the vector's top
level. Omitted values remain omitted so the server selects model defaults.
Server responses normalized to `managedEmbedding: true` decode into the legacy
`IndexConfigsManagedEmbeddingVector` union member. Existing legacy types/helpers
are retained. See [native embedding configuration](../../bayesian-native-embeddings.md#native-embedding-configuration).
