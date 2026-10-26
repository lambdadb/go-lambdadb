package lambdadb_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/optionalnullable"
	"github.com/lambdadb/go-lambdadb/retry"
)

// Contract: lambdadb/docs@ce2d8f7bd8ee2ab4f240be022085aab37cca0b91.
// These are wire tests, not simulated server validation. Even out-of-contract
// requests must reach the server unchanged and preserve its error classification.
func TestCurrentContractReadLimits(t *testing.T) {
	for _, size := range []int64{101, 1000, 1001} {
		for _, operation := range []string{"query", "bayesian", "rerank", "fetch", "list_get", "list_post"} {
			t.Run(fmt.Sprintf("%s/%d", operation, size), func(t *testing.T) {
				mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
					func(req *http.Request) *http.Response {
						if operation == "list_get" {
							if req.Method != http.MethodGet || req.URL.Query().Get("size") != strconv.FormatInt(size, 10) {
								t.Fatalf("list request: %s %s", req.Method, req.URL)
							}
						} else {
							body := decodeJSONBody(t, req)
							switch operation {
							case "fetch":
								ids := body["ids"].([]any)
								if len(ids) != int(size) || ids[0] != "doc-0" || ids[len(ids)-1] != fmt.Sprintf("doc-%d", size-1) {
									t.Fatal("fetch IDs changed")
								}
							case "bayesian":
								if body["candidateSize"] != float64(size) || body["size"] != float64(10) {
									t.Fatal("Bayesian candidate budget changed")
								}
							case "rerank":
								config := body["rerank"].(map[string]any)
								if config["candidateSize"] != float64(size) || body["size"] != float64(10) || config["fields"].([]any)[0] != "article.body" {
									t.Fatal("rerank budget or unindexed path changed")
								}
							default:
								if body["size"] != float64(size) {
									t.Fatal("size changed")
								}
							}
							if operation == "list_post" && body["ref"].(map[string]any)["name"] != "main" {
								t.Fatal("extended list lost ref")
							}
							if operation == "query" || operation == "bayesian" || operation == "rerank" {
								query := body["query"].(map[string]any)
								if operation == "bayesian" {
									query = query["bayesian"].([]any)[1].(map[string]any)
								}
								if query["knn"].(map[string]any)["k"] != float64(size) {
									t.Fatal("KNN k changed")
								}
							}
						}
						if size == 1001 {
							return jsonResponse(400, `{"message":"server rejected limit"}`)
						}
						return jsonResponse(200, `{"docs":[],"total":0,"took":0,"isDocsInline":true}`)
					},
				}}
				client := lambdadb.New(lambdadb.WithClient(mock), lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}))
				collection := client.Collection("articles")
				var err error
				switch operation {
				case "fetch":
					ids := make([]string, size)
					for i := range ids {
						ids[i] = fmt.Sprintf("doc-%d", i)
					}
					_, err = collection.Docs().Fetch(context.Background(), lambdadb.FetchDocsInput{Ids: ids})
				case "list_get", "list_post":
					options := &lambdadb.ListDocsOpts{Size: lambdadb.Int64(size)}
					if operation == "list_post" {
						options.Ref = lambdadb.BranchRef("main")
					}
					_, err = collection.Docs().List(context.Background(), options)
				default:
					vector := make([]float64, 64)
					vector[0] = 1
					knn := map[string]any{"knn": map[string]any{"field": "vector", "queryVector": vector, "k": size}}
					input := lambdadb.QueryInput{Size: lambdadb.Int64(size), Query: knn}
					if operation == "bayesian" {
						input.Size = lambdadb.Int64(10)
						input.CandidateSize = lambdadb.Int64(size)
						input.Query = map[string]any{"bayesian": []any{map[string]any{"queryString": map[string]any{"query": "*:*"}}, knn}}
					}
					if operation == "rerank" {
						input.Size = lambdadb.Int64(10)
						config := components.RerankConfig{Provider: "typesafe", Model: "jev-1.13.0", QueryText: "restore", Fields: []string{"article.body"}, CandidateSize: lambdadb.Int64(size)}
						input.Rerank = optionalnullable.From(&config)
					}
					_, err = collection.Query(context.Background(), input)
				}
				if size == 1001 {
					var bad *apierrors.BadRequestError
					if !errors.As(err, &bad) {
						t.Fatalf("expected server BadRequestError, got %T", err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				mock.assertDone()
			})
		}
	}
}

func TestCurrentContractReadErrors(t *testing.T) {
	for _, operation := range []string{"query", "fetch", "list_get", "list_post"} {
		for _, status := range []int{400, 503} {
			t.Run(fmt.Sprintf("%s/%d", operation, status), func(t *testing.T) {
				mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
					func(*http.Request) *http.Response {
						return jsonResponse(status, `{"message":"result budget exceeded"}`)
					},
				}}
				client := lambdadb.New(lambdadb.WithClient(mock), lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}))
				collection := client.Collection("articles")
				var err error
				switch operation {
				case "query":
					_, err = collection.Query(context.Background(), lambdadb.QueryInput{})
				case "fetch":
					_, err = collection.Docs().Fetch(context.Background(), lambdadb.FetchDocsInput{Ids: []string{"1"}})
				default:
					options := &lambdadb.ListDocsOpts{}
					if operation == "list_post" {
						options.Ref = lambdadb.BranchRef("main")
					}
					_, err = collection.Docs().List(context.Background(), options)
				}
				if status == 400 {
					var bad *apierrors.BadRequestError
					if !errors.As(err, &bad) {
						t.Fatalf("expected BadRequestError, got %T", err)
					}
				} else {
					var api *apierrors.APIError
					if !errors.As(err, &api) || api.StatusCode != status || api.Body != `{"message":"result budget exceeded"}` {
						t.Fatalf("lost server error context: %v", err)
					}
				}
				mock.assertDone()
			})
		}
	}
}
