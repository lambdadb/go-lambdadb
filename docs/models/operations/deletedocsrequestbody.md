# DeleteDocsRequestBody


## Fields

| Field                                                                     | Type                                                                      | Required                                                                  | Description                                                               |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `Ids`                                                                     | []*string*                                                                | :heavy_minus_sign:                                                        | A list of document IDs.                                                   |
| `Filter`                                                                  | map[string]*any*                                                          | :heavy_minus_sign:                                                        | Query filter.                                                             |
| `PartitionFilter`                                                         | [*components.PartitionFilter](../../models/components/partitionfilter.md) | :heavy_minus_sign:                                                        | N/A                                                                       |