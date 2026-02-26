package lambdadb

import (
	"context"

	"github.com/lambdadb/go-lambdadb/models/operations"
)

// ProjectCollections provides project-level collection operations.
type ProjectCollections struct {
	client *Client
}

// List returns all collections in the project.
func (p *ProjectCollections) List(ctx context.Context, opts ...operations.Option) (*operations.ListCollectionsResponse, error) {
	return p.client.collections.List(ctx, opts...)
}

// Create creates a new collection.
func (p *ProjectCollections) Create(ctx context.Context, request CreateCollectionOptions, callOpts ...operations.Option) (*operations.CreateCollectionResponse, error) {
	return p.client.collections.Create(ctx, request, callOpts...)
}

// Collection is a handle for a single collection (by name).
// Use it for collection-level operations and document operations without passing the collection name each time.
type Collection struct {
	client *Client
	name   string
}

// Get returns metadata for the collection.
func (c *Collection) Get(ctx context.Context, opts ...operations.Option) (*operations.GetCollectionResponse, error) {
	return c.client.collections.Get(ctx, c.name, opts...)
}

// Update updates the collection configuration.
func (c *Collection) Update(ctx context.Context, body UpdateCollectionOptions, opts ...operations.Option) (*operations.UpdateCollectionResponse, error) {
	return c.client.collections.Update(ctx, c.name, body, opts...)
}

// Delete deletes the collection.
func (c *Collection) Delete(ctx context.Context, opts ...operations.Option) (*operations.DeleteCollectionResponse, error) {
	return c.client.collections.Delete(ctx, c.name, opts...)
}

// Query runs a search query on the collection.
func (c *Collection) Query(ctx context.Context, input QueryInput, opts ...operations.Option) (*operations.QueryCollectionResponse, error) {
	return c.client.collections.Query(ctx, c.name, input, opts...)
}

// Docs returns a handle for document operations on this collection.
func (c *Collection) Docs() *CollectionDocs {
	return &CollectionDocs{client: c.client, name: c.name}
}

// CollectionDocs provides document operations for a single collection.
type CollectionDocs struct {
	client *Client
	name   string
}

// List lists documents in the collection.
// Pass nil for listOpts to use defaults (no size limit, no page token).
func (d *CollectionDocs) List(ctx context.Context, listOpts *ListDocsOpts, opts ...operations.Option) (*operations.ListDocsResponse, error) {
	var size *int64
	var pageToken *string
	if listOpts != nil {
		size = listOpts.Size
		pageToken = listOpts.PageToken
	}
	return d.client.docs.List(ctx, d.name, size, pageToken, opts...)
}

// Upsert upserts documents into the collection.
func (d *CollectionDocs) Upsert(ctx context.Context, body UpsertDocsInput, opts ...operations.Option) (*operations.UpsertDocsResponse, error) {
	return d.client.docs.Upsert(ctx, d.name, body, opts...)
}

// GetBulkUpsertInfo returns info required for bulk upload.
func (d *CollectionDocs) GetBulkUpsertInfo(ctx context.Context, opts ...operations.Option) (*operations.GetBulkUpsertDocsResponse, error) {
	return d.client.docs.GetBulkUpsertInfo(ctx, d.name, opts...)
}

// BulkUpsert bulk upserts documents into the collection.
func (d *CollectionDocs) BulkUpsert(ctx context.Context, body BulkUpsertInput, opts ...operations.Option) (*operations.BulkUpsertDocsResponse, error) {
	return d.client.docs.BulkUpsert(ctx, d.name, body, opts...)
}

// Update updates documents in the collection.
func (d *CollectionDocs) Update(ctx context.Context, body UpdateDocsInput, opts ...operations.Option) (*operations.UpdateDocsResponse, error) {
	return d.client.docs.Update(ctx, d.name, body, opts...)
}

// Delete deletes documents from the collection.
func (d *CollectionDocs) Delete(ctx context.Context, body DeleteDocsInput, opts ...operations.Option) (*operations.DeleteDocsResponse, error) {
	return d.client.docs.Delete(ctx, d.name, body, opts...)
}

// Fetch fetches documents by IDs from the collection.
func (d *CollectionDocs) Fetch(ctx context.Context, body FetchDocsInput, opts ...operations.Option) (*operations.FetchDocsResponse, error) {
	return d.client.docs.Fetch(ctx, d.name, body, opts...)
}
