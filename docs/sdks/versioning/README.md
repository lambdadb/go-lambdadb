# Data Versioning

Use collection-scoped Branches, Tags, and Aliases to control document history
and read targets.

## Available operations

| Resource | Operations |
| --- | --- |
| `collection.Branches()` | `Create`, `Get`, `Update`, `List`, `ListPage`, `ListIterator`, `Delete` |
| `collection.Tags()` | `Create`, `Get`, `List`, `ListPage`, `ListIterator`, `Delete` |
| `collection.Aliases()` | `Create`, `Get`, `List`, `ListPage`, `ListIterator`, `Retarget`, `Delete` |

All methods are synchronous and accept `context.Context` and optional
`operations.Option` values. There is no separate async client or method set.
Branch Create/Get/Update return `*BranchDetails`; Tag Create/Get return
`*TagDetails`; Alias Create/Get/Retarget return `*AliasDetails`. `List` returns
every ref as a slice, `ListPage` one page (`*ListBranchesResult`,
`*ListTagsResult`, `*ListAliasesResult`). Delete methods return a
`MessageResponse`.

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

A Branch can be created from another Branch or from a Tag in the same
Collection. `Branches().Create` rejects Alias and unknown source kinds locally
with a validation error before sending an HTTP request.

```go
hotfix, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
	BranchName: "hotfix",
	Source:     lambdadb.TagSource("validated-2026-09"),
})
```

A Branch created from a Tag inherits the Tag's snapshot, its pinned index
schema, and the writes the Tag captured when it was created; later writes to
the Tag's original Branch are excluded. Its `ParentBranch` is nil and its
`ParentSnapshot` is the Tag's snapshot. `AsOf` with a Tag source returns 400.

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
ancestor snapshot selected through `AsOf`. It is nil for `main`, for a Branch
created from a Tag, or when no parent was recorded. This historical identity survives parent deletion or
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

`TagSource(name)` also creates another Tag. It copies the same fixed state
without creating a chain of Tags; `AsOf` is allowed only for Branch sources.
Alias sources are not allowed.

```go
copiedTag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{
	TagName: "validated-copy",
	Source:  lambdadb.TagSource(tag.Name),
})
```

Tags are immutable. Without `AsOf`, a new Branch starts from the source's
committed snapshot and also takes over the writes still pending on the source
at creation. They are not part of its snapshot yet: ordinary reads see them
after the Branch's own build commits them, and `ConsistentRead: true` on the
Branch includes the eligible ones immediately. Writes that reach the source
afterwards do not appear on the new Branch. A Tag created without `AsOf` pins
the source's committed snapshot together with the writes still pending at that
moment; Query and Fetch through the Tag include those captured writes, except
pending bulk imports, which appear after a background build materializes them.
With `AsOf`, a Branch or Tag takes only the latest retained committed snapshot
at or before the cutoff, without pending or inherited writes.

An empty source can produce an empty Branch with a non-nil `ParentBranch` and
nil snapshots; a Tag requires a nonempty committed head.
Unavailable `asOf` history returns 400 for either Branch or Tag creation.
Retention is Collection scoped. Index schemas are Branch scoped; Tag reads
retain their pinned schema.

## Read one ref

`Get` returns one Branch, Tag, or Alias. A name that is not a ref of that kind
returns `ResourceNotFoundError`.

```go
branch, err := collection.Branches().Get(ctx, "candidate")
tag, err := collection.Tags().Get(ctx, "validated-2026-09")
alias, err := collection.Aliases().Get(ctx, "production")
```

Every Branch, Tag, and Alias response (Create, Retarget, Get, Update, and list
entries) carries `IndexConfigs`. Branches and Tags also carry `NumDocs` and
`DataUpdatedAt`:

- A Branch reports its own schema and the statistics of its committed head.
  An empty Branch has `NumDocs` 0 and a nil `DataUpdatedAt`;
  `GetDataUpdatedAt()` then returns the zero time.
- A Tag reports the schema and statistics of the snapshot it pins. Both are
  immutable.
