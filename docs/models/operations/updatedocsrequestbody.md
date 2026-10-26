# UpdateDocsRequestBody


## Fields

| Field                                                                               | Type                                                                                | Required                                                                            | Description                                                                         |
| ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `Docs`                                                                              | []map[string]*any*                                                                  | :heavy_check_mark:                                                                  | A nonempty list of nonempty document objects to update. Each document must contain 'id' field to be updated. For native embedding vector fields, omit the native embedding vector field and update only the configured source text field. |
| `Branch`                                                                            | **string*                                                                           | :heavy_minus_sign:                                                                  | Write target branch. Defaults to `main` when omitted. |
