package lambdadb

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_DefaultConfig(t *testing.T) {
	client := New(WithAPIKey("test-key"))
	if client == nil {
		t.Fatal("New() returned nil")
	}
	if client.SDKVersion != Version {
		t.Errorf("SDKVersion = %q, want %q", client.SDKVersion, Version)
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
	res, err := client.Collections.List(ctx, nil)
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
	if len(res.Collections) != 0 {
		t.Errorf("expected empty collections list, got %v", res.Collections)
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
	if res == nil {
		t.Fatal("Get() returned nil")
	}
	wantURL := "https://api.lambdadb.ai/projects/playground/collections/my-coll"
	if rt.req.URL.String() != wantURL {
		t.Errorf("request URL = %q, want %q", rt.req.URL.String(), wantURL)
	}
	if rt.req.Method != http.MethodGet {
		t.Errorf("request Method = %q, want GET", rt.req.Method)
	}
	if res.CollectionName != "my-coll" {
		t.Errorf("collection name = %q", res.CollectionName)
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

func TestDocListIterator_NextAndDone(t *testing.T) {
	callCount := 0
	rt := &mockRoundTripper{
		doFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			// First call: return one page with nextPageToken; second: no next token
			if callCount == 1 {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": {"application/json"}},
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"total":2,"docs":[{"id":"1"}],"nextPageToken":"tok"}`))),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"application/json"}},
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"total":2,"docs":[{"id":"2"}],"nextPageToken":""}`))),
			}, nil
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("p1"),
		WithClient(rt),
	)
	ctx := context.Background()

	it := client.Collection("c1").Docs().ListIterator(ctx, &ListDocsOpts{Size: Int64(10)})
	page1, err := it.Next(ctx)
	if err != nil {
		t.Fatalf("First Next() err = %v", err)
	}
	if page1 == nil {
		t.Fatal("First Next() returned nil page")
	}
	if len(page1.Docs) != 1 || page1.Docs[0]["id"] != "1" {
		t.Errorf("First page docs = %v", page1.Docs)
	}

	page2, err := it.Next(ctx)
	if err != nil {
		t.Fatalf("Second Next() err = %v", err)
	}
	if page2 == nil {
		t.Fatal("Second Next() returned nil page")
	}
	if len(page2.Docs) != 1 || page2.Docs[0]["id"] != "2" {
		t.Errorf("Second page docs = %v", page2.Docs)
	}

	page3, err := it.Next(ctx)
	if err != nil {
		t.Fatalf("Third Next() err = %v", err)
	}
	if page3 != nil {
		t.Errorf("Third Next() should return (nil, nil) when done, got page with %d docs", len(page3.Docs))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestCollectionListIterator_NextAndDone(t *testing.T) {
	callCount := 0
	rt := &mockRoundTripper{
		doFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": {"application/json"}},
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"collections":[{"collectionName":"a"}],"nextPageToken":"t"}`))),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"application/json"}},
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"collections":[{"collectionName":"b"}],"nextPageToken":""}`))),
			}, nil
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("p1"),
		WithClient(rt),
	)
	ctx := context.Background()

	it := client.Collections.ListIterator(ctx, &ListCollectionsOpts{Size: Int64(20)})
	page1, err := it.Next(ctx)
	if err != nil {
		t.Fatalf("First Next() err = %v", err)
	}
	if page1 == nil || len(page1.Collections) != 1 || page1.Collections[0].CollectionName != "a" {
		t.Fatalf("First page = %+v", page1)
	}
	page2, err := it.Next(ctx)
	if err != nil {
		t.Fatalf("Second Next() err = %v", err)
	}
	if page2 == nil || len(page2.Collections) != 1 || page2.Collections[0].CollectionName != "b" {
		t.Fatalf("Second page = %+v", page2)
	}
	page3, err := it.Next(ctx)
	if err != nil {
		t.Fatalf("Third Next() err = %v", err)
	}
	if page3 != nil {
		t.Error("Third Next() should return (nil, nil) when done")
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestNilOption_DoesNotPanic(t *testing.T) {
	rt := &mockRoundTripper{
		resp: &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"collections":[]}`))),
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("p1"),
		WithClient(rt),
	)
	ctx := context.Background()

	// Passing nil as a variadic option (e.g. List(ctx, nil, nil)) must not panic.
	res, err := client.Collections.List(ctx, nil, nil)
	if err != nil {
		t.Fatalf("List(ctx, nil, nil) err = %v", err)
	}
	if res == nil {
		t.Fatal("List returned nil response")
	}
}

func TestBulkUpsertDocuments_Flow(t *testing.T) {
	// Server that receives the PUT upload (presigned URL target).
	uploadReceived := false
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("upload server: expected PUT, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uploadReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	// SDK client mock: GET bulk-upsert -> info with presigned URL; POST bulk-upsert -> success.
	callCount := 0
	rt := &mockRoundTripper{
		doFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				// GetBulkUpsertInfo (GET)
				body := []byte(`{"url":"` + uploadServer.URL + `","type":"application/json","httpMethod":"PUT","objectKey":"test-key","sizeLimitBytes":209715200}`)
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": {"application/json"}},
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil
			}
			// BulkUpsert (POST) - API returns 202 Accepted
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     http.Header{"Content-Type": {"application/json"}},
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"accepted"}`))),
			}, nil
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("p1"),
		WithClient(rt),
	)
	ctx := context.Background()

	body := UpsertDocsInput{Docs: []map[string]any{{"id": "1", "name": "a"}}}
	res, err := client.Collection("my-coll").Docs().BulkUpsertDocuments(ctx, body)
	if err != nil {
		t.Fatalf("BulkUpsertDocuments err = %v", err)
	}
	if res == nil {
		t.Fatal("BulkUpsertDocuments returned nil response")
	}
	if !uploadReceived {
		t.Error("upload server did not receive PUT request")
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (GetBulkUpsertInfo + BulkUpsert), got %d", callCount)
	}
}

