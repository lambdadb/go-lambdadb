package lambdadb_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"github.com/lambdadb/go-lambdadb/retry"
)

// Contract: lambdadb/docs@1714009a2659a18b26cb68193dd7a071fd6efb42,
// reference/api/openapi.json. These tests exercise the public SDK wire boundary.
func TestPublicAPI_BulkCompletionExplicitType(t *testing.T) {
	for _, contentType := range []*operations.Type{nil, operations.TypeApplicationJSON.ToPointer()} {
		name := "omitted"
		if contentType != nil {
			name = "explicit"
		}
		t.Run(name, func(t *testing.T) {
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					body := decodeJSONBody(t, req)
					want := map[string]any{"objectKey": "uploads/docs.json", "type": "application/json", "branch": "candidate"}
					if !reflect.DeepEqual(body, want) {
						t.Fatalf("completion body = %#v, want %#v", body, want)
					}
					return jsonResponse(http.StatusAccepted, `{"message":"accepted"}`)
				},
			}}
			input := lambdadb.BulkUpsertInput{ObjectKey: "uploads/docs.json", Type: contentType, Branch: lambdadb.String("candidate")}
			client := lambdadb.New(lambdadb.WithClient(mock))
			if _, err := client.Collection("articles").Docs().BulkUpsert(context.Background(), input); err != nil {
				t.Fatal(err)
			}
			if input.Type != contentType {
				t.Fatal("caller input was modified")
			}
			mock.assertDone()
		})
	}
}

func TestPublicAPI_CollectionPatchSemantics(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input lambdadb.UpdateCollectionOptions
		want  map[string]any
	}{
		{"omit", lambdadb.UpdateCollectionOptions{SnapshotRetentionInDays: lambdadb.Int64(31)}, map[string]any{"snapshotRetentionInDays": float64(31)}},
		{"clear", lambdadb.UpdateCollectionOptions{Description: lambdadb.String(""), Tags: map[string]string{}}, map[string]any{"description": "", "tags": map[string]any{}}},
		{"replace", lambdadb.UpdateCollectionOptions{Tags: map[string]string{"env": "test"}}, map[string]any{"tags": map[string]any{"env": "test"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					if req.Method != http.MethodPatch {
						t.Fatalf("method = %s, want PATCH", req.Method)
					}
					if body := decodeJSONBody(t, req); !reflect.DeepEqual(body, tc.want) {
						t.Fatalf("PATCH body = %#v, want %#v", body, tc.want)
					}
					return jsonResponse(http.StatusOK, `{"collection":{"collectionName":"articles","projectName":"playground","indexConfigs":{"title":{"type":"text"}},"description":"","snapshotRetentionInDays":30,"numPartitions":1,"defaultBranchName":"main","numDocs":0,"tags":{},"createdAt":1788825600000,"updatedAt":1788825600000}}`)
				},
			}}
			client := lambdadb.New(lambdadb.WithClient(mock))
			result, err := client.Collection("articles").Update(context.Background(), tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if result == nil || !result.GetDataUpdatedAt().IsZero() {
				t.Fatalf("absent dataUpdatedAt should remain zero: %#v", result)
			}
			mock.assertDone()
		})
	}
}

