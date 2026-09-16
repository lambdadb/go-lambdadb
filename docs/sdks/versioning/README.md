# Data Versioning

Use collection-scoped Branches, Tags, and Aliases to control document history
and read targets.

## Available operations

| Resource | Operations |
| --- | --- |
| `collection.Branches()` | `Create`, `List`, `Delete` |
| `collection.Tags()` | `Create`, `List`, `Delete` |
| `collection.Aliases()` | `Create`, `List`, `Retarget`, `Delete` |

All methods are synchronous and accept `context.Context` and optional
`operations.Option` values. There is no separate async client or method set.
Branch Create/List return `*BranchDetails`/`[]BranchDetails`; Tag Create/List
return `*TagDetails`/`[]TagDetails`. Alias Create returns `*AliasDetails`.
Delete methods return a `MessageResponse`.

## Create refs

```go
collection := client.Collection("my-collection")

branch, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
	BranchName: "candidate",
	Source:     lambdadb.BranchSource("main"),
})

// First verify the intended committed data on candidate with ConsistentRead: false.
// Tag creation requires a nonempty committed head.
tag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{
	TagName: "validated-2026-09",
	Source:  lambdadb.BranchSource(branch.Name),
})

alias, err := collection.Aliases().Create(ctx, lambdadb.CreateAliasInput{
	AliasName: "production",
	Target:    lambdadb.TagTarget(tag.Name),
})
```

A Branch can be created only from another Branch in the same Collection.
`Branches().Create` rejects Tag, Alias, and unknown source kinds locally with a
validation error before sending an HTTP request.

Omit `Source` to create a Branch or Tag from `main`. Use `BranchSourceAt` to
select the latest committed snapshot at or before a `time.Time` cutoff without
manually converting it to Unix milliseconds:

```go
historical, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
	BranchName: "historical",
	Source:     lambdadb.BranchSourceAt("main", cutoff),
})
```

Branch responses also expose nullable `ParentBranch` with `BranchID` and `Name`.
It records the direct source at creation, including an empty source or an
ancestor snapshot selected through `AsOf`. It is nil for `main` or when no
parent was recorded. This historical identity survives parent deletion or
name reuse and does not prevent deletion.

```go
if parent := branch.GetParentBranch(); parent != nil {
	fmt.Println(parent.BranchID, parent.Name)
}
```

Branch responses expose nullable `HeadSnapshot` and `ParentSnapshot`, each
containing `SnapshotID` and `SnapshotCommittedAt`. The parent is the fixed
creation source, not the previous head. Tag responses expose `SnapshotID` and
`SnapshotCommittedAt` at the top level. Snapshot commit times and `CreatedAt`
are independent epoch-millisecond timestamps.

Use `TagSource(name)` only when creating another Tag. It pins the same snapshot
without creating a chain of Tags; `AsOf` is allowed only for Branch sources.
Alias sources are not allowed.

```go
copiedTag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{
	TagName: "validated-copy",
	Source:  lambdadb.TagSource(tag.Name),
})
```

Tags are immutable. Branch
creation copies committed data only, excluding pending writes. An empty source
can produce an empty Branch with a non-nil `ParentBranch` and nil snapshots; a
Tag requires a nonempty committed head.
Unavailable `asOf` history returns 400 for either Branch or Tag creation.
Schema and retention updates are Collection scoped; Tag reads retain their
pinned schema.

## Retarget an alias

Aliases bind to target identity. Deleting a non-default Branch or a Tag returns
409 while any Alias references it. Delete or retarget every referencing Alias
before deleting the target. Deleting an Alias does not delete its target.
If a dangling Alias is encountered, reads return 400; recreating the target
name does not repair its old identity binding.

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

`ConsistentRead: true` is valid only for a directly selected Branch, including the
omitted-ref default `main`. It overlays eligible pending writes, excludes
pending bulk imports, and can return 429 when the pending payload exceeds the
limit. Tag and Alias reads reject true, even when an Alias targets a Branch.
List does not support consistent reads.

Page tokens are opaque search positions and do not pin a Snapshot. For a stable
export, keep the same immutable Tag and unchanged filters/projection on every
page, as in the `ListAll` example above. Tags are immutable, while Aliases
resolve at request time and may be retargeted. `List`
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

Lifecycle operations preserve server messages and HTTP response metadata.
Their 409 errors use `ResourceAlreadyExistsError`, including conditional
catalog conflicts and deletion of alias-referenced targets. For an in-use
target, delete or retarget all referencing Aliases before retrying; the SDK
does not automatically retry 409. Deleting `main` returns 400.
See [Error Handling](../../../README.md#error-handling) for Gateway status
handling and retry control:

```go
_, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
	BranchName: "candidate",
})
var conflict *apierrors.ResourceAlreadyExistsError
if errors.As(err, &conflict) {
	// Refresh state: this can be a name collision or conditional catalog conflict.
}
```

For response and request field details, see the
[Data Versioning models](../../models/components/versioning.md).
