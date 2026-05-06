# IndexConfigsVector


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `Type`                                                          | [components.TypeVector](../../models/components/typevector.md)  | :heavy_check_mark:                                              | N/A                                                             |
| `ManagedEmbedding`                                              | **bool*                                                         | :heavy_minus_sign:                                              | Set to false or omit for unmanaged vector fields.               |
| `Dimensions`                                                    | *int64*                                                         | :heavy_check_mark:                                              | Vector dimensions for unmanaged vector fields.                  |
| `Similarity`                                                    | [*components.Similarity](../../models/components/similarity.md) | :heavy_minus_sign:                                              | Vector similarity metric for unmanaged vector fields.           |
