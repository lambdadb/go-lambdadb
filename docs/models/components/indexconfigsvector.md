# IndexConfigsVector


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `Type`                                                          | [components.TypeVector](../../models/components/typevector.md)  | :heavy_check_mark:                                              | N/A                                                             |
| `ManagedEmbedding`                                              | **bool*                                                         | :heavy_minus_sign:                                              | Set to false or omit for vector fields with caller-provided values.               |
| `Dimensions`                                                    | *int64*                                                         | :heavy_check_mark:                                              | Vector dimension count, from 64 to 4096.                  |
| `Similarity`                                                    | [*components.Similarity](../../models/components/similarity.md) | :heavy_minus_sign:                                              | `cosine` or `euclidean`; default: `cosine`.           |

The server validates dimension bounds. The SDK preserves its existing local
checks and legacy similarity compatibility; see the [current contract](../../api-contract.md#vector-dimensions-and-similarity).
