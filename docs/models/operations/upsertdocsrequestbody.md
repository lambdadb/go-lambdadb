# UpsertDocsRequestBody


## Fields

| Field                          | Type                           | Required                       | Description                    |
| ------------------------------ | ------------------------------ | ------------------------------ | ------------------------------ |
| `Docs`                         | []map[string]*any*             | :heavy_check_mark:             | A list of documents to upsert. For native embedding vector fields, omit the native embedding vector field and provide only the configured source text field. |
| `Branch`                       | **string*                      | :heavy_minus_sign:             | Write target branch. Defaults to `main` when omitted. |
