package lambdadb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/optionalnullable"
)

// Contract: lambdadb/lambdadb@1c3e7d453cca296a7b2af85a375b4c5016504101 api/openapi.json
// (SHA-256 8a78c2abb9f4169abed31176d9ee86deb1f4782bb2134bee513e52899a9c3aea).
// Native query inputs and Cohere options pass through unchanged for server validation.
func TestCohereNativeQueryAndRerankWire(t *testing.T) {
	const cohereRerank = `{"provider":"cohere","model":"rerank-v5.0-fast","queryText":"red shoes","fields":["title"]}`
	cases := map[string]string{
		"query_inputs":        `{"knn":{"field":"embedding","k":10,"queryInputs":[{"type":"text","value":"red shoes"},{"type":"image","value":"https://example.com/shoe.png"}]}}`,
		"native_query_vector": `{"knn":{"field":"embedding","k":10,"queryVector":[1,0]}}`,
	}
	for name, query := range cases {
		for _, rerank := range []string{"", cohereRerank, `{"provider":"cohere","model":"rerank-v5.0-pro","queryText":"red shoes","fields":["title"],"criteria":null}`} {
			t.Run(name, func(t *testing.T) {
				raw := `{"size":10,"query":` + query + `,"consistentRead":false,"includeVectors":false`
				if rerank != "" {
					raw += `,"rerank":` + rerank
				}
				raw += `}`
				var input lambdadb.QueryInput
				if err := json.Unmarshal([]byte(raw), &input); err != nil {
					t.Fatal(err)
				}
				var want map[string]any
				if err := json.Unmarshal([]byte(raw), &want); err != nil {
					t.Fatal(err)
				}
				mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{func(req *http.Request) *http.Response {
					if got := decodeJSONBody(t, req); !reflect.DeepEqual(got, want) {
						t.Fatalf("wire request = %#v, want %#v", got, want)
					}
					return jsonResponse(200, `{"took":1,"total":0,"docs":[],"isDocsInline":true}`)
				}}}
				if _, err := lambdadb.New(lambdadb.WithClient(mock)).Collection("products").Query(context.Background(), input); err != nil {
					t.Fatal(err)
				}
				mock.assertDone()
			})
		}
	}
}

func TestCohereRerankTypedRequestAndMetadata(t *testing.T) {
	config := components.RerankConfig{Provider: "cohere", Model: "rerank-v5.0-pro", QueryText: "red shoes", Fields: []string{"title"}}
	encoded, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"provider":"cohere","model":"rerank-v5.0-pro","queryText":"red shoes","fields":["title"]}`; string(encoded) != want {
		t.Fatalf("encoded = %s, want %s", encoded, want)
	}
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{func(req *http.Request) *http.Response {
		return jsonResponse(200, `{"took":9,"total":1,"maxScore":0.75,"isDocsInline":true,"docs":[{"collection":"products","score":0.75,"retrievalScore":2.5,"doc":{"id":"p1"}}],"rerank":{"status":"applied","provider":"cohere","model":"rerank-v5.0-pro","candidateCount":1,"scoredCount":1,"took":4,"criteriaVersion":"cohere-relevance-v1"}}`)
	}}}
	result, err := lambdadb.New(lambdadb.WithClient(mock)).Collection("products").Query(context.Background(), lambdadb.QueryInput{
		Size:   lambdadb.Int64(1),
		Query:  map[string]any{"queryString": map[string]any{"query": "title:shoes"}},
		Rerank: optionalnullable.From(&config),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := components.RerankResponse{Status: "applied", Provider: "cohere", Model: "rerank-v5.0-pro", CandidateCount: 1, ScoredCount: 1, Took: 4, CriteriaVersion: lambdadb.String("cohere-relevance-v1")}
	if result.Rerank == nil || !reflect.DeepEqual(*result.Rerank, want) {
		t.Fatalf("rerank metadata = %#v", result.Rerank)
	}
	mock.assertDone()
}

// Backend #515: Get returns the server-managed id field, and requests may send it
// back unchanged. The SDK forwards it, including an id-only schema, without validation.
func TestServerManagedIDIndexConfigPassThrough(t *testing.T) {
	const returned = `{"id":{"type":"keyword"},"title":{"type":"text","analyzers":["standard"]}}`
	const idOnly = `{"id":{"type":"keyword"}}`
	var want, wantIDOnly map[string]any
	if err := json.Unmarshal([]byte(returned), &want); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(idOnly), &wantIDOnly); err != nil {
		t.Fatal(err)
	}
	collection := `{"collection":{"collectionName":"products","indexConfigs":` + returned + `}}`
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response { return jsonResponse(200, collection) },
		func(req *http.Request) *http.Response {
			if got := decodeJSONBody(t, req)["indexConfigs"]; req.Method != http.MethodPatch || !reflect.DeepEqual(got, want) {
				t.Fatalf("update %s indexConfigs = %#v, want %#v", req.Method, got, want)
			}
			return jsonResponse(200, collection)
		},
		func(req *http.Request) *http.Response {
			if got := decodeJSONBody(t, req)["indexConfigs"]; req.Method != http.MethodPost || !reflect.DeepEqual(got, wantIDOnly) {
				t.Fatalf("create %s indexConfigs = %#v, want %#v", req.Method, got, wantIDOnly)
			}
			return jsonResponse(201, `{"collection":{"collectionName":"ids","indexConfigs":`+idOnly+`}}`)
		},
	}}
	client := lambdadb.New(lambdadb.WithClient(mock))
	stored, err := client.Collection("products").Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if id := stored.IndexConfigs["id"].IndexConfigs; id == nil || id.Type != components.TypeKeyword {
		t.Fatalf("decoded id config = %#v", stored.IndexConfigs["id"])
	}
	if _, err := client.Collection("products").Update(context.Background(), lambdadb.UpdateCollectionOptions{IndexConfigs: stored.IndexConfigs}); err != nil {
		t.Fatal(err)
	}
	schema := map[string]components.IndexConfigsUnion{"id": components.CreateIndexConfigsUnionKeyword(components.IndexConfigs{})}
	if _, err := client.Collections.Create(context.Background(), lambdadb.CreateCollectionOptions{CollectionName: "ids", IndexConfigs: schema}); err != nil {
		t.Fatal(err)
	}
	mock.assertDone()
}