func TestPublicAPI_GatewayErrorContext(t *testing.T) {
	for _, op := range []struct {
		name string
		call func(*lambdadb.Client) error
	}{
		{"collection", func(c *lambdadb.Client) error {
			_, err := c.Collection("articles").Update(context.Background(), lambdadb.UpdateCollectionOptions{Description: lambdadb.String("updated")})
			return err
		}},
		{"document", func(c *lambdadb.Client) error {
			_, err := c.Collection("articles").Docs().BulkUpsert(context.Background(), lambdadb.BulkUpsertInput{ObjectKey: "uploads/docs.json"})
			return err
		}},
		{"ref", func(c *lambdadb.Client) error {
			_, err := c.Collection("articles").Aliases().Retarget(context.Background(), "serving", lambdadb.RetargetAliasInput{Target: lambdadb.TagTarget("release-001")})
			return err
		}},
	} {
		for _, status := range []int{409, 413, 429, 502, 503, 504} {
			if op.name == "document" && status == 409 {
				continue // Not a documented bulk completion response.
			}
			t.Run(op.name+"/"+http.StatusText(status), func(t *testing.T) {
				const message = "server context"
				response := jsonResponse(status, `{"message":"`+message+`"}`)
				if status == 429 {
					response.Header.Set("Retry-After", "7")
				}
				mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
					func(req *http.Request) *http.Response { return response },
				}}
				client := lambdadb.New(lambdadb.WithClient(mock), lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}))
				err := op.call(client)
				if err == nil || !strings.Contains(err.Error(), message) {
					t.Fatalf("error lost server message: %v", err)
				}
				var raw *http.Response
				switch e := err.(type) {
				case *apierrors.APIError:
					raw = e.RawResponse
					if e.StatusCode != status || !strings.Contains(e.Body, message) {
						t.Fatalf("API error lost status/body: %#v", e)
					}
				case *apierrors.ResourceAlreadyExistsError:
					raw = e.HTTPMeta.Response
				case *apierrors.TooManyRequestsError:
					raw = e.HTTPMeta.Response
				case *apierrors.InternalServerError:
					raw = e.HTTPMeta.Response
				default:
					t.Fatalf("unexpected error type %T", err)
				}
				if raw != response || raw.StatusCode != status {
					t.Fatal("error lost response metadata")
				}
				if status == 429 && raw.Header.Get("Retry-After") != "7" {
					t.Fatal("error lost Retry-After header")
				}
				mock.assertDone()
			})
		}
	}
}

