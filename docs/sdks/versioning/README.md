# Data Versioning

Use collection-scoped Branches, Tags, and Aliases to control document history
and read targets.

## Available operations

| Resource | Operations |
| --- | --- |
| `collection.Branches()` | `Create`, `List`, `Delete` |
| `collection.Tags()` | `Create`, `List`, `Delete` |
| `collection.Aliases()` | `Create`, `List`, `Retarget`, `Delete` |

All methods accept `context.Context` and optional `operations.Option` values.
Create methods return the created ref or alias. Delete methods return a
`MessageResponse`.

## Create refs

```go
collection := client.Collection("my-collection")

branch, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
	BranchName: "candidate",
	Source: &lambdadb.RefSource{
		Kind: lambdadb.RefSourceKindBranch,
		Name: "main",
	},
})

tag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{
	TagName: "validated-2026-09",
	Source: &lambdadb.RefSource{
		Kind: lambdadb.RefSourceKindBranch,
		Name: branch.Name,
	},
})

alias, err := collection.Aliases().Create(ctx, lambdadb.CreateAliasInput{
	AliasName: "production",
	Target: lambdadb.AliasTarget{
		Kind: lambdadb.RefSourceKindTag,
		Name: tag.Name,
	},
})
```

Omit `Source` to create a Branch or Tag from `main`. A Branch source may also
set `AsOf` to a Unix epoch millisecond cutoff. Tags are immutable.

## Retarget an alias

```go
alias, err := collection.Aliases().Retarget(
	ctx,
	"production",
	lambdadb.RetargetAliasInput{
		Target: lambdadb.AliasTarget{
			Kind: lambdadb.RefSourceKindTag,
			Name: "validated-2026-10",
		},
	},
)
```

## Select a ref for reads

Set `Ref` on Query, Fetch, or List options:

```go
result, err := collection.Query(ctx, lambdadb.QueryInput{
	Query: map[string]any{"queryString": map[string]any{"query": "*:*"}},
	Ref: &lambdadb.RefContext{
		Kind: lambdadb.RefKindAlias,
		Name: "production",
	},
})
```

`ConsistentRead` is valid only when `Ref` directly selects a Branch. Tags are
immutable, while Aliases resolve at request time and may be retargeted.

## Select a Branch for writes

Set `Branch` on Upsert, Update, Delete, or BulkUpsert inputs. It defaults to
`main` when omitted.

```go
_, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{
	Docs:   docs,
	Branch: lambdadb.String("candidate"),
})
```

For the two-step bulk flow, request an upload URL for the same Branch used in
the completion request:

```go
info, err := collection.Docs().GetBulkUpsertInfoForBranch(ctx, "candidate")
// Upload with info.Type and every entry in info.Headers, then:
_, err = collection.Docs().BulkUpsert(ctx, lambdadb.BulkUpsertInput{
	ObjectKey: info.ObjectKey,
	Type:      info.Type,
	Branch:    lambdadb.String("candidate"),
})
```

`BulkUpsertDocuments` applies the Branch to both control calls automatically.

For response and request field details, see the
[Data Versioning models](../../models/components/versioning.md).
