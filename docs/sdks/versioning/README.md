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
	Source:     lambdadb.BranchSource("main"),
})

tag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{
	TagName: "validated-2026-09",
	Source:  lambdadb.BranchSource(branch.Name),
})

alias, err := collection.Aliases().Create(ctx, lambdadb.CreateAliasInput{
	AliasName: "production",
	Target:    lambdadb.TagTarget(tag.Name),
})
```

Omit `Source` to create a Branch or Tag from `main`. Use `BranchSourceAt` to
select the latest committed snapshot at or before a `time.Time` cutoff without
manually converting it to Unix milliseconds:

```go
historical, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
	BranchName: "historical",
	Source:     lambdadb.BranchSourceAt("main", cutoff),
})
```

Use `TagSource(name)` when creating from a Tag. Tags are immutable.

## Retarget an alias

```go
alias, err := collection.Aliases().Retarget(
	ctx,
	"production",
	lambdadb.RetargetAliasInput{
		Target: lambdadb.TagTarget("validated-2026-10"),
	},
)
```

## Select a ref for reads

Set `Ref` on Query, Fetch, List, ListIterator, or ListAll options:

```go
result, err := collection.Query(ctx, lambdadb.QueryInput{
	Query: map[string]any{"queryString": map[string]any{"query": "*:*"}},
	Ref:   lambdadb.AliasRef("production"),
})

fetched, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{
	Ids:            []string{"doc-1"},
	Ref:            lambdadb.BranchRef("candidate"),
	ConsistentRead: lambdadb.Bool(true),
})

docs, err := collection.Docs().ListAll(ctx, &lambdadb.ListDocsOpts{
	Size: lambdadb.Int64(100),
	Ref:  lambdadb.TagRef("validated-2026-09"),
})
```

`ConsistentRead` is valid only when `Ref` directly selects a Branch. Tags are
immutable, while Aliases resolve at request time and may be retargeted. `List`
and iterator pages return `ListDocsDoc` wrappers; `ListAll` returns document
content directly as `[]map[string]any`.

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
Use `WithTransferClient` when presigned uploads or out-of-line result downloads
need a custom proxy, TLS configuration, timeout, or instrumentation.

## Handle errors

Lifecycle operations return the same typed API errors as generated operations:

```go
_, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
	BranchName: "candidate",
})
var conflict *apierrors.ResourceAlreadyExistsError
if errors.As(err, &conflict) {
	// The Branch already exists.
}
```

For response and request field details, see the
[Data Versioning models](../../models/components/versioning.md).
