package lambdadb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
)

// ProjectCollections provides project-level collection operations.
type ProjectCollections struct {
	client *Client
}

// List returns a page of collections in the project.
// Pass nil for listOpts to use defaults (no size/page token).
func (p *ProjectCollections) List(ctx context.Context, listOpts *ListCollectionsOpts, opts ...operations.Option) (*ListCollectionsResult, error) {
	res, err := p.client.collections.List(ctx, listOpts, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		return &ListCollectionsResult{Collections: nil, NextPageToken: nil}, nil
	}
	return &ListCollectionsResult{
		Collections:   res.Object.Collections,
		NextPageToken: res.Object.NextPageToken,
	}, nil
}

// Create creates a new collection and returns the created collection metadata.
func (p *ProjectCollections) Create(ctx context.Context, request CreateCollectionOptions, callOpts ...operations.Option) (*components.CollectionResponse, error) {
	res, err := p.client.collections.Create(ctx, request, callOpts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		return nil, nil
	}
	return &res.Object.Collection, nil
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

// Next fetches the next page. It returns (result, nil), (nil, error), or (nil, nil) when there are no more pages.
func (it *CollectionListIterator) Next(ctx context.Context) (*ListCollectionsResult, error) {
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
		return &ListCollectionsResult{Collections: nil, NextPageToken: nil}, nil
	}
	if res.Object.NextPageToken == nil || *res.Object.NextPageToken == "" {
		it.done = true
	} else {
		it.nextToken = res.Object.NextPageToken
	}
	return &ListCollectionsResult{
		Collections:   res.Object.Collections,
		NextPageToken: res.Object.NextPageToken,
	}, nil
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
		if len(page.Collections) > 0 {
			out = append(out, page.Collections...)
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
func (c *Collection) Get(ctx context.Context, opts ...operations.Option) (*components.CollectionResponse, error) {
	res, err := c.client.collections.Get(ctx, c.name, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		return nil, nil
	}
	return &res.Object.Collection, nil
}

// Update updates the collection configuration and returns the updated collection metadata.
func (c *Collection) Update(ctx context.Context, body UpdateCollectionOptions, opts ...operations.Option) (*components.CollectionResponse, error) {
	res, err := c.client.collections.Update(ctx, c.name, body, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		return nil, nil
	}
	return &res.Object.Collection, nil
}

// Delete deletes the collection.
func (c *Collection) Delete(ctx context.Context, opts ...operations.Option) (*operations.DeleteCollectionResponse, error) {
	return c.client.collections.Delete(ctx, c.name, opts...)
}

// Query runs a search query on the collection.
// When the API returns isDocsInline=false and docsUrl, the SDK fetches docs from the presigned URL automatically so result.Docs is always populated.
func (c *Collection) Query(ctx context.Context, input QueryInput, opts ...operations.Option) (*QueryResult, error) {
	res, err := c.client.collections.Query(ctx, c.name, input, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		return &QueryResult{Docs: nil, Took: 0, Total: 0, MaxScore: nil}, nil
	}
	obj := res.Object
	if !obj.IsDocsInline && obj.DocsURL != nil && *obj.DocsURL != "" {
		if err := fetchJSONFromURL(ctx, *obj.DocsURL, &obj.Docs); err != nil {
			return nil, fmt.Errorf("fetch query docs from URL: %w", err)
		}
	}
	return &QueryResult{
		Docs:     obj.Docs,
		Took:     obj.Took,
		Total:    obj.Total,
		MaxScore: obj.MaxScore,
	}, nil
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

// List lists documents in the collection (one page).
// Pass nil for listOpts to use defaults (no size limit, no page token).
func (d *CollectionDocs) List(ctx context.Context, listOpts *ListDocsOpts, opts ...operations.Option) (*ListDocsResult, error) {
	var size *int64
	var pageToken *string
	if listOpts != nil {
		size = listOpts.Size
		pageToken = listOpts.PageToken
	}
	res, err := d.client.docs.List(ctx, d.name, size, pageToken, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		return &ListDocsResult{Docs: nil, Total: 0, NextPageToken: nil}, nil
	}
	return &ListDocsResult{
		Docs:          res.Object.Docs,
		Total:         res.Object.Total,
		NextPageToken: res.Object.NextPageToken,
	}, nil
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

// Next fetches the next page. It returns (result, nil), (nil, error), or (nil, nil) when there are no more pages.
func (it *DocListIterator) Next(ctx context.Context) (*ListDocsResult, error) {
	if it.done {
		return nil, nil
	}
	res, err := it.docs.List(ctx, it.name, it.size, it.nextToken, it.callOpts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		it.done = true
		return &ListDocsResult{Docs: nil, Total: 0, NextPageToken: nil}, nil
	}
	obj := res.Object
	if obj.NextPageToken == nil || *obj.NextPageToken == "" {
		it.done = true
	} else {
		it.nextToken = obj.NextPageToken
	}
	return &ListDocsResult{
		Docs:          obj.Docs,
		Total:         obj.Total,
		NextPageToken: obj.NextPageToken,
	}, nil
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
		if len(page.Docs) > 0 {
			out = append(out, page.Docs...)
		}
	}
	return out, nil
}

// Upsert upserts documents into the collection.
func (d *CollectionDocs) Upsert(ctx context.Context, body UpsertDocsInput, opts ...operations.Option) (*operations.UpsertDocsResponse, error) {
	return d.client.docs.Upsert(ctx, d.name, body, opts...)
}

// GetBulkUpsertInfo returns info required for bulk upload (presigned URL, object key, and optionally sizeLimitBytes).
// When sizeLimitBytes is present, the upload payload must not exceed it (e.g. LambdaDB uses 200MB).
func (d *CollectionDocs) GetBulkUpsertInfo(ctx context.Context, opts ...operations.Option) (*GetBulkUpsertInfoResult, error) {
	res, err := d.client.docs.GetBulkUpsertInfo(ctx, d.name, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		return nil, nil
	}
	o := res.Object
	return &GetBulkUpsertInfoResult{
		URL:            o.URL,
		ObjectKey:      o.ObjectKey,
		Type:           o.Type,
		HTTPMethod:     o.HTTPMethod,
		SizeLimitBytes: o.SizeLimitBytes,
	}, nil
}

// MaxBulkUpsertPayloadBytes is the typical LambdaDB bulk upsert payload limit (200MB). The actual limit is returned by GetBulkUpsertInfo (sizeLimitBytes); use this constant only for reference or when validating before calling the API.
const MaxBulkUpsertPayloadBytes = 200 * 1024 * 1024

// BulkUpsert bulk upserts documents into the collection.
// The uploaded object (via presigned URL) must not exceed the size limit returned by GetBulkUpsertInfo (e.g. 200MB).
func (d *CollectionDocs) BulkUpsert(ctx context.Context, body BulkUpsertInput, opts ...operations.Option) (*operations.BulkUpsertDocsResponse, error) {
	return d.client.docs.BulkUpsert(ctx, d.name, body, opts...)
}

// BulkUpsertDocuments runs the full bulk upsert flow: obtains a presigned URL, uploads the documents as JSON, and completes the bulk upsert.
// It is a convenience over calling GetBulkUpsertInfo, uploading to the presigned URL, and BulkUpsert separately.
// The body format is the same as Upsert (docs array). When the API returns sizeLimitBytes in GetBulkUpsertInfo, payload size is validated against it before uploading.
func (d *CollectionDocs) BulkUpsertDocuments(ctx context.Context, body UpsertDocsInput, opts ...operations.Option) (*operations.BulkUpsertDocsResponse, error) {
	info, err := d.GetBulkUpsertInfo(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("get bulk upsert info: %w", err)
	}
	if info == nil {
		return nil, fmt.Errorf("get bulk upsert info: empty response")
	}

	payload := operations.UpsertDocsRequestBody{Docs: body.Docs}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal documents: %w", err)
	}

	if info.SizeLimitBytes != nil && *info.SizeLimitBytes > 0 && int64(len(jsonBody)) > *info.SizeLimitBytes {
		return nil, fmt.Errorf("upload size %d bytes exceeds limit %d bytes (from API)", len(jsonBody), *info.SizeLimitBytes)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", info.URL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if info.HTTPMethod != nil && string(*info.HTTPMethod) != "" {
		req.Method = string(*info.HTTPMethod)
	}

	uploadClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := uploadClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload to presigned URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return d.client.docs.BulkUpsert(ctx, d.name, operations.BulkUpsertDocsRequestBody{ObjectKey: info.ObjectKey}, opts...)
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
// When the API returns isDocsInline=false and docsUrl, the SDK fetches docs from the presigned URL automatically so result.Docs is always populated.
func (d *CollectionDocs) Fetch(ctx context.Context, body FetchDocsInput, opts ...operations.Option) (*FetchResult, error) {
	res, err := d.client.docs.Fetch(ctx, d.name, body, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Object == nil {
		return &FetchResult{Docs: nil, Total: 0, Took: 0}, nil
	}
	obj := res.Object
	if !obj.IsDocsInline && obj.DocsURL != nil && *obj.DocsURL != "" {
		if err := fetchJSONFromURL(ctx, *obj.DocsURL, &obj.Docs); err != nil {
			return nil, fmt.Errorf("fetch docs from URL: %w", err)
		}
	}
	return &FetchResult{
		Docs:  obj.Docs,
		Total: obj.Total,
		Took:  obj.Took,
	}, nil
}

// fetchJSONFromURL GETs the URL and unmarshals the response body as JSON into v.
func fetchJSONFromURL(ctx context.Context, url string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed: status %d, body: %s", resp.StatusCode, string(body))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
