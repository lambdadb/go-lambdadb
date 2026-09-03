package lambdadb

import (
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
)

// ListDocsOpts holds optional parameters for listing documents (request only).
// Pass nil to use defaults.
// Note: isDocsInline and docsUrl are response-only fields from the API; they are not request options.
// When the API returns isDocsInline=false and docsUrl, the SDK automatically fetches the document list from that URL.
type ListDocsOpts struct {
	// Max number of documents to return at once.
	Size *int64
	// Next page token for pagination.
	PageToken *string
	// Include vector values in the response.
	IncludeVectors *bool
	// Filter applied before pagination. When set, the SDK uses the extended list endpoint.
	Filter map[string]any
	// Partition filter applied before pagination. When set, the SDK uses the extended list endpoint.
	PartitionFilter *components.PartitionFilter
	// Field selector. When set, the SDK uses the extended list endpoint.
	Fields *components.FieldsSelectorUnion
	// Collection branch, tag, or alias to read. When set, the SDK uses the
	// extended list endpoint.
	Ref *components.RefContext
}

// ListCollectionsOpts holds optional parameters for listing collections.
// Pass nil to use defaults.
type ListCollectionsOpts struct {
	// Max number of collections to return at once.
	Size *int64
	// Next page token for pagination.
	PageToken *string
}

// Public API type aliases for request/response bodies.
// These map to the underlying operations types for a cleaner public API.

// CreateCollectionOptions configures a new collection (alias of operations.CreateCollectionRequest).
type CreateCollectionOptions = operations.CreateCollectionRequest

// UpdateCollectionOptions configures a collection update (alias of operations.UpdateCollectionRequestBody).
type UpdateCollectionOptions = operations.UpdateCollectionRequestBody

// QueryInput is the query body for collection search (alias of operations.QueryCollectionRequestBody).
type QueryInput = operations.QueryCollectionRequestBody

// UpsertDocsInput is the body for upserting documents (alias of operations.UpsertDocsRequestBody).
type UpsertDocsInput = operations.UpsertDocsRequestBody

// UpdateDocsInput is the body for updating documents (alias of operations.UpdateDocsRequestBody).
type UpdateDocsInput = operations.UpdateDocsRequestBody

// DeleteDocsInput is the body for deleting documents (alias of operations.DeleteDocsRequestBody).
type DeleteDocsInput = operations.DeleteDocsRequestBody

// FetchDocsInput is the body for fetching documents by ID (alias of operations.FetchDocsRequestBody).
type FetchDocsInput = operations.FetchDocsRequestBody

// BulkUpsertInput is the body for bulk upsert (alias of operations.BulkUpsertDocsRequestBody).
type BulkUpsertInput = operations.BulkUpsertDocsRequestBody

// RefContext selects a branch, tag, or alias for a read.
type RefContext = components.RefContext

// RefKind identifies the kind of ref selected for a read.
type RefKind = components.RefKind

// RefSource selects a branch or tag as the source of a new branch or tag.
type RefSource = components.RefSource

// RefSourceKind identifies a branch or tag source.
type RefSourceKind = components.RefSourceKind

// AliasTarget selects a branch or tag for an alias.
type AliasTarget = components.AliasTarget

// RefDetails describes a branch or tag returned by the Data Versioning API.
type RefDetails = components.RefDetails

// AliasDetails describes an alias returned by the Data Versioning API.
type AliasDetails = components.AliasDetails

// AliasTargetKind is the resolved target kind returned by the API.
type AliasTargetKind = components.AliasTargetKind

const (
	RefKindBranch = components.RefKindBranch
	RefKindTag    = components.RefKindTag
	RefKindAlias  = components.RefKindAlias

	RefSourceKindBranch   = components.RefSourceKindBranch
	RefSourceKindTag      = components.RefSourceKindTag
	AliasTargetKindBranch = components.AliasTargetKindBranch
	AliasTargetKindTag    = components.AliasTargetKindTag
)
