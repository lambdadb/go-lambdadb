# FieldsSelectorUnion

An object to specify a list of field names to include and/or exclude in the result.


## Supported Types

### FieldsSelector1

```go
fieldsSelectorUnion := components.CreateFieldsSelectorUnionFieldsSelector1(components.FieldsSelector1{/* values here */})
```

### FieldsSelector2

```go
fieldsSelectorUnion := components.CreateFieldsSelectorUnionFieldsSelector2(components.FieldsSelector2{/* values here */})
```

## Union Discrimination

Use the `Type` field to determine which variant is active, then access the corresponding field:

```go
switch fieldsSelectorUnion.Type {
	case components.FieldsSelectorUnionTypeFieldsSelector1:
		// fieldsSelectorUnion.FieldsSelector1 is populated
	case components.FieldsSelectorUnionTypeFieldsSelector2:
		// fieldsSelectorUnion.FieldsSelector2 is populated
}
```
