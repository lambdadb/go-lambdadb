package lambdadb_test

import (
	"context"
	"net/http"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
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
		Query: map[string]any{"matchAll": map[string]any{}},
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
