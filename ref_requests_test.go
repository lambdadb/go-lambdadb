package lambdadb_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/operations"
)

func TestPublicAPI_RefReadsAndBranchWrites(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-refs/collections/articles/docs/list")
				assertRefBody(t, req, "alias", "production")
				return jsonResponse(http.StatusOK, `{"total":0,"docs":[],"isDocsInline":true}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-refs/collections/articles/query")
				assertRefBody(t, req, "tag", "validated-2026-09")
				return jsonResponse(http.StatusOK, `{"took":1,"total":0,"docs":[],"isDocsInline":true}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-refs/collections/articles/docs/fetch")
				body := decodeJSONBody(t, req)
				assertRef(t, body, "branch", "candidate")
				if body["consistentRead"] != true {
					t.Fatalf("consistentRead = %v, want true", body["consistentRead"])
				}
				return jsonResponse(http.StatusOK, `{"took":1,"total":0,"docs":[],"isDocsInline":true}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-refs/collections/articles/docs/upsert")
				assertBranchBody(t, req, "candidate")
				return jsonResponse(http.StatusAccepted, `{"message":"accepted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-refs/collections/articles/docs/update")
				assertBranchBody(t, req, "candidate")
				return jsonResponse(http.StatusAccepted, `{"message":"accepted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-refs/collections/articles/docs/delete")
				assertBranchBody(t, req, "candidate")
				return jsonResponse(http.StatusAccepted, `{"message":"accepted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-refs/collections/articles/docs/bulk-upsert")
				body := decodeJSONBody(t, req)
				if body["branch"] != "candidate" || body["type"] != "application/json" {
					t.Fatalf("bulk upsert body = %#v", body)
				}
				return jsonResponse(http.StatusAccepted, `{"message":"accepted"}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-refs"),
		lambdadb.WithClient(mock),
	)
	docs := client.Collection("articles").Docs()

	if _, err := docs.List(context.Background(), &lambdadb.ListDocsOpts{
		Ref: &lambdadb.RefContext{Kind: lambdadb.RefKindAlias, Name: "production"},
	}); err != nil {
		t.Fatalf("Docs().List() error = %v", err)
	}
	if _, err := client.Collection("articles").Query(context.Background(), lambdadb.QueryInput{
		Query: map[string]any{"queryString": map[string]any{"query": "*:*"}},
		Ref:   &lambdadb.RefContext{Kind: lambdadb.RefKindTag, Name: "validated-2026-09"},
	}); err != nil {
		t.Fatalf("Collection().Query() error = %v", err)
	}
	if _, err := docs.Fetch(context.Background(), lambdadb.FetchDocsInput{
		Ids:            []string{"doc-1"},
		ConsistentRead: lambdadb.Bool(true),
		Ref:            &lambdadb.RefContext{Kind: lambdadb.RefKindBranch, Name: "candidate"},
	}); err != nil {
		t.Fatalf("Docs().Fetch() error = %v", err)
	}
	if _, err := docs.Upsert(context.Background(), lambdadb.UpsertDocsInput{
		Docs:   []map[string]any{{"id": "doc-1"}},
		Branch: lambdadb.String("candidate"),
	}); err != nil {
		t.Fatalf("Docs().Upsert() error = %v", err)
	}
	if _, err := docs.Update(context.Background(), lambdadb.UpdateDocsInput{
		Docs:   []map[string]any{{"id": "doc-1"}},
		Branch: lambdadb.String("candidate"),
	}); err != nil {
		t.Fatalf("Docs().Update() error = %v", err)
	}
	if _, err := docs.Delete(context.Background(), lambdadb.DeleteDocsInput{
		Ids:    []string{"doc-1"},
		Branch: lambdadb.String("candidate"),
	}); err != nil {
		t.Fatalf("Docs().Delete() error = %v", err)
	}
	contentType := operations.TypeApplicationJSON
	if _, err := docs.BulkUpsert(context.Background(), lambdadb.BulkUpsertInput{
		ObjectKey: "uploads/articles.json",
		Type:      &contentType,
		Branch:    lambdadb.String("candidate"),
	}); err != nil {
		t.Fatalf("Docs().BulkUpsert() error = %v", err)
	}
	mock.assertDone()
}

func TestPublicAPI_RefReadErrorSemantics(t *testing.T) {
	type readOperation struct {
		name string
		path string
		call func(context.Context, *lambdadb.Client, *lambdadb.RefContext) error
	}
	operations := []readOperation{
		{
			name: "list",
			path: "https://api.example.com/projects/project-refs/collections/articles/docs/list",
			call: func(ctx context.Context, client *lambdadb.Client, ref *lambdadb.RefContext) error {
				_, err := client.Collection("articles").Docs().List(ctx, &lambdadb.ListDocsOpts{Ref: ref})
				return err
			},
		},
		{
			name: "query",
			path: "https://api.example.com/projects/project-refs/collections/articles/query",
			call: func(ctx context.Context, client *lambdadb.Client, ref *lambdadb.RefContext) error {
				_, err := client.Collection("articles").Query(ctx, lambdadb.QueryInput{
					Query: map[string]any{"queryString": map[string]any{"query": "*:*"}},
					Ref:   ref,
				})
				return err
			},
		},
		{
			name: "fetch",
			path: "https://api.example.com/projects/project-refs/collections/articles/docs/fetch",
			call: func(ctx context.Context, client *lambdadb.Client, ref *lambdadb.RefContext) error {
				_, err := client.Collection("articles").Docs().Fetch(ctx, lambdadb.FetchDocsInput{
					Ids: []string{"doc-1"},
					Ref: ref,
				})
				return err
			},
		},
	}
	errorCases := []struct {
		name       string
		statusCode int
		message    string
		ref        *lambdadb.RefContext
	}{
		{
			name:       "dangling_alias",
			statusCode: http.StatusBadRequest,
			message:    "alias target does not exist",
			ref:        lambdadb.AliasRef("production"),
		},
		{
			name:       "missing_ref",
			statusCode: http.StatusNotFound,
			message:    "ref does not exist",
			ref:        lambdadb.BranchRef("missing"),
		},
	}

	for _, errorCase := range errorCases {
		for _, operation := range operations {
			t.Run(errorCase.name+"/"+operation.name, func(t *testing.T) {
				mock := &publicAPIMockClient{
					t: t,
					handlers: []func(*http.Request) *http.Response{
						func(req *http.Request) *http.Response {
							assertRequest(t, req, http.MethodPost, operation.path)
							assertRefBody(t, req, string(errorCase.ref.Kind), errorCase.ref.Name)
							return jsonResponse(errorCase.statusCode, `{"message":"`+errorCase.message+`"}`)
						},
					},
				}
				client := lambdadb.New(
					lambdadb.WithBaseURL("https://api.example.com"),
					lambdadb.WithProjectName("project-refs"),
					lambdadb.WithClient(mock),
				)

				err := operation.call(context.Background(), client, errorCase.ref)
				assertRefReadError(t, err, errorCase.statusCode, errorCase.message)
				mock.assertDone()
			})
		}
	}
}

func TestPublicAPI_RefIsPreservedAcrossPaginatedDocumentReads(t *testing.T) {
	assertListRequest := func(t *testing.T, req *http.Request, kind, name, pageToken string) {
		t.Helper()
		assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-refs/collections/articles/docs/list")
		body := decodeJSONBody(t, req)
		assertRef(t, body, kind, name)
		if got, _ := body["pageToken"].(string); got != pageToken {
			t.Fatalf("pageToken = %q, want %q", got, pageToken)
		}
	}

	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertListRequest(t, req, "alias", "production", "")
				return jsonResponse(http.StatusOK, `{"total":2,"docs":[{"collection":"articles","doc":{"id":"iterator-1"}}],"nextPageToken":"iterator-next","isDocsInline":true}`)
			},
			func(req *http.Request) *http.Response {
				assertListRequest(t, req, "alias", "production", "iterator-next")
				return jsonResponse(http.StatusOK, `{"total":2,"docs":[{"collection":"articles","doc":{"id":"iterator-2"}}],"isDocsInline":true}`)
			},
			func(req *http.Request) *http.Response {
				assertListRequest(t, req, "branch", "candidate", "")
				return jsonResponse(http.StatusOK, `{"total":2,"docs":[{"collection":"articles","doc":{"id":"all-1"}}],"nextPageToken":"all-next","isDocsInline":true}`)
			},
			func(req *http.Request) *http.Response {
				assertListRequest(t, req, "branch", "candidate", "all-next")
				return jsonResponse(http.StatusOK, `{"total":2,"docs":[{"collection":"articles","doc":{"id":"all-2"}}],"isDocsInline":true}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-refs"),
		lambdadb.WithClient(mock),
	)
	docs := client.Collection("articles").Docs()

	iterator := docs.ListIterator(context.Background(), &lambdadb.ListDocsOpts{
		Ref: &lambdadb.RefContext{Kind: lambdadb.RefKindAlias, Name: "production"},
	})
	for pageNumber := 1; pageNumber <= 2; pageNumber++ {
		page, err := iterator.Next(context.Background())
		if err != nil {
			t.Fatalf("ListIterator().Next() page %d error = %v", pageNumber, err)
		}
		if page == nil || len(page.Docs) != 1 {
			t.Fatalf("ListIterator().Next() page %d = %#v, want one document", pageNumber, page)
		}
	}

	allDocs, err := docs.ListAll(context.Background(), &lambdadb.ListDocsOpts{
		Ref: &lambdadb.RefContext{Kind: lambdadb.RefKindBranch, Name: "candidate"},
	})
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(allDocs) != 2 || allDocs[0]["id"] != "all-1" || allDocs[1]["id"] != "all-2" {
		t.Fatalf("ListAll() docs = %#v, want all-1 and all-2", allDocs)
	}
	mock.assertDone()
}

func assertRefReadError(t *testing.T, err error, statusCode int, message string) {
	t.Helper()
	if err == nil {
		t.Fatal("ref read error = nil")
	}

	switch statusCode {
	case http.StatusBadRequest:
		var target *apierrors.BadRequestError
		if !errors.As(err, &target) {
			t.Fatalf("ref read error = %T %v, want BadRequestError", err, err)
		}
		if target.Message == nil || *target.Message != message {
			t.Fatalf("BadRequestError message = %v, want %q", target.Message, message)
		}
		if target.HTTPMeta.Response == nil || target.HTTPMeta.Response.StatusCode != statusCode {
			t.Fatalf("BadRequestError status = %#v, want %d", target.HTTPMeta.Response, statusCode)
		}
	case http.StatusNotFound:
		var target *apierrors.ResourceNotFoundError
		if !errors.As(err, &target) {
			t.Fatalf("ref read error = %T %v, want ResourceNotFoundError", err, err)
		}
		if target.Message == nil || *target.Message != message {
			t.Fatalf("ResourceNotFoundError message = %v, want %q", target.Message, message)
		}
		if target.HTTPMeta.Response == nil || target.HTTPMeta.Response.StatusCode != statusCode {
			t.Fatalf("ResourceNotFoundError status = %#v, want %d", target.HTTPMeta.Response, statusCode)
		}
	default:
		t.Fatalf("unsupported status code %d", statusCode)
	}
}

func assertRefBody(t *testing.T, req *http.Request, kind, name string) {
	t.Helper()
	assertRef(t, decodeJSONBody(t, req), kind, name)
}

func assertRef(t *testing.T, body map[string]any, kind, name string) {
	t.Helper()
	ref, ok := body["ref"].(map[string]any)
	if !ok || ref["kind"] != kind || ref["name"] != name {
		t.Fatalf("ref body = %#v, want %s/%s", body["ref"], kind, name)
	}
}

func assertBranchBody(t *testing.T, req *http.Request, branch string) {
	t.Helper()
	body := decodeJSONBody(t, req)
	if body["branch"] != branch {
		t.Fatalf("branch body = %v, want %s", body["branch"], branch)
	}
}
