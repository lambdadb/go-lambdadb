package lambdadb_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"github.com/lambdadb/go-lambdadb/optionalnullable"
	"github.com/lambdadb/go-lambdadb/retry"
)

// Contract: lambdadb/lambdadb@55d888299fee44466326a9db8016af9811ade13b,
// RerankConfig/RerankResponse DTOs, RerankContractTest, ManagedRerankService,
// and docs/design/managed-reranking.md; identical to the supplied local checkout
// a5e06d49be06d95dc5f4046aeecaf51f8a7733c0 for these files. The upstream OpenAPI at
// lambdadb/docs@961561c379acb079aec20191e13b89809ef096e9 has no rerank schema.
// These tests verify the SDK wire boundary, not provider execution or deployment.
func TestPublicAPI_RerankRequestCompatibility(t *testing.T) {
	cases := []struct{ name, rerankJSON string }{
		{"legacy", ""},
		{"null", "null"},
		{"default", `{"provider":"typesafe","model":"jev-1.13.0","queryText":"restore","fields":["title","body"]}`},
		{"null_criteria", `{"provider":"typesafe","model":"jev-1.13.0","queryText":"restore","fields":["title","body"],"criteria":null}`},
		{"empty_criteria_preserved_for_server_validation", `{"provider":"typesafe","model":"jev-1.13.0","queryText":"restore","fields":["title","body"],"criteria":[]}`},
	}
	for _, levels := range []int{2, 3, 10} {
		criteria := make([]string, levels)
		for i := range criteria {
			criteria[i] = fmt.Sprintf("Relevance level %d", i)
		}
		raw, err := json.Marshal(criteria)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, struct{ name, rerankJSON string }{
			fmt.Sprintf("custom_%d", levels),
			`{"provider":"typesafe","model":"jev-1.13.0","queryText":"restore","fields":["title","body"],"candidateSize":50,"onFailure":"returnOriginal","criteria":` + string(raw) + `}`,
		})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// candidateSize is independent of knn.k. Neither size nor k is rewritten.
			raw := `{"size":10,"query":{"knn":{"field":"embedding","queryVector":[1,0],"k":20}},"consistentRead":false,"includeVectors":false`
			if tc.rerankJSON != "" {
				raw += `,"rerank":` + tc.rerankJSON
			}
			raw += `}`
			var input lambdadb.QueryInput
			if err := json.Unmarshal([]byte(raw), &input); err != nil {
				t.Fatal(err)
			}
			if input.Rerank.IsSet() != (tc.rerankJSON != "") || input.Rerank.IsNull() != (tc.rerankJSON == "null") {
				t.Fatalf("rerank presence lost: %#v", input.Rerank)
			}
			var want map[string]any
			if err := json.Unmarshal([]byte(raw), &want); err != nil {
				t.Fatal(err)
			}
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					assertRequest(t, req, http.MethodPost, "https://api.lambdadb.ai/projects/playground/collections/articles/query")
					if got := decodeJSONBody(t, req); !reflect.DeepEqual(got, want) {
						t.Fatalf("request = %#v, want %#v", got, want)
					}
					return jsonResponse(200, `{"docs":[],"isDocsInline":true,"total":0,"took":1}`)
				},
			}}
			if _, err := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles").Query(context.Background(), input); err != nil {
				t.Fatal(err)
			}
			mock.assertDone()
		})
	}
}

func TestRerankTypedConstructionAndUnsupportedOptions(t *testing.T) {
	criteria := []string{"Does not meet the requirements.", "Fully meets the requirements."}
	config := components.RerankConfig{Provider: "typesafe", Model: "jev-1.13.0", QueryText: "restore", Fields: []string{"body"}, Criteria: optionalnullable.From(&criteria)}
	input := lambdadb.QueryInput{Rerank: optionalnullable.From(&config)}
	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var decoded lambdadb.QueryInput
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Rerank, input.Rerank) {
		t.Fatalf("typed rerank lost: %#v", decoded.Rerank)
	}
	for _, option := range []string{"weights", "threshold", "rubricVersion"} {
		payload := `{"rerank":{"provider":"typesafe","model":"jev-1.13.0","queryText":"q","fields":["body"],"` + option + `":[]}}`
		if err := json.Unmarshal([]byte(payload), &decoded); err == nil {
			t.Errorf("unsupported option %s was silently discarded", option)
		}
	}
}

