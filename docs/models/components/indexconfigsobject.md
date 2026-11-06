# IndexConfigsObject


## Fields

| Field                                                          | Type                                                           | Required                                                       | Description                                                    |
| -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| `Type`                                                         | [components.TypeObject](../../models/components/typeobject.md) | :heavy_check_mark:                                             | N/A                                                            |
| `ObjectIndexConfigs`                                           | map[string]*any*                                               | :heavy_check_mark:                                             | Nested field configurations. Each nested field accepts the same optional `description` key. |
| `Description` | **string* | :heavy_minus_sign: | Optional explanation of the field, at most 255 characters. Nil omits it; `lambdadb.String("")` clears it on update. Returned only when set. See [field descriptions](indexconfigsunion.md#field-descriptions). |