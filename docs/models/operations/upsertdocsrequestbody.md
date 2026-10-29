# UpsertDocsRequestBody


## Fields

| Field                          | Type                           | Required                       | Description                    |
| ------------------------------ | ------------------------------ | ------------------------------ | ------------------------------ |
| `Docs`                         | []map[string]*any*             | :heavy_check_mark:             | A nonempty list of nonempty document objects to upsert. For native embedding fields, supply source fields to generate a vector, or a vector directly to skip inference for that field. A full replacement with neither does not preserve the previous vector. |
| `Branch`                       | **string*                      | :heavy_minus_sign:             | Write target branch. Defaults to `main` when omitted. |
