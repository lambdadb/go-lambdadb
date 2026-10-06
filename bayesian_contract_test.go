package lambdadb_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/retry"
)

// Contract: lambdadb/lambdadb@9072a1bc8925954369a887f558f1eaf387b7ea0e.
// Queries remain free-form; invalid fusion and budget requests reach the server.
func TestBayesianRequestSerialization(t *testing.T) {
	const signals = `[{"queryString":{"query":"body:restore"}},{"knn":{"field":"vector","queryVector":[1,0],"k":30}}]`
	const rerank = `{"provider":"typesafe","model":"jev-1.13.0","queryText":"restore","fields":["body"]}`
	cases := []struct {
		name, query, extra string
		status             int
	}{
		{"retrieval", `{"bayesian":` + signals + `}`, `,"size":10,"candidateSize":30`, 200},
		{"null_rerank", `{"bayesian":` + signals + `}`, `,"size":10,"candidateSize":30,"rerank":null`, 200},
		{"rerank_default", `{"bayesian":` + signals + `}`, `,"size":10,"rerank":` + rerank, 200},
		{"rerank_explicit", `{"bayesian":` + signals + `}`, `,"size":10,"rerank":{"provider":"typesafe","model":"jev-1.13.0","queryText":"restore","fields":["body"],"candidateSize":30}`, 200},
		{"missing_budget", `{"bayesian":` + signals + `}`, `,"size":10`, 400},
		{"zero_budget", `{"bayesian":` + signals + `}`, `,"candidateSize":0`, 400},
		{"conflicting_budget", `{"bayesian":` + signals + `}`, `,"candidateSize":30,"rerank":` + rerank, 400},
		{"one_signal", `{"bayesian":[{"queryString":{"query":"*:*"}}]}`, `,"candidateSize":30`, 400},
		{"boost_one", `{"bayesian":[{"queryString":{"query":"*:*"},"boost":1},{"queryString":{"query":"body:restore"}}]}`, `,"candidateSize":30`, 400},
	}
	for _, fusion := range []string{"rrf", "mm", "l2", "bayesian"} {
		cases = append(cases, struct {
			name, query, extra string
			status             int
		}{"nested_" + fusion, `{"bayesian":[{"bool":[{"` + fusion + `":` + signals + `}]},{"queryString":{"query":"*:*"}}]}`, `,"candidateSize":30`, 400})
	}
	for _, query := range []string{`{"queryString":{"query":"*:*"}}`, `{"knn":{"field":"vector","queryVector":[1,0]}}`, `{"rrf":` + signals + `}`, `{"mm":` + signals + `}`, `{"l2":` + signals + `}`} {
		cases = append(cases, struct {
			name, query, extra string
			status             int
		}{"ordinary", query, "", 200})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := `{"query":` + tc.query + `,"consistentRead":false,"includeVectors":false` + tc.extra + `}`
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
				if tc.status == 400 {
					return jsonResponse(400, `{"message":"invalid query"}`)
				}
				return jsonResponse(200, `{"took":1,"total":0,"docs":[],"isDocsInline":true}`)
			}}}
			_, err := lambdadb.New(lambdadb.WithClient(mock), lambdadb.WithRetryConfig(retry.Config{Strategy: "none"})).Collection("articles").Query(context.Background(), input)
			if tc.status == 400 {
				var bad *apierrors.BadRequestError
				if !errors.As(err, &bad) {
					t.Fatalf("expected BadRequestError, got %T", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			mock.assertDone()
		})
	}
	var nilInput *lambdadb.QueryInput
	if nilInput.GetCandidateSize() != nil {
		t.Fatal("nil getter")
	}
	value := lambdadb.QueryInput{CandidateSize: lambdadb.Int64(30)}
	if *value.GetCandidateSize() != 30 {
		t.Fatal("candidate getter")
	}
}

func TestNativeEmbeddingCreateUpdateWire(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(map[bool]string{false: "native", true: "legacy"}[legacy], func(t *testing.T) {
			embedding := components.EmbeddingConfig{Provider: components.EmbeddingConfigProviderOpenai, Model: "text-embedding-3-small", SourceField: "body"}
			vector := components.CreateIndexConfigsUnionNativeEmbeddingVector(components.IndexConfigsNativeEmbeddingVector{Embedding: embedding})
			if legacy {
				vector = components.CreateIndexConfigsUnionManagedEmbeddingVector(components.IndexConfigsManagedEmbeddingVector{Embedding: embedding})
			}
			schema := map[string]components.IndexConfigsUnion{"vector": vector}
			handler := func(req *http.Request) *http.Response {
				body := decodeJSONBody(t, req)
				want := map[string]any{"type": "vector", "embedding": map[string]any{"provider": "openai", "model": "text-embedding-3-small", "sourceField": "body"}}
				if legacy {
					want["managedEmbedding"] = true
				}
				if got := body["indexConfigs"].(map[string]any)["vector"]; !reflect.DeepEqual(got, want) {
					t.Fatalf("wire vector = %#v, want %#v", got, want)
				}
				status := 200
				if req.Method == http.MethodPost {
					status = 201
				}
				return jsonResponse(status, `{"collection":{"collectionName":"articles","indexConfigs":{"vector":{"type":"vector","managedEmbedding":true,"embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"body","dimensions":1536,"similarity":"cosine"}}}}}`)
			}
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{handler, handler}}
			client := lambdadb.New(lambdadb.WithClient(mock))
			_, err := client.Collections.Create(context.Background(), lambdadb.CreateCollectionOptions{CollectionName: "articles", IndexConfigs: schema})
			if err != nil {
				t.Fatal(err)
			}
			result, err := client.Collection("articles").Update(context.Background(), lambdadb.UpdateCollectionOptions{IndexConfigs: schema})
			if err != nil {
				t.Fatal(err)
			}
			stored := result.IndexConfigs["vector"].IndexConfigsManagedEmbeddingVector
			if stored == nil || !stored.ManagedEmbedding || stored.Embedding.Dimensions == nil || *stored.Embedding.Dimensions != 1536 {
				t.Fatal("normalized response lost")
			}
			mock.assertDone()
		})
	}
}