func TestQuery_WithDocsURL_FetchesDocsInline(t *testing.T) {
	docsFromURL := []byte(`[{"collection":"c1","score":0.9,"doc":{"id":"1","name":"a"}}]`)
	docsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(docsFromURL)
	}))
	defer docsServer.Close()

	rt := &mockRoundTripper{
		doFunc: func(req *http.Request) (*http.Response, error) {
			// Query API returns isDocsInline=false and docsUrl
			body := []byte(`{"took":1,"total":1,"docs":[],"isDocsInline":false,"docsUrl":"` + docsServer.URL + `"}`)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"application/json"}},
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("p1"),
		WithClient(rt),
	)
	ctx := context.Background()

	res, err := client.Collection("my-coll").Query(ctx, QueryInput{Query: map[string]any{"match_all": map[string]any{}}})
	if err != nil {
		t.Fatalf("Query err = %v", err)
	}
	if res == nil {
		t.Fatal("Query returned nil response")
	}
	if len(res.Docs) != 1 {
		t.Fatalf("expected 1 doc after fetch from URL, got %d", len(res.Docs))
	}
	if res.Docs[0].Doc["id"] != "1" || res.Docs[0].Doc["name"] != "a" {
		t.Errorf("doc = %v", res.Docs[0].Doc)
	}
}

func TestFetch_WithDocsURL_FetchesDocsInline(t *testing.T) {
	docsFromURL := []byte(`[{"collection":"c1","doc":{"id":"x","value":42}}]`)
	docsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(docsFromURL)
	}))
	defer docsServer.Close()

	rt := &mockRoundTripper{
		doFunc: func(req *http.Request) (*http.Response, error) {
			// Fetch API returns isDocsInline=false and docsUrl
			body := []byte(`{"total":1,"took":0,"docs":[],"isDocsInline":false,"docsUrl":"` + docsServer.URL + `"}`)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"application/json"}},
				Body:       io.NopCloser(bytes.NewReader(body)),
			}, nil
		},
	}
	client := New(
		WithAPIKey("key"),
		WithBaseURL("https://api.lambdadb.ai"),
		WithProjectName("p1"),
		WithClient(rt),
	)
	ctx := context.Background()

	res, err := client.Collection("my-coll").Docs().Fetch(ctx, FetchDocsInput{Ids: []string{"x"}})
	if err != nil {
		t.Fatalf("Fetch err = %v", err)
	}
	if res == nil {
		t.Fatal("Fetch returned nil response")
	}
	if len(res.Docs) != 1 {
		t.Fatalf("expected 1 doc after fetch from URL, got %d", len(res.Docs))
	}
	if res.Docs[0].Doc["id"] != "x" || res.Docs[0].Doc["value"] != float64(42) {
		t.Errorf("doc = %v", res.Docs[0].Doc)
	}
}
