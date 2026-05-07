# UpsertDocsRequestBody


## Fields

| Field                          | Type                           | Required                       | Description                    |
| ------------------------------ | ------------------------------ | ------------------------------ | ------------------------------ |
| `Docs`                         | []map[string]*any*             | :heavy_check_mark:             | A list of documents to upsert. For managed embedding vector fields, omit the managed vector field and provide only the configured source text field. |
