package lambdadb_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/retry"
)

type publicAPIMockClient struct {
	t        *testing.T
	handlers []func(*http.Request) *http.Response
	calls    int
}

func (m *publicAPIMockClient) Do(req *http.Request) (*http.Response, error) {
	m.t.Helper()
	if m.calls >= len(m.handlers) {
		m.t.Fatalf("unexpected request %d: %s %s", m.calls+1, req.Method, req.URL.String())
	}
	handler := m.handlers[m.calls]
	m.calls++
	return handler(req), nil
}

func (m *publicAPIMockClient) assertDone() {
	m.t.Helper()
	if m.calls != len(m.handlers) {
		m.t.Fatalf("handled %d requests, want %d", m.calls, len(m.handlers))
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestPublicAPI_ListCollectionsFromExternalPackage(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.lambdadb.ai/projects/project-a/collections")
				if got := req.Header.Get("x-api-key"); got != "public-key" {
					t.Fatalf("x-api-key header = %q, want public-key", got)
				}
				if got := req.URL.Query().Get("size"); got != "2" {
					t.Fatalf("size query = %q, want 2", got)
				}
				if got := req.URL.Query().Get("pageToken"); got != "page-1" {
					t.Fatalf("pageToken query = %q, want page-1", got)
				}
				return jsonResponse(http.StatusOK, `{
					"collections": [{
						"projectName": "project-a",
						"collectionName": "articles",
						"indexConfigs": {},
						"numPartitions": 1,
						"numDocs": 3,
						"collectionStatus": "ACTIVE",
						"createdAt": 1700000000,
						"updatedAt": 1700000100,
						"dataUpdatedAt": 1700000200
					}],
					"nextPageToken": "page-2"
				}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.lambdadb.ai"),
		lambdadb.WithProjectName("project-a"),
		lambdadb.WithClient(mock),
	)

	res, err := client.Collections.List(context.Background(), &lambdadb.ListCollectionsOpts{
		Size:      lambdadb.Int64(2),
		PageToken: lambdadb.String("page-1"),
	})
	if err != nil {
		t.Fatalf("Collections.List() error = %v", err)
	}
	if len(res.Collections) != 1 {
		t.Fatalf("len(Collections) = %d, want 1", len(res.Collections))
	}
	if got := res.Collections[0].GetCollectionName(); got != "articles" {
		t.Fatalf("collection name = %q, want articles", got)
	}
	if res.NextPageToken == nil || *res.NextPageToken != "page-2" {
		t.Fatalf("next page token = %v, want page-2", res.NextPageToken)
	}
	mock.assertDone()
}

func TestPublicAPI_OptionsAndHelpersFromExternalPackage(t *testing.T) {
	if got := *lambdadb.String("value"); got != "value" {
		t.Fatalf("String() = %q, want value", got)
	}
	if got := *lambdadb.Bool(true); !got {
		t.Fatalf("Bool() = %v, want true", got)
	}
	if got := *lambdadb.Int(1); got != 1 {
		t.Fatalf("Int() = %d, want 1", got)
	}
	if got := *lambdadb.Int64(2); got != 2 {
		t.Fatalf("Int64() = %d, want 2", got)
	}
	if got := *lambdadb.Float32(1.5); got != 1.5 {
		t.Fatalf("Float32() = %v, want 1.5", got)
	}
	if got := *lambdadb.Float64(2.5); got != 2.5 {
		t.Fatalf("Float64() = %v, want 2.5", got)
	}
	if got := *lambdadb.Pointer("generic"); got != "generic" {
		t.Fatalf("Pointer() = %q, want generic", got)
	}

	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-security/collections")
				if got := req.Header.Get("x-api-key"); got != "source-key" {
					t.Fatalf("x-api-key header = %q, want source-key", got)
				}
				return jsonResponse(http.StatusOK, `{"collections":[]}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-security"),
		lambdadb.WithTimeout(time.Second),
		lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}),
		lambdadb.WithSecuritySource(func(context.Context) (components.Security, error) {
			return components.Security{ProjectAPIKey: lambdadb.String("source-key")}, nil
		}),
		lambdadb.WithClient(mock),
	)

	if _, err := client.Collections.List(context.Background(), nil); err != nil {
		t.Fatalf("Collections.List() with security source error = %v", err)
	}
	mock.assertDone()
}