- An Alias reports the schema its readers currently get: the target Branch's
  current schema or the target Tag's pinned schema. A dangling Alias is still
  returned by `Get`, with `Dangling` true and nil `IndexConfigs`.

## Branch index schemas

A Branch owns its index schema. It inherits the schema when it is created (the
source Branch's current schema, the Tag's pinned schema, or with `AsOf` the
schema of the selected snapshot). After that, a schema update changes only the
Branch it targets.

```go
branch, err := collection.Branches().Get(ctx, "candidate")
if err != nil {
	return err
}
schema := branch.IndexConfigs // complete existing schema, including the server-managed id
schema["category"] = components.CreateIndexConfigsUnionKeyword(components.IndexConfigs{})

updated, err := collection.Branches().Update(ctx, "candidate", lambdadb.UpdateBranchInput{
	IndexConfigs: schema,
})
```

`UpdateBranchInput` takes `IndexConfigs` only, required and nonempty. The same
additive rules as a Collection schema update apply within the Branch: send the
Branch's complete existing schema, preserve every existing field and its
settings, and add new fields at any depth. A field `Description` is the only
existing setting that can change. Concurrent additive updates of the same
Branch are merged.

- Updating `main` also updates the `IndexConfigs` shown on the Collection,
  which mirror `main`. `IndexConfigs` on `collection.Update` is deprecated in
  favor of this method.
- Only a Branch can be updated. Naming a Tag or an Alias returns 400; a name
  that does not exist returns 404.
- To bring a schema change from a Branch to `main`, apply the same additive
  change to `main`. There is no automatic promotion.
- An Alias can be retargeted between Branches whose schemas differ; the server
  does not check it. Read the Alias with `Aliases().Get` to see which schema
  its readers currently get.

## List refs

The server returns Branches, Tags, and Aliases one page at a time in name
order (10 entries by default, at most 100). `List` follows every page and
returns all refs. `ListPage` returns one page and `ListIterator` walks pages:

```go
all, err := collection.Branches().List(ctx)

page, err := collection.Tags().ListPage(ctx, &lambdadb.ListRefsOpts{
	Size: lambdadb.Int64(100),
})
if page.NextPageToken != nil {
	page, err = collection.Tags().ListPage(ctx, &lambdadb.ListRefsOpts{
		Size:      lambdadb.Int64(100),
		PageToken: page.NextPageToken,
	})
}

it := collection.Aliases().ListIterator(ctx, nil)
for {
	page, err := it.Next(ctx)
	if err != nil {
		return err
	}
	if page == nil {
		break
	}
	for _, alias := range page.Aliases {
		_ = alias
	}
}
```

`NextPageToken` is nil on the last page. The SDK sends `size` and `pageToken`
only when they are set and never sends an empty `pageToken`; the server
rejects an empty or malformed token, an out-of-range size, and any other query
parameter with 400. Each entry has the same shape as the single-ref response.

## Retarget an alias

Aliases bind to target identity. Deleting a non-default Branch or a Tag returns
409 while any Alias references it. Delete or retarget every referencing Alias
before deleting the target. Deleting an Alias does not delete its target.
If a dangling Alias is encountered, document reads return 400, while
`Aliases().Get` still returns it with nil `IndexConfigs`; recreating the target
name does not repair its old identity binding. Create and Retarget responses
carry the target's `IndexConfigs`, so a retarget shows the schema the Alias now
serves.

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

`ConsistentRead: true` is valid for a directly selected Branch, including the
omitted-ref default `main`, and for an Alias that currently targets a Branch;
the server resolves the Alias to its Branch when the request is admitted. It
overlays eligible inherited and own pending writes on the Branch's committed
snapshot, excludes pending bulk imports until a background build materializes
them, and can return 429 when the eligible pending payload exceeds the limit.
Ordinary Branch reads use only the snapshot. Tag reads and an Alias that
targets a Tag always include the writes the Tag captured at creation and
reject true with 400. Document List does not support consistent reads: every
ref, including Tags and Aliases targeting Tags, lists only the currently
published snapshot.

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
