package lambdadb_test

import (
	"context"
	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"net/http"
	"testing"
)

func TestPublicAPI_FacetsSurviveDocumentDownload(t *testing.T) {
	for _, download := range []bool{false, true} {
		t.Run(map[bool]string{false: "facet_only", true: "download"}[download], func(t *testing.T) {
			size := int64(0)
			if download {
				size = 1
			}
			handlers := []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					body := decodeJSONBody(t, req)
					if body["size"] != float64(size) {
						t.Fatalf("size lost: %#v", body)
					}
					facets, ok := body["facets"].(map[string]any)
					if !ok || facets["tags"].(map[string]any)["size"] != float64(3) {
						t.Fatalf("facets lost: %#v", body)
					}
					suffix := `"isDocsInline":true,"total":0`
					if download {
						suffix = `"isDocsInline":false,"total":1,"docsUrl":"https://files.example/docs"`
					}
					return jsonResponse(200, `{"took":1,"docs":[],"facets":{"tags":{"buckets":[{"value":"한글","count":2147483648}]}},`+suffix+`}`)
				},
			}
			if download {
				handlers = append(handlers, func(req *http.Request) *http.Response {
					return jsonResponse(200, `[{"collection":"items","doc":{"id":"1"}}]`)
				})
			}
			mock := &publicAPIMockClient{t: t, handlers: handlers}
			result, err := lambdadb.New(lambdadb.WithClient(mock), lambdadb.WithTransferClient(mock)).Collection("items").Query(context.Background(), lambdadb.QueryInput{
				Size: &size, Facets: map[string]components.FacetRequest{"tags": {Size: lambdadb.Int64(3)}},
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.Facets["tags"].Buckets[0].Count != 2147483648 || result.Facets["tags"].Buckets[0].Value != "한글" {
				t.Fatalf("facets lost: %#v", result)
			}
			if len(result.Docs) != int(size) {
				t.Fatalf("documents lost: %#v", result)
			}
			mock.assertDone()
		})
	}
}