func TestPublicAPI_CollectionHandleReadFlowsFromExternalPackage(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-b/collections/articles")
				return jsonResponse(http.StatusOK, `{
					"collection": {
						"projectName": "project-b",
						"collectionName": "articles",
						"indexConfigs": {},
						"numPartitions": 1,
						"numDocs": 1,
						"collectionStatus": "ACTIVE",
						"createdAt": 1700000000,
						"updatedAt": 1700000100,
						"dataUpdatedAt": 1700000200
					}
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-b/collections/articles/docs")
				if got := req.URL.Query().Get("size"); got != "10" {
					t.Fatalf("size query = %q, want 10", got)
				}
				return jsonResponse(http.StatusOK, `{
					"total": 1,
					"docs": [{
						"collection": "articles",
						"doc": {"id": "doc-1", "title": "Hello"}
					}],
					"nextPageToken": "docs-page-2",
					"isDocsInline": true
				}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-b"),
		lambdadb.WithClient(mock),
	)

	collection := client.Collection("articles")
	metadata, err := collection.Get(context.Background())
	if err != nil {
		t.Fatalf("Collection.Get() error = %v", err)
	}
	if got := metadata.GetCollectionName(); got != "articles" {
		t.Fatalf("collection name = %q, want articles", got)
	}

	docs, err := collection.Docs().List(context.Background(), &lambdadb.ListDocsOpts{Size: lambdadb.Int64(10)})
	if err != nil {
		t.Fatalf("Docs().List() error = %v", err)
	}
	if docs.Total != 1 || len(docs.Docs) != 1 {
		t.Fatalf("docs result = total %d len %d, want total 1 len 1", docs.Total, len(docs.Docs))
	}
	if got := docs.Docs[0].Doc["id"]; got != "doc-1" {
		t.Fatalf("doc id = %v, want doc-1", got)
	}
	mock.assertDone()
}

func TestPublicAPI_WriteRequestBodiesFromExternalPackage(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-c/collections")
				body := decodeJSONBody(t, req)
				if got := body["collectionName"]; got != "articles" {
					t.Fatalf("collectionName body = %v, want articles", got)
				}
				return jsonResponse(http.StatusAccepted, `{
					"collection": {
						"projectName": "project-c",
						"collectionName": "articles",
						"indexConfigs": {},
						"numPartitions": 1,
						"numDocs": 0,
						"collectionStatus": "CREATING",
						"createdAt": 1700000000,
						"updatedAt": 1700000000,
						"dataUpdatedAt": 1700000000
					}
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-c/collections/articles/docs/upsert")
				body := decodeJSONBody(t, req)
				docs, ok := body["docs"].([]any)
				if !ok || len(docs) != 1 {
					t.Fatalf("docs body = %#v, want one document", body["docs"])
				}
				doc, ok := docs[0].(map[string]any)
				if !ok || doc["id"] != "doc-1" {
					t.Fatalf("first doc = %#v, want id doc-1", docs[0])
				}
				return jsonResponse(http.StatusAccepted, `{"message": "accepted"}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-c"),
		lambdadb.WithClient(mock),
	)

	created, err := client.Collections.Create(context.Background(), lambdadb.CreateCollectionOptions{
		CollectionName: "articles",
	})
	if err != nil {
		t.Fatalf("Collections.Create() error = %v", err)
	}
	if got := created.GetCollectionName(); got != "articles" {
		t.Fatalf("created collection name = %q, want articles", got)
	}

	_, err = client.Collection("articles").Docs().Upsert(context.Background(), lambdadb.UpsertDocsInput{
		Docs: []map[string]any{{"id": "doc-1", "title": "Hello"}},
	})
	if err != nil {
		t.Fatalf("Docs().Upsert() error = %v", err)
	}
	mock.assertDone()
}

