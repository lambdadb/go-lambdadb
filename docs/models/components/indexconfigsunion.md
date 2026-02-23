# IndexConfigsUnion


## Supported Types

### IndexConfigsText

```go
indexConfigsUnion := components.CreateIndexConfigsUnionText(components.IndexConfigsText{/* values here */})
```

### IndexConfigsVector

```go
indexConfigsUnion := components.CreateIndexConfigsUnionVector(components.IndexConfigsVector{/* values here */})
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

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch indexConfigsUnion.Type {
	case components.IndexConfigsUnionTypeText:
		// indexConfigsUnion.IndexConfigsText is populated
	case components.IndexConfigsUnionTypeVector:
		// indexConfigsUnion.IndexConfigsVector is populated
	case components.IndexConfigsUnionTypeKeyword:
		// indexConfigsUnion.IndexConfigs is populated
	case components.IndexConfigsUnionTypeLong:
		// indexConfigsUnion.IndexConfigs is populated
	case components.IndexConfigsUnionTypeDouble:
		// indexConfigsUnion.IndexConfigs is populated
	case components.IndexConfigsUnionTypeDatetime:
		// indexConfigsUnion.IndexConfigs is populated
	case components.IndexConfigsUnionTypeBoolean:
		// indexConfigsUnion.IndexConfigs is populated
	case components.IndexConfigsUnionTypeSparseVector:
		// indexConfigsUnion.IndexConfigs is populated
	case components.IndexConfigsUnionTypeObject:
		// indexConfigsUnion.IndexConfigsObject is populated
}
```