func TestPublicAPI_RerankResponsesInlineAndDownloaded(t *testing.T) {
	applied := `{"status":"applied","provider":"typesafe","model":"jev-1.13.0","resolvedModel":"jev-1.13.0","candidateCount":2,"scoredCount":2,"took":12,"criteriaVersion":"default-relevance-v1"}`
	cases := []struct {
		name, docsJSON, metadataJSON string
		max                          *float64
		scores, retrieval            []float64
	}{
		{"default", `[{"collection":"articles","score":0.80000002,"retrievalScore":3.5,"doc":{"id":"second"}},{"collection":"articles","score":0.80000001,"retrievalScore":8.25,"doc":{"id":"first"}}]`, applied, lambdadb.Float64(0.80000002), []float64{0.80000002, 0.80000001}, []float64{3.5, 8.25}},
		{"custom_zero", `[{"collection":"articles","score":0,"retrievalScore":0,"doc":{"id":"second"}},{"collection":"articles","score":0,"retrievalScore":8.25,"doc":{"id":"first"}}]`, `{"status":"applied","provider":"typesafe","model":"jev-1.13.0","candidateCount":2,"scoredCount":2,"took":0,"criteriaVersion":"custom"}`, lambdadb.Float64(0), []float64{0, 0}, []float64{0, 8.25}},
		{"skipped", `[]`, `{"status":"skipped","provider":"typesafe","model":"jev-1.13.0","candidateCount":0,"scoredCount":0,"took":0,"reason":"noCandidates"}`, nil, nil, nil},
		{"fallback", `[{"collection":"articles","score":3.5,"doc":{"id":"second"}},{"collection":"articles","score":8.25,"doc":{"id":"first"}}]`, `{"status":"fallback","provider":"typesafe","model":"jev-1.13.0","candidateCount":2,"scoredCount":0,"took":12,"reason":"timeout"}`, lambdadb.Float64(8.25), []float64{3.5, 8.25}, nil},
		{"legacy", `[{"collection":"articles","score":3.5,"doc":{"id":"second"}},{"collection":"articles","score":8.25,"doc":{"id":"first"}}]`, "", lambdadb.Float64(8.25), []float64{3.5, 8.25}, nil},
	}
	for _, tc := range cases {
		for _, download := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/download=%t", tc.name, download), func(t *testing.T) {
				response := fmt.Sprintf(`{"took":20,"total":%d,"isDocsInline":%t,"facets":{"tags":{"buckets":[{"value":"guide","count":200}]}},"docs":`, len(tc.scores), !download)
				if download {
					response += `[],"docsUrl":"https://files.example/reranked"`
				} else {
					response += tc.docsJSON
				}
				if tc.max != nil {
					raw, _ := json.Marshal(*tc.max)
					response += `,"maxScore":` + string(raw)
				}
				if tc.metadataJSON != "" {
					response += `,"rerank":` + tc.metadataJSON
				}
				response += `}`
				handlers := []func(*http.Request) *http.Response{
					func(req *http.Request) *http.Response { return jsonResponse(200, response) },
				}
				if download {
					handlers = append(handlers, func(req *http.Request) *http.Response {
						assertRequest(t, req, http.MethodGet, "https://files.example/reranked")
						return jsonResponse(200, tc.docsJSON)
					})
				}
				mock := &publicAPIMockClient{t: t, handlers: handlers}
				result, err := lambdadb.New(lambdadb.WithClient(mock), lambdadb.WithTransferClient(mock)).Collection("articles").Query(context.Background(), lambdadb.QueryInput{
					Query:  map[string]any{"queryString": map[string]any{"query": "body:restore"}},
					Facets: map[string]components.FacetRequest{"tags": {}},
				})
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(result.MaxScore, tc.max) || result.Total != int64(len(tc.scores)) || result.Took != 20 {
					t.Fatalf("response values lost: %#v", result)
				}
				var want *components.RerankResponse
				if tc.metadataJSON != "" {
					if err := json.Unmarshal([]byte(tc.metadataJSON), &want); err != nil {
						t.Fatal(err)
					}
				}
				if !reflect.DeepEqual(result.Rerank, want) {
					t.Fatalf("metadata = %#v, want %#v", result.Rerank, want)
				}
				if result.Facets["tags"].Buckets[0].Count != 200 {
					t.Fatal("facet counts lost")
				}
				if len(result.Docs) != len(tc.scores) {
					t.Fatalf("documents lost: %#v", result.Docs)
				}
				for i, doc := range result.Docs {
					if doc.Score == nil || *doc.Score != tc.scores[i] {
						t.Fatalf("score[%d] = %v", i, doc.Score)
					}
					if tc.retrieval == nil {
						if doc.RetrievalScore != nil {
							t.Fatal("unexpected retrievalScore")
						}
					} else if doc.RetrievalScore == nil || *doc.RetrievalScore != tc.retrieval[i] {
						t.Fatalf("retrievalScore[%d] lost", i)
					}
					if doc.Doc["id"] != []string{"second", "first"}[i] || len(doc.Doc) != 1 {
						t.Fatalf("order or projection changed: %#v", doc.Doc)
					}
					raw, err := json.Marshal(doc)
					if err != nil {
						t.Fatal(err)
					}
					var roundtrip operations.QueryCollectionDoc
					if err := json.Unmarshal(raw, &roundtrip); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(doc, roundtrip) {
						t.Fatalf("score precision lost: %s", raw)
					}
				}
				mock.assertDone()
			})
		}
	}
}

func TestPublicAPI_RerankErrorsRemainServerErrors(t *testing.T) {
	for _, status := range []int{400, 408, 429, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			config := components.RerankConfig{
				Provider: "typesafe", Model: "jev-1.13.0", QueryText: "restore", Fields: []string{"body"},
				OnFailure: lambdadb.String("returnOriginal"),
			}
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					return jsonResponse(status, `{"message":"server rejected query"}`)
				},
			}}
			result, err := lambdadb.New(lambdadb.WithClient(mock), lambdadb.WithRetryConfig(retry.Config{Strategy: "none"})).Collection("articles").Query(context.Background(), lambdadb.QueryInput{
				Query:  map[string]any{"queryString": map[string]any{"query": "body:restore"}},
				Rerank: optionalnullable.From(&config),
			})
			if err == nil || result != nil {
				t.Fatalf("server error hidden by client fallback: result=%#v, err=%v", result, err)
			}
			mock.assertDone()
		})
	}
}

func TestRerankMetadataZeroAndOmission(t *testing.T) {
	metadata := components.RerankResponse{Status: "skipped", Provider: "typesafe", Model: "jev-1.13.0", Reason: lambdadb.String("noCandidates")}
	raw, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"candidateCount", "scoredCount", "took"} {
		if value, exists := fields[name]; !exists || value != float64(0) {
			t.Fatalf("zero %s missing: %s", name, raw)
		}
	}
	for _, name := range []string{"resolvedModel", "criteriaVersion", "rubricVersion", "rerankScore"} {
		if _, exists := fields[name]; exists {
			t.Fatalf("unexpected %s: %s", name, raw)
		}
	}
}
