package lambdadb

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

func TestNew_DefaultConfig(t *testing.T) {
	client := New(WithAPIKey("test-key"))
	if client == nil {
		t.Fatal("New() returned nil")
	}
	if client.SDKVersion != "0.2.0" {
		t.Errorf("SDKVersion = %q, want 0.2.0", client.SDKVersion)
	}
	if client.Collections == nil {
		t.Error("Collections is nil")
	}
	// Default base URL is applied internally; we can't read it without exporting, but we can verify via an API call
}

func TestNew_WithBaseURLAndProjectName(t *testing.T) {
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://custom.api.com"),
		WithProjectName("my-project"),
	)
	if client == nil {
		t.Fatal("New() returned nil")
	}
	// Verify by making a request and checking URL (done in TestCollectionsList_RequestURL)
}

func TestCollection_ReturnsHandleWithName(t *testing.T) {
	client := New(WithAPIKey("key"))
	coll := client.Collection("my-collection")
	if coll == nil {
		t.Fatal("Collection() returned nil")
	}
	if coll.name != "my-collection" {
		t.Errorf("Collection name = %q, want my-collection", coll.name)
	}
	docs := coll.Docs()
	if docs == nil {
		t.Fatal("Docs() returned nil")
	}
	if docs.name != "my-collection" {
		t.Errorf("CollectionDocs name = %q, want my-collection", docs.name)
	}
}

// mockRoundTripper records the last request and returns a configurable response.
type mockRoundTripper struct {
	req     *http.Request
	resp    *http.Response
	doFunc  func(*http.Request) (*http.Response, error)
	doCount int
}

func (m *mockRoundTripper) Do(req *http.Request) (*http.Response, error) {
	m.doCount++
	m.req = req
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return m.resp, nil
}

func TestCollectionsList_RequestURLAndHeaders(t *testing.T) {
	rt := &mockRoundTripper{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"collections":[]}`))),
		},
	}
	client := New(
		WithAPIKey("test-api-key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("playground"),
		WithClient(rt),
	)

	ctx := context.Background()
	res, err := client.Collections.List(ctx)
	if err != nil {
		t.Fatalf("Collections.List() err = %v", err)
	}
	if res == nil {
		t.Fatal("List() returned nil response")
	}
	if rt.req == nil {
		t.Fatal("mock did not receive request")
	}
	if rt.req.URL.String() != "https://api.lambdadb.ai/projects/playground/collections" {
		t.Errorf("request URL = %q, want https://api.lambdadb.ai/projects/playground/collections", rt.req.URL.String())
	}
	if rt.req.Method != http.MethodGet {
		t.Errorf("request Method = %q, want GET", rt.req.Method)
	}
	if key := rt.req.Header.Get("x-api-key"); key != "test-api-key" {
		t.Errorf("x-api-key header = %q, want test-api-key", key)
	}
	if res.Object == nil || len(res.Object.Collections) != 0 {
		t.Errorf("expected empty collections list, got %v", res.Object)
	}
}

func TestCollectionGet_RequestURL(t *testing.T) {
	// Get collection returns 200 with a single collection body
	body := `{"collection":{"projectName":"playground","collectionName":"my-coll","indexConfigs":{},"numPartitions":1,"numDocs":0,"collectionStatus":"ACTIVE"}}`
	rt := &mockRoundTripper{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("playground"),
		WithClient(rt),
	)
	ctx := context.Background()

	res, err := client.Collection("my-coll").Get(ctx)
	if err != nil {
		t.Fatalf("Collection.Get() err = %v", err)
	}
	if res == nil || res.Object == nil {
		t.Fatal("Get() returned nil")
	}
	wantURL := "https://api.lambdadb.ai/projects/playground/collections/my-coll"
	if rt.req.URL.String() != wantURL {
		t.Errorf("request URL = %q, want %q", rt.req.URL.String(), wantURL)
	}
	if rt.req.Method != http.MethodGet {
		t.Errorf("request Method = %q, want GET", rt.req.Method)
	}
	if res.Object.Collection.CollectionName != "my-coll" {
		t.Errorf("collection name = %q", res.Object.Collection.CollectionName)
	}
}

func TestCollectionDocsList_RequestURL(t *testing.T) {
	rt := &mockRoundTripper{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"total":0,"docs":[]}`))),
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("p1"),
		WithClient(rt),
	)
	ctx := context.Background()

	res, err := client.Collection("my-docs").Docs().List(ctx, nil)
	if err != nil {
		t.Fatalf("Docs().List() err = %v", err)
	}
	if res == nil {
		t.Fatal("List() returned nil")
	}
	wantURL := "https://api.lambdadb.ai/projects/p1/collections/my-docs/docs"
	if rt.req.URL.String() != wantURL {
		t.Errorf("request URL = %q, want %q", rt.req.URL.String(), wantURL)
	}
}

func TestListDocsOpts_PassedAsQueryParams(t *testing.T) {
	rt := &mockRoundTripper{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"total":0,"docs":[]}`))),
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("p1"),
		WithClient(rt),
	)
	ctx := context.Background()

	opts := &ListDocsOpts{Size: Int64(10), PageToken: String("token")}
	_, err := client.Collection("c1").Docs().List(ctx, opts)
	if err != nil {
		t.Fatalf("List() err = %v", err)
	}
	q := rt.req.URL.Query()
	if q.Get("size") != "10" {
		t.Errorf("query size = %q, want 10", q.Get("size"))
	}
	if q.Get("pageToken") != "token" {
		t.Errorf("query pageToken = %q, want token", q.Get("pageToken"))
	}
}
