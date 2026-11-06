# IndexConfigsUnion


## Field descriptions

Every member has an optional `Description *string` (`description` on the
wire): a free-form explanation of what the field holds, at most 255
characters, validated by the server. It never affects indexing. Nested fields
inside `IndexConfigsObject.ObjectIndexConfigs` accept the same `description`
key.

- Create: nil omits it, which means no description.
- Update: nil keeps the stored description, `lambdadb.String("")` clears it,
  and any other value replaces it. A description-only change on an existing
  field is accepted; its other settings remain immutable.
- Reads return `description` only when one is set, so nil means none.
- The server-managed `id` field rejects a description.

```go
components.CreateIndexConfigsUnionText(components.IndexConfigsText{
    Type:        components.TypeTextText,
    Description: lambdadb.String("Book title as printed on the cover"),
})
```

## Supported Types

### IndexConfigsText

```go
indexConfigsUnion := components.CreateIndexConfigsUnionText(components.IndexConfigsText{/* values here */})
```

### IndexConfigsVector

```go
indexConfigsUnion := components.CreateIndexConfigsUnionVector(components.IndexConfigsVector{/* values here */})
```

### IndexConfigsNativeEmbeddingVector

```go
indexConfigsUnion := components.CreateIndexConfigsUnionNativeEmbeddingVector(components.IndexConfigsNativeEmbeddingVector{/* values here */})
```

See [IndexConfigsNativeEmbeddingVector](indexconfigsnativeembeddingvector.md)
for its fields and [native embeddings](../../bayesian-native-embeddings.md#native-embedding-configuration)
for embedding-only create/update input and legacy compatibility.

### IndexConfigsManagedEmbeddingVector

```go
indexConfigsUnion := components.CreateIndexConfigsUnionManagedEmbeddingVector(components.IndexConfigsManagedEmbeddingVector{/* values here */})
```

### IndexConfigs

```go
indexConfigsUnion := components.CreateIndexConfigsUnionKeyword(components.IndexConfigs{/* values here */})
```

### IndexConfigs

```go
indexConfigsUnion := components.CreateIndexConfigsUnionLong(components.IndexConfigs{/* values here */})
```

### IndexConfigs

```go
indexConfigsUnion := components.CreateIndexConfigsUnionDouble(components.IndexConfigs{/* values here */})
```

### IndexConfigs

```go
indexConfigsUnion := components.CreateIndexConfigsUnionDatetime(components.IndexConfigs{/* values here */})
```

### IndexConfigs

```go
indexConfigsUnion := components.CreateIndexConfigsUnionBoolean(components.IndexConfigs{/* values here */})
```

### IndexConfigs

```go
indexConfigsUnion := components.CreateIndexConfigsUnionSparseVector(components.IndexConfigs{/* values here */})
```

### IndexConfigsObject

```go
indexConfigsUnion := components.CreateIndexConfigsUnionObject(components.IndexConfigsObject{/* values here */})
```