func TestPublicAPI_BulkUploadPreconditionFailure(t *testing.T) {
	api := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"url":"https://storage.example.com/upload","objectKey":"uploads/docs.json","type":"application/json","httpMethod":"PUT","headers":{"If-None-Match":"*"}}`)
		},
	}}
	transfer := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			if req.Method != http.MethodPut || req.Header.Get("If-None-Match") != "*" || req.Header.Get("x-api-key") != "" {
				t.Fatalf("incorrect signed upload request: %s %#v", req.Method, req.Header)
			}
			response := jsonResponse(http.StatusPreconditionFailed, `<Error><Code>PreconditionFailed</Code></Error>`)
			response.Header.Set("Content-Type", "application/xml")
			return response
		},
	}}
	client := lambdadb.New(lambdadb.WithClient(api), lambdadb.WithTransferClient(transfer), lambdadb.WithAPIKey("test-key"))
	_, err := client.Collection("articles").Docs().BulkUpsertDocuments(context.Background(), lambdadb.UpsertDocsInput{Docs: []map[string]any{{"id": "doc-1"}}})
	var apiErr *apierrors.APIError
	if err == nil || !strings.Contains(err.Error(), "412") || !strings.Contains(err.Error(), "PreconditionFailed") || errors.As(err, &apiErr) {
		t.Fatalf("expected storage error with status and body, got %T %v", err, err)
	}
	api.assertDone()      // No completion request after an unsuccessful upload.
	transfer.assertDone() // No repeat PUT to the create-only URL.
}

func TestPublicAPI_NestedSchemaUpdate(t *testing.T) {
	// Full schema retains profile.name and profile.address.country while adding
	// profile.city, profile.address.postcode, and a top-level category field.
	const schema = `{"title":{"type":"text","analyzers":["standard"]},"profile":{"type":"object","objectIndexConfigs":{"name":{"type":"keyword"},"city":{"type":"keyword"},"address":{"type":"object","objectIndexConfigs":{"country":{"type":"keyword"},"postcode":{"type":"keyword"}}}}},"category":{"type":"keyword"}}`
	var input lambdadb.UpdateCollectionOptions
	if err := json.Unmarshal([]byte(`{"indexConfigs":`+schema+`}`), &input); err != nil {
		t.Fatal(err)
	}
	var want map[string]any
	if err := json.Unmarshal([]byte(`{"indexConfigs":`+schema+`}`), &want); err != nil {
		t.Fatal(err)
	}
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			if req.Method != http.MethodPatch {
				t.Fatalf("method = %s", req.Method)
			}
			if got := decodeJSONBody(t, req); !reflect.DeepEqual(got, want) {
				t.Fatalf("schema changed in transit: %#v", got)
			}
			return jsonResponse(http.StatusOK, `{"collection":{"collectionName":"articles","indexConfigs":`+schema+`}}`)
		},
	}}
	if _, err := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles").Update(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	mock.assertDone()
}

// consistentRead: true is valid for a direct branch or an alias that currently targets a
// branch; tag refs and tag-targeting aliases are rejected by the server. The SDK adds no
// local ref validation, so every combination is forwarded unchanged.
func TestPublicAPI_ConsistentReadRefConditions(t *testing.T) {
	for _, operation := range []string{"query", "fetch"} {
		for _, ref := range []*lambdadb.RefContext{nil, lambdadb.BranchRef("candidate"), lambdadb.TagRef("release"), lambdadb.AliasRef("branch-alias"), lambdadb.AliasRef("tag-alias")} {
			for _, consistent := range []*bool{nil, lambdadb.Bool(false), lambdadb.Bool(true)} {
				name := operation + "/default"
				if ref != nil {
					name = operation + "/" + string(ref.Kind) + "/" + ref.Name
				}
				mode := "omitted"
				if consistent != nil {
					mode = strconv.FormatBool(*consistent)
				}
				t.Run(name+"/"+mode, func(t *testing.T) {
					wantTrue := consistent != nil && *consistent
					rejected := wantTrue && ref != nil && (ref.Kind == lambdadb.RefKindTag || ref.Name == "tag-alias")
					mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
						func(req *http.Request) *http.Response {
							body := decodeJSONBody(t, req)
							if got, _ := body["consistentRead"].(bool); got != wantTrue {
								t.Fatalf("consistentRead = %v, want %v", body["consistentRead"], wantTrue)
							}
							if ref == nil {
								if _, exists := body["ref"]; exists {
									t.Fatalf("unexpected default ref: %#v", body)
								}
							} else {
								assertRef(t, body, string(ref.Kind), ref.Name)
							}
							if rejected {
								return jsonResponse(http.StatusBadRequest, `{"message":"consistentRead requires a branch or an alias that targets a branch"}`)
							}
							return jsonResponse(http.StatusOK, `{"took":1,"total":0,"docs":[],"isDocsInline":true}`)
						},
					}}
					collection := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles")
					var err error
					if operation == "query" {
						_, err = collection.Query(context.Background(), lambdadb.QueryInput{Query: map[string]any{"queryString": map[string]any{"query": "*:*"}}, Ref: ref, ConsistentRead: consistent})
					} else {
						_, err = collection.Docs().Fetch(context.Background(), lambdadb.FetchDocsInput{Ids: []string{"doc-1"}, Ref: ref, ConsistentRead: consistent})
					}
					if rejected {
						assertRefReadError(t, err, http.StatusBadRequest, "consistentRead requires a branch or an alias that targets a branch")
					} else if err != nil {
						t.Fatal(err)
					}
					mock.assertDone()
				})
			}
		}
	}
}

// Contract: lambdadb/docs@9a1f563b9b53bc29e5278105f8fc8c25a17e0cb8. Sort and
// ConsistentRead select the extended list endpoint and are sent only when set;
// an unset value adds no key, so existing extended list bodies are unchanged.
func TestPublicAPI_ListDocsSortAndConsistentRead(t *testing.T) {
	const listURL = "https://api.example.com/projects/project-list/collections/articles/docs/list"
	for _, tc := range []struct {
		name     string
		listOpts *lambdadb.ListDocsOpts
		want     map[string]any // nil means the GET endpoint
	}{
		{"neither", &lambdadb.ListDocsOpts{Size: lambdadb.Int64(10)}, nil},
		{"sort", &lambdadb.ListDocsOpts{Sort: []map[string]any{{"category": "asc"}, {"publishedAt": "desc"}}},
			map[string]any{"sort": []any{map[string]any{"category": "asc"}, map[string]any{"publishedAt": "desc"}}, "includeVectors": false}},
		{"empty_sort", &lambdadb.ListDocsOpts{Sort: []map[string]any{}}, map[string]any{"sort": []any{}, "includeVectors": false}},
		{"consistent_read", &lambdadb.ListDocsOpts{ConsistentRead: lambdadb.Bool(true)}, map[string]any{"consistentRead": true, "includeVectors": false}},
		{"consistent_read_false", &lambdadb.ListDocsOpts{ConsistentRead: lambdadb.Bool(false)}, map[string]any{"consistentRead": false, "includeVectors": false}},
		{"ref_without_new_fields", &lambdadb.ListDocsOpts{Ref: lambdadb.BranchRef("candidate")},
			map[string]any{"ref": map[string]any{"kind": "branch", "name": "candidate"}, "includeVectors": false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					if tc.want == nil {
						assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-list/collections/articles/docs")
					} else {
						assertRequest(t, req, http.MethodPost, listURL)
						if body := decodeJSONBody(t, req); !reflect.DeepEqual(body, tc.want) {
							t.Fatalf("body = %#v, want %#v", body, tc.want)
						}
					}
					return jsonResponse(http.StatusOK, `{"total":0,"docs":[],"isDocsInline":true}`)
				},
			}}
			client := lambdadb.New(lambdadb.WithBaseURL("https://api.example.com"), lambdadb.WithProjectName("project-list"), lambdadb.WithClient(mock))
			if _, err := client.Collection("articles").Docs().List(context.Background(), tc.listOpts); err != nil {
				t.Fatal(err)
			}
			mock.assertDone()
		})
	}
}

// A page token is a position in the requested order, so every page repeats the
// original request body, including Sort and ConsistentRead, with the token.
func TestPublicAPI_ListDocsIteratorRepeatsSortAndConsistentRead(t *testing.T) {
	sortOrder := []any{map[string]any{"category": "asc"}}
	wantTokens := []any{nil, "page-2"}
	nextTokens := []string{`,"nextPageToken":"page-2"`, ""}
	var handlers []func(*http.Request) *http.Response
	for i := range wantTokens {
		handlers = append(handlers, func(req *http.Request) *http.Response {
			want := map[string]any{"size": float64(10), "sort": sortOrder, "consistentRead": true, "includeVectors": false,
				"ref": map[string]any{"kind": "alias", "name": "serving"}}
			if wantTokens[i] != nil {
				want["pageToken"] = wantTokens[i]
			}
			if body := decodeJSONBody(t, req); !reflect.DeepEqual(body, want) {
				t.Fatalf("page %d body = %#v, want %#v", i+1, body, want)
			}
			return jsonResponse(http.StatusOK, `{"total":2,"docs":[{"collection":"articles","doc":{"id":"doc-`+strconv.Itoa(i+1)+`"}}],"isDocsInline":true`+nextTokens[i]+`}`)
		})
	}
	mock := &publicAPIMockClient{t: t, handlers: handlers}
	docs, err := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles").Docs().ListAll(context.Background(), &lambdadb.ListDocsOpts{
		Size:           lambdadb.Int64(10),
		Sort:           []map[string]any{{"category": "asc"}},
		ConsistentRead: lambdadb.Bool(true),
		Ref:            lambdadb.AliasRef("serving"),
	})
	if err != nil || len(docs) != 2 || docs[0]["id"] != "doc-1" || docs[1]["id"] != "doc-2" {
		t.Fatalf("ListAll() = %#v, %v", docs, err)
	}
	mock.assertDone()
}

// consistentRead: true on a tag or tag alias, a changed request condition and an
// incompatible sort are rejected by the server; the SDK forwards them unchanged
// and preserves the 400.
func TestPublicAPI_ListDocsServerRejections(t *testing.T) {
	for _, tc := range []struct {
		name     string
		listOpts *lambdadb.ListDocsOpts
	}{
		{"consistent_read_tag", &lambdadb.ListDocsOpts{Ref: lambdadb.TagRef("release"), ConsistentRead: lambdadb.Bool(true)}},
		{"consistent_read_tag_alias", &lambdadb.ListDocsOpts{Ref: lambdadb.AliasRef("tag-alias"), ConsistentRead: lambdadb.Bool(true)}},
		{"changed_condition", &lambdadb.ListDocsOpts{Sort: []map[string]any{{"category": "desc"}}, PageToken: lambdadb.String("token-for-asc")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					body := decodeJSONBody(t, req)
					if tc.listOpts.Ref != nil {
						assertRef(t, body, string(tc.listOpts.Ref.Kind), tc.listOpts.Ref.Name)
						if body["consistentRead"] != true {
							t.Fatalf("consistentRead = %v, want true", body["consistentRead"])
						}
					} else if body["pageToken"] != "token-for-asc" {
						t.Fatalf("pageToken = %v", body["pageToken"])
					}
					return jsonResponse(http.StatusBadRequest, `{"message":"invalid list request"}`)
				},
			}}
			_, err := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles").Docs().List(context.Background(), tc.listOpts)
			assertRefReadError(t, err, http.StatusBadRequest, "invalid list request")
			mock.assertDone()
		})
	}
}
