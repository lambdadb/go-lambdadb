package lambdadb

import "github.com/lambdadb/go-lambdadb/models/operations"

// ListDocsOpts holds optional parameters for listing documents.
// Pass nil to use defaults.
type ListDocsOpts struct {
	// Max number of documents to return at once.
	Size *int64
	// Next page token for pagination.
	PageToken *string
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
