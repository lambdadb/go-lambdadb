package lambdadb

import (
	"context"

	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
)

// ProjectCollections provides project-level collection operations.
type ProjectCollections struct {
	client *Client
}

// List returns a page of collections in the project.
// Pass nil for listOpts to use defaults (no size/page token).
func (p *ProjectCollections) List(ctx context.Context, listOpts *ListCollectionsOpts, opts ...operations.Option) (*operations.ListCollectionsResponse, error) {
	return p.client.collections.List(ctx, listOpts, opts...)
}

// Create creates a new collection.
func (p *ProjectCollections) Create(ctx context.Context, request CreateCollectionOptions, callOpts ...operations.Option) (*operations.CreateCollectionResponse, error) {
	return p.client.collections.Create(ctx, request, callOpts...)
}

// CollectionListIterator iterates over collection list pages without managing page tokens manually.
// Create it with ListIterator; then call Next until it returns (nil, nil).
type CollectionListIterator struct {
	collections *Collections
	size       *int64
	nextToken  *string
	done       bool
	callOpts   []operations.Option
}

// ListIterator returns an iterator that fetches collection list pages. Pass nil for listOpts to use defaults.
// Call Next(ctx) in a loop; when it returns (nil, nil), there are no more pages.
func (p *ProjectCollections) ListIterator(ctx context.Context, listOpts *ListCollectionsOpts, opts ...operations.Option) *CollectionListIterator {
	if opts == nil {
		opts = []operations.Option{}
	}
	it := &CollectionListIterator{collections: p.client.collections, callOpts: opts}
	if listOpts != nil {
		it.size = listOpts.Size
		if listOpts.PageToken != nil {
			tok := *listOpts.PageToken
			it.nextToken = &tok
		}
	}
	return it
}

// Next fetches the next page. It returns (response, nil), (nil, error), or (nil, nil) when there are no more pages.
func (it *CollectionListIterator) Next(ctx context.Context) (*operations.ListCollectionsResponse, error) {
	if it.done {
		return nil, nil
	}
	opts := (*ListCollectionsOpts)(nil)
	if it.size != nil || it.nextToken != nil {
		opts = &ListCollectionsOpts{Size: it.size, PageToken: it.nextToken}
	}
	res, err := it.collections.List(ctx, opts, it.callOpts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		it.done = true
		return res, nil
	}
	if res.Object.NextPageToken == nil || *res.Object.NextPageToken == "" {
		it.done = true
	} else {
		it.nextToken = res.Object.NextPageToken
	}
	return res, nil
}

// ListAll fetches all collection pages and returns a single slice. Use with care on projects with many collections.
func (p *ProjectCollections) ListAll(ctx context.Context, listOpts *ListCollectionsOpts, opts ...operations.Option) ([]components.CollectionResponse, error) {
	it := p.ListIterator(ctx, listOpts, opts...)
	var out []components.CollectionResponse
	for {
		page, err := it.Next(ctx)
		if err != nil {
			return nil, err
		}
		if page == nil {
			break
		}
		if page.Object != nil && len(page.Object.Collections) > 0 {
			out = append(out, page.Object.Collections...)
		}
	}
	return out, nil
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

// DocListIterator iterates over document list pages without managing page tokens manually.
// Create it with ListIterator; then call Next until it returns (nil, nil).
// Whether there are more pages is determined only by the API's nextPageToken—the number of
// documents per page may be less than the requested size (e.g. due to payload size limits).
type DocListIterator struct {
	docs     *Docs
	name     string
	size     *int64
	nextToken *string
	done     bool
	callOpts []operations.Option
}

// ListIterator returns an iterator that fetches document list pages. Pass nil for listOpts to use defaults.
// Call Next(ctx) in a loop; when it returns (nil, nil), there are no more pages.
func (d *CollectionDocs) ListIterator(ctx context.Context, listOpts *ListDocsOpts, opts ...operations.Option) *DocListIterator {
	if opts == nil {
		opts = []operations.Option{}
	}
	it := &DocListIterator{docs: d.client.docs, name: d.name, callOpts: opts}
	if listOpts != nil {
		it.size = listOpts.Size
		if listOpts.PageToken != nil {
			tok := *listOpts.PageToken
			it.nextToken = &tok
		}
	}
	return it
}

// Next fetches the next page. It returns (response, nil), (nil, error), or (nil, nil) when there are no more pages.
// When the response is non-nil, use res.Object.Docs and res.Object.NextPageToken as needed.
func (it *DocListIterator) Next(ctx context.Context) (*operations.ListDocsResponse, error) {
	if it.done {
		return nil, nil
	}
	res, err := it.docs.List(ctx, it.name, it.size, it.nextToken, it.callOpts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		it.done = true
		return res, nil
	}
	if res.Object.NextPageToken == nil || *res.Object.NextPageToken == "" {
		it.done = true
	} else {
		it.nextToken = res.Object.NextPageToken
	}
	return res, nil
}

// ListAll fetches all document pages and returns a single slice of docs. Use with care on large collections.
// listOpts.Size is recommended (e.g. 100) to control page size; nil listOpts uses API defaults.
// The API may return fewer docs per page than requested (e.g. due to payload limits); iteration
// continues until the API reports no more pages via nextPageToken.
func (d *CollectionDocs) ListAll(ctx context.Context, listOpts *ListDocsOpts, opts ...operations.Option) ([]map[string]any, error) {
	it := d.ListIterator(ctx, listOpts, opts...)
	var out []map[string]any
	for {
		page, err := it.Next(ctx)
		if err != nil {
			return nil, err
		}
		if page == nil {
			break
		}
		if page.Object != nil && len(page.Object.Docs) > 0 {
			out = append(out, page.Object.Docs...)
		}
	}
	return out, nil
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
