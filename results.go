package lambdadb

import (
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
)

// QueryResult is the flattened result of a collection query.
type QueryResult struct {
	Docs     []operations.QueryCollectionDoc
	Took     int64
	Total    int64
	MaxScore *float64
}

// FetchResult is the flattened result of fetching documents by ID.
type FetchResult struct {
	Docs  []operations.FetchDocsDoc
	Total int64
	Took  int64
}

// ListCollectionsResult is the flattened result of listing collections (one page).
type ListCollectionsResult struct {
	Collections   []components.CollectionResponse
	NextPageToken *string
}

// ListDocsResult is the flattened result of listing documents (one page).
type ListDocsResult struct {
	Docs          []map[string]any
	Total         int64
	NextPageToken *string
}

// GetBulkUpsertInfoResult is the flattened result of getting bulk upsert upload info.
type GetBulkUpsertInfoResult struct {
	URL            string
	ObjectKey      string
	Type           *operations.Type
	HTTPMethod     *operations.HTTPMethod
	SizeLimitBytes *int64
}
