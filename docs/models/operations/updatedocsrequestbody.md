# UpdateDocsRequestBody


## Fields

| Field                                                                               | Type                                                                                | Required                                                                            | Description                                                                         |
| ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `Docs`                                                                              | []map[string]*any*                                                                  | :heavy_check_mark:                                                                  | A nonempty list of nonempty document objects to update. Each document must contain 'id' field to be updated. For native embedding fields, an explicitly supplied vector takes precedence; otherwise unchanged source input reuses the existing vector and changed source input regenerates it. |
| `Branch`                                                                            | **string*                                                                           | :heavy_minus_sign:                                                                  | Write target branch. Defaults to `main` when omitted. |