func TestPublicAPI_ManagedEmbeddingCollectionConfigFromExternalPackage(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-managed/collections")
				body := decodeJSONBody(t, req)
				indexConfigs, ok := body["indexConfigs"].(map[string]any)
				if !ok {
					t.Fatalf("indexConfigs body = %#v, want object", body["indexConfigs"])
				}
				vectorConfig, ok := indexConfigs["bodyEmbedding"].(map[string]any)
				if !ok {
					t.Fatalf("bodyEmbedding config = %#v, want object", indexConfigs["bodyEmbedding"])
				}
				if got := vectorConfig["type"]; got != "vector" {
					t.Fatalf("bodyEmbedding.type = %v, want vector", got)
				}
				if got := vectorConfig["managedEmbedding"]; got != true {
					t.Fatalf("bodyEmbedding.managedEmbedding = %v, want true", got)
				}
				embedding, ok := vectorConfig["embedding"].(map[string]any)
				if !ok {
					t.Fatalf("bodyEmbedding.embedding = %#v, want object", vectorConfig["embedding"])
				}
				if got := embedding["provider"]; got != "openai" {
					t.Fatalf("embedding.provider = %v, want openai", got)
				}
				if got := embedding["model"]; got != "text-embedding-3-small" {
					t.Fatalf("embedding.model = %v, want text-embedding-3-small", got)
				}
				if got := embedding["sourceField"]; got != "body" {
					t.Fatalf("embedding.sourceField = %v, want body", got)
				}
				if _, ok := embedding["dimensions"]; ok {
					t.Fatalf("embedding.dimensions was sent in create request: %#v", embedding["dimensions"])
				}
				if _, ok := embedding["similarity"]; ok {
					t.Fatalf("embedding.similarity was sent in create request: %#v", embedding["similarity"])
				}
				return jsonResponse(http.StatusAccepted, `{
					"collection": {
						"projectName": "project-managed",
						"collectionName": "semantic-articles",
						"indexConfigs": {
							"body": {
								"type": "text",
								"analyzers": ["english"]
							},
							"bodyEmbedding": {
								"type": "vector",
								"managedEmbedding": true,
								"embedding": {
									"provider": "openai",
									"model": "text-embedding-3-small",
									"sourceField": "body",
									"dimensions": 1536,
									"similarity": "cosine"
								}
							}
						},
						"numPartitions": 1,
						"numDocs": 0,
						"collectionStatus": "CREATING",
						"createdAt": 1700000000,
						"updatedAt": 1700000000,
						"dataUpdatedAt": 1700000000
					}
				}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-managed"),
		lambdadb.WithClient(mock),
	)

	created, err := client.Collections.Create(context.Background(), lambdadb.CreateCollectionOptions{
		CollectionName: "semantic-articles",
		IndexConfigs: map[string]components.IndexConfigsUnion{
			"body": components.CreateIndexConfigsUnionText(components.IndexConfigsText{
				Analyzers: []components.Analyzer{components.AnalyzerEnglish},
			}),
			"bodyEmbedding": components.CreateIndexConfigsUnionManagedEmbeddingVector(components.IndexConfigsManagedEmbeddingVector{
				Embedding: components.EmbeddingConfig{
					Provider:    components.EmbeddingConfigProviderOpenai,
					Model:       "text-embedding-3-small",
					SourceField: "body",
				},
			}),
		},
	})
	if err != nil {
		t.Fatalf("Collections.Create() error = %v", err)
	}
	managed := created.IndexConfigs["bodyEmbedding"].IndexConfigsManagedEmbeddingVector
	if managed == nil {
		t.Fatalf("bodyEmbedding union = %#v, want managed embedding vector", created.IndexConfigs["bodyEmbedding"])
	}
	if got := managed.Embedding.GetDimensions(); got == nil || *got != 1536 {
		t.Fatalf("embedding dimensions = %v, want 1536", got)
	}
	if got := managed.Embedding.GetSimilarity(); got == nil || *got != components.SimilarityCosine {
		t.Fatalf("embedding similarity = %v, want cosine", got)
	}
	mock.assertDone()
}

func TestPublicAPI_CollectionMutationAndQueryFromExternalPackage(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPatch, "https://api.example.com/projects/project-d/collections/articles")
				return jsonResponse(http.StatusOK, collectionResponseBody("project-d", "articles", "ACTIVE"))
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-d/collections/articles/query")
				body := decodeJSONBody(t, req)
				if got := body["query"].(map[string]any)["title"]; got != "hello" {
					t.Fatalf("query title = %v, want hello", got)
				}
				if got := body["size"]; got != float64(1) {
					t.Fatalf("query size = %v, want 1", got)
				}
				return jsonResponse(http.StatusOK, `{
					"took": 3,
					"total": 1,
					"maxScore": 0.9,
					"docs": [{
						"collection": "articles",
						"score": 0.9,
						"doc": {"id": "doc-1"}
					}],
					"isDocsInline": true
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodDelete, "https://api.example.com/projects/project-d/collections/articles")
				return jsonResponse(http.StatusAccepted, `{"message": "accepted"}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-d"),
		lambdadb.WithClient(mock),
	)
	collection := client.Collection("articles")

	updated, err := collection.Update(context.Background(), lambdadb.UpdateCollectionOptions{})
	if err != nil {
		t.Fatalf("Collection.Update() error = %v", err)
	}
	if got := updated.GetCollectionName(); got != "articles" {
		t.Fatalf("updated collection name = %q, want articles", got)
	}

	queryRes, err := collection.Query(context.Background(), lambdadb.QueryInput{
		Size:  lambdadb.Int64(1),
		Query: map[string]any{"title": "hello"},
	})
	if err != nil {
		t.Fatalf("Collection.Query() error = %v", err)
	}
	if queryRes.Total != 1 || len(queryRes.Docs) != 1 {
		t.Fatalf("query result = total %d len %d, want total 1 len 1", queryRes.Total, len(queryRes.Docs))
	}

	if _, err := collection.Delete(context.Background()); err != nil {
		t.Fatalf("Collection.Delete() error = %v", err)
	}
	mock.assertDone()
}

func TestPublicAPI_DocumentMethodsFromExternalPackage(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-e/collections/articles/docs/bulk-upsert")
				return jsonResponse(http.StatusOK, `{
					"url": "https://uploads.example.com/presigned",
					"type": "application/json",
					"httpMethod": "PUT",
					"objectKey": "uploads/articles.json",
					"sizeLimitBytes": 209715200
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-e/collections/articles/docs/bulk-upsert")
				body := decodeJSONBody(t, req)
				if got := body["objectKey"]; got != "uploads/articles.json" {
					t.Fatalf("objectKey body = %v, want uploads/articles.json", got)
				}
				return jsonResponse(http.StatusAccepted, `{"message": "accepted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-e/collections/articles/docs/update")
				assertDocsBody(t, req, "doc-1")
				return jsonResponse(http.StatusAccepted, `{"message": "accepted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-e/collections/articles/docs/delete")
				body := decodeJSONBody(t, req)
				ids, ok := body["ids"].([]any)
				if !ok || len(ids) != 1 || ids[0] != "doc-1" {
					t.Fatalf("ids body = %#v, want [doc-1]", body["ids"])
				}
				return jsonResponse(http.StatusAccepted, `{"message": "accepted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-e/collections/articles/docs/fetch")
				body := decodeJSONBody(t, req)
				ids, ok := body["ids"].([]any)
				if !ok || len(ids) != 1 || ids[0] != "doc-1" {
					t.Fatalf("ids body = %#v, want [doc-1]", body["ids"])
				}
				return jsonResponse(http.StatusOK, `{
					"total": 1,
					"took": 2,
					"docs": [{
						"collection": "articles",
						"doc": {"id": "doc-1"}
					}],
					"isDocsInline": true
				}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-e"),
		lambdadb.WithClient(mock),
	)
	docs := client.Collection("articles").Docs()

	info, err := docs.GetBulkUpsertInfo(context.Background())
	if err != nil {
		t.Fatalf("Docs().GetBulkUpsertInfo() error = %v", err)
	}
	if info.URL != "https://uploads.example.com/presigned" || info.ObjectKey != "uploads/articles.json" {
		t.Fatalf("bulk info = %#v, want URL and object key", info)
	}

	if _, err := docs.BulkUpsert(context.Background(), lambdadb.BulkUpsertInput{ObjectKey: "uploads/articles.json"}); err != nil {
		t.Fatalf("Docs().BulkUpsert() error = %v", err)
	}
	if _, err := docs.Update(context.Background(), lambdadb.UpdateDocsInput{Docs: []map[string]any{{"id": "doc-1"}}}); err != nil {
		t.Fatalf("Docs().Update() error = %v", err)
	}
	if _, err := docs.Delete(context.Background(), lambdadb.DeleteDocsInput{Ids: []string{"doc-1"}}); err != nil {
		t.Fatalf("Docs().Delete() error = %v", err)
	}

	fetched, err := docs.Fetch(context.Background(), lambdadb.FetchDocsInput{Ids: []string{"doc-1"}})
	if err != nil {
		t.Fatalf("Docs().Fetch() error = %v", err)
	}
	if fetched.Total != 1 || len(fetched.Docs) != 1 {
		t.Fatalf("fetch result = total %d len %d, want total 1 len 1", fetched.Total, len(fetched.Docs))
	}
	mock.assertDone()
}

func TestPublicAPI_BulkUpsertDocumentsFromExternalPackage(t *testing.T) {
	var uploadBody map[string]any
	var uploadErr error
	uploadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut {
			uploadErr = fmt.Errorf("upload method = %q, want PUT", req.Method)
			http.Error(w, uploadErr.Error(), http.StatusInternalServerError)
			return
		}
		if req.URL.Path != "/" {
			uploadErr = fmt.Errorf("upload path = %q, want /", req.URL.Path)
			http.Error(w, uploadErr.Error(), http.StatusInternalServerError)
			return
		}
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			uploadErr = fmt.Errorf("upload content-type = %q, want application/json", got)
			http.Error(w, uploadErr.Error(), http.StatusInternalServerError)
			return
		}
		data, err := io.ReadAll(req.Body)
		if err != nil {
			uploadErr = fmt.Errorf("read upload body: %w", err)
			http.Error(w, uploadErr.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(data, &uploadBody); err != nil {
			uploadErr = fmt.Errorf("decode upload body %q: %w", string(data), err)
			http.Error(w, uploadErr.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer uploadServer.Close()

	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-f/collections/articles/docs/bulk-upsert")
				return jsonResponse(http.StatusOK, `{
					"url": "`+uploadServer.URL+`",
					"type": "application/json",
					"httpMethod": "PUT",
					"objectKey": "uploads/articles.json",
					"sizeLimitBytes": 209715200
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-f/collections/articles/docs/bulk-upsert")
				body := decodeJSONBody(t, req)
				if got := body["objectKey"]; got != "uploads/articles.json" {
					t.Fatalf("objectKey body = %v, want uploads/articles.json", got)
				}
				return jsonResponse(http.StatusAccepted, `{"message": "accepted"}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-f"),
		lambdadb.WithClient(mock),
	)

	_, err := client.Collection("articles").Docs().BulkUpsertDocuments(context.Background(), lambdadb.UpsertDocsInput{
		Docs: []map[string]any{{"id": "doc-1"}},
	})
	if uploadErr != nil {
		t.Fatal(uploadErr)
	}
	if err != nil {
		t.Fatalf("Docs().BulkUpsertDocuments() error = %v", err)
	}
	docs, ok := uploadBody["docs"].([]any)
	if !ok || len(docs) != 1 {
		t.Fatalf("upload docs body = %#v, want one document", uploadBody["docs"])
	}
	mock.assertDone()
}

func TestPublicAPI_ListHelpersFromExternalPackage(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-g/collections")
				if got := req.URL.Query().Get("size"); got != "1" {
					t.Fatalf("size query = %q, want 1", got)
				}
				return jsonResponse(http.StatusOK, `{
					"collections": [`+collectionObject("project-g", "one", "ACTIVE")+`],
					"nextPageToken": "next"
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-g/collections")
				if got := req.URL.Query().Get("pageToken"); got != "next" {
					t.Fatalf("pageToken query = %q, want next", got)
				}
				return jsonResponse(http.StatusOK, `{
					"collections": [`+collectionObject("project-g", "two", "ACTIVE")+`]
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-g/collections/articles/docs")
				if got := req.URL.Query().Get("size"); got != "1" {
					t.Fatalf("size query = %q, want 1", got)
				}
				return jsonResponse(http.StatusOK, `{
					"total": 2,
					"docs": [{"collection": "articles", "doc": {"id": "doc-1"}}],
					"nextPageToken": "next-docs",
					"isDocsInline": true
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-g/collections/articles/docs")
				if got := req.URL.Query().Get("pageToken"); got != "next-docs" {
					t.Fatalf("pageToken query = %q, want next-docs", got)
				}
				return jsonResponse(http.StatusOK, `{
					"total": 2,
					"docs": [{"collection": "articles", "doc": {"id": "doc-2"}}],
					"isDocsInline": true
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-g/collections")
				return jsonResponse(http.StatusOK, `{
					"collections": [`+collectionObject("project-g", "iterated", "ACTIVE")+`]
				}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-g/collections/articles/docs")
				return jsonResponse(http.StatusOK, `{
					"total": 1,
					"docs": [{"collection": "articles", "doc": {"id": "iterated-doc"}}],
					"isDocsInline": true
				}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-g"),
		lambdadb.WithClient(mock),
	)

	collections, err := client.Collections.ListAll(context.Background(), &lambdadb.ListCollectionsOpts{Size: lambdadb.Int64(1)})
	if err != nil {
		t.Fatalf("Collections.ListAll() error = %v", err)
	}
	if len(collections) != 2 {
		t.Fatalf("len(collections) = %d, want 2", len(collections))
	}

	docs, err := client.Collection("articles").Docs().ListAll(context.Background(), &lambdadb.ListDocsOpts{Size: lambdadb.Int64(1)})
	if err != nil {
		t.Fatalf("Docs().ListAll() error = %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("len(docs) = %d, want 2", len(docs))
	}

	collectionIterator := client.Collections.ListIterator(context.Background(), nil)
	collectionPage, err := collectionIterator.Next(context.Background())
	if err != nil {
		t.Fatalf("Collections.ListIterator().Next() error = %v", err)
	}
	if len(collectionPage.Collections) != 1 || collectionPage.Collections[0].GetCollectionName() != "iterated" {
		t.Fatalf("collection iterator page = %#v, want iterated collection", collectionPage)
	}
	collectionPage, err = collectionIterator.Next(context.Background())
	if err != nil {
		t.Fatalf("Collections.ListIterator().Next() after done error = %v", err)
	}
	if collectionPage != nil {
		t.Fatalf("collection iterator after done = %#v, want nil", collectionPage)
	}

	docIterator := client.Collection("articles").Docs().ListIterator(context.Background(), nil)
	docPage, err := docIterator.Next(context.Background())
	if err != nil {
		t.Fatalf("Docs().ListIterator().Next() error = %v", err)
	}
	if len(docPage.Docs) != 1 || docPage.Docs[0].Doc["id"] != "iterated-doc" {
		t.Fatalf("doc iterator page = %#v, want iterated-doc", docPage)
	}
	docPage, err = docIterator.Next(context.Background())
	if err != nil {
		t.Fatalf("Docs().ListIterator().Next() after done error = %v", err)
	}
	if docPage != nil {
		t.Fatalf("doc iterator after done = %#v, want nil", docPage)
	}
	mock.assertDone()
}

func assertRequest(t *testing.T, req *http.Request, method string, url string) {
	t.Helper()
	if req.Method != method {
		t.Fatalf("method = %q, want %q", req.Method, method)
	}
	reqURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	if reqURL != url {
		t.Fatalf("url = %q, want %q", reqURL, url)
	}
}

func decodeJSONBody(t *testing.T, req *http.Request) map[string]any {
	t.Helper()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode body %q: %v", string(data), err)
	}
	return body
}

func assertDocsBody(t *testing.T, req *http.Request, id string) {
	t.Helper()
	body := decodeJSONBody(t, req)
	docs, ok := body["docs"].([]any)
	if !ok || len(docs) != 1 {
		t.Fatalf("docs body = %#v, want one document", body["docs"])
	}
	doc, ok := docs[0].(map[string]any)
	if !ok || doc["id"] != id {
		t.Fatalf("first doc = %#v, want id %s", docs[0], id)
	}
}

func collectionResponseBody(projectName string, collectionName string, status string) string {
	return `{"collection": ` + collectionObject(projectName, collectionName, status) + `}`
}

func collectionObject(projectName string, collectionName string, status string) string {
	return `{
		"projectName": "` + projectName + `",
		"collectionName": "` + collectionName + `",
		"indexConfigs": {},
		"numPartitions": 1,
		"numDocs": 0,
		"collectionStatus": "` + status + `",
		"createdAt": 1700000000,
		"updatedAt": 1700000000,
		"dataUpdatedAt": 1700000000
	}`
}
