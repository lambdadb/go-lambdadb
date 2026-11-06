# IndexConfigs

Types that do not need additional parameters.


## Fields

| Field                                              | Type                                               | Required                                           | Description                                        |
| -------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------- |
| `Type`                                             | [components.Type](../../models/components/type.md) | :heavy_check_mark:                                 | N/A                                                |
| `Description` | **string* | :heavy_minus_sign: | Optional explanation of the field, at most 255 characters. Nil omits it; `lambdadb.String("")` clears it on update. Returned only when set. See [field descriptions](indexconfigsunion.md#field-descriptions). |