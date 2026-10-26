package lambdadb_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
)

// Exercise the real HTTP transport's gzip handling and every high-level download
// call site. Documents are generated here rather than maintained as a fixture.
func TestDownloadedDocuments(t *testing.T) {
	for _, compressed := range []bool{false, true} {
		for _, operation := range []string{"query", "fetch", "list", "iterator"} {
			t.Run(fmt.Sprintf("%s/gzip=%t", operation, compressed), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if req.Header.Get("x-api-key") != "" {
						t.Error("API key leaked to document storage")
					}
					w.Header().Set("Content-Type", "application/json")
					var writer io.Writer = w
					if compressed {
						w.Header().Set("Content-Encoding", "gzip")
						zipped := gzip.NewWriter(w)
						defer zipped.Close()
						writer = zipped
					}
					io.WriteString(writer, "[")
					for i := 0; i < 1000; i++ {
						if i > 0 {
							io.WriteString(writer, ",")
						}
						fmt.Fprintf(writer, `{"collection":"articles","score":0.75,"retrievalScore":0.4,"doc":{"id":"doc-%d","article":{"body":"stored text"}}}`, i)
					}
					io.WriteString(writer, "]")
				}))
				defer server.Close()
				envelope := fmt.Sprintf(`{"docs":[],"total":1000,"took":7,"maxScore":0.75,"isDocsInline":false,"docsUrl":%q,"nextPageToken":"next","facets":{"kind":{"buckets":[{"value":"guide","count":1000}]}},"rerank":{"status":"applied","provider":"typesafe","model":"jev-1.13.0","candidateCount":1000,"scoredCount":1000,"took":2}}`, server.URL)
				mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
					func(*http.Request) *http.Response { return jsonResponse(200, envelope) },
				}}
				collection := lambdadb.New(lambdadb.WithClient(mock), lambdadb.WithAPIKey("local-test-key")).Collection("articles")
				check := func(count int, first, last map[string]any) {
					t.Helper()
					if count != 1000 || first["id"] != "doc-0" || last["id"] != "doc-999" || last["article"].(map[string]any)["body"] != "stored text" {
						t.Fatal("download lost documents, order, or stored fields")
					}
				}
				switch operation {
				case "query":
					result, err := collection.Query(context.Background(), lambdadb.QueryInput{})
					if err != nil {
						t.Fatal(err)
					}
					if len(result.Docs) != 1000 {
						t.Fatalf("got %d documents", len(result.Docs))
					}
					check(len(result.Docs), result.Docs[0].Doc, result.Docs[999].Doc)
					if result.Total != 1000 || result.Took != 7 || result.MaxScore == nil || *result.MaxScore != 0.75 || result.Rerank == nil || result.Rerank.ScoredCount != 1000 || result.Facets["kind"].Buckets[0].Count != 1000 || *result.Docs[999].RetrievalScore != 0.4 || *result.Docs[999].Score != 0.75 {
						t.Fatal("query download lost scores or metadata")
					}
				case "fetch":
					result, err := collection.Docs().Fetch(context.Background(), lambdadb.FetchDocsInput{Ids: []string{"doc-0"}})
					if err != nil {
						t.Fatal(err)
					}
					if len(result.Docs) != 1000 {
						t.Fatalf("got %d documents", len(result.Docs))
					}
					check(len(result.Docs), result.Docs[0].Doc, result.Docs[999].Doc)
					if result.Total != 1000 || result.Took != 7 {
						t.Fatal("fetch metadata lost")
					}
				default:
					var count int
					var first, last map[string]any
					if operation == "iterator" {
						result, err := collection.Docs().ListIterator(context.Background(), nil).Next(context.Background())
						if err != nil {
							t.Fatal(err)
						}
						count = len(result.Docs)
						if count != 1000 {
							t.Fatalf("got %d documents", count)
						}
						first, last = result.Docs[0].Doc, result.Docs[999].Doc
						if result.NextPageToken == nil || *result.NextPageToken != "next" {
							t.Fatal("iterator lost page token")
						}
					} else {
						result, err := collection.Docs().List(context.Background(), nil)
						if err != nil {
							t.Fatal(err)
						}
						count = len(result.Docs)
						if count != 1000 {
							t.Fatalf("got %d documents", count)
						}
						first, last = result.Docs[0].Doc, result.Docs[999].Doc
						if result.NextPageToken == nil || *result.NextPageToken != "next" {
							t.Fatal("list lost page token")
						}
					}
					check(count, first, last)
				}
				mock.assertDone()
			})
		}
	}
}

type downloadReadCloser struct {
	io.Reader
	closed bool
}

func (r *downloadReadCloser) Close() error { r.closed = true; return nil }

type downloadFailureReader struct{}

func (downloadFailureReader) Read([]byte) (int, error) {
	return 0, errors.New("injected download failure")
}

func TestDownloadRejectsIncompleteOrTrailingData(t *testing.T) {
	const doc = `{"collection":"articles","doc":{"id":"1"}}`
	var zipped bytes.Buffer
	writer := gzip.NewWriter(&zipped)
	io.WriteString(writer, "["+doc+"]")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	corrupted := append([]byte(nil), zipped.Bytes()...)
	corrupted[len(corrupted)-1] ^= 0xff
	for _, tc := range []struct {
		name, body  string
		readFailure bool
	}{
		{"truncated_array", "[" + doc, false},
		{"truncated_document", "[{", false},
		{"trailing_value", "[" + doc + "] {}", false},
		{"trailing_garbage", "[" + doc + "] junk", false},
		{"object", "{}", false},
		{"transfer_failure", "[" + doc + "]", true},
		{"gzip_checksum", string(corrupted), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader = strings.NewReader(tc.body)
			if tc.readFailure {
				reader = io.MultiReader(reader, downloadFailureReader{})
			}
			if tc.name == "gzip_checksum" {
				gzipReader, err := gzip.NewReader(reader)
				if err != nil {
					t.Fatal(err)
				}
				defer gzipReader.Close()
				reader = gzipReader
			}
			body := &downloadReadCloser{Reader: reader}
			transfer := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(*http.Request) *http.Response { return &http.Response{StatusCode: 200, Body: body} },
			}}
			api := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(*http.Request) *http.Response {
					return jsonResponse(200, `{"docs":[],"isDocsInline":false,"docsUrl":"https://storage.example/docs"}`)
				},
			}}
			result, err := lambdadb.New(lambdadb.WithClient(api), lambdadb.WithTransferClient(transfer)).Collection("articles").Query(context.Background(), lambdadb.QueryInput{})
			if err == nil || result != nil {
				t.Fatalf("corrupted download returned partial success: %v", err)
			}
			if !body.closed {
				t.Fatal("download body was not closed")
			}
			api.assertDone()
			transfer.assertDone()
		})
	}
}

func TestDownloadEmptyAndNullArrays(t *testing.T) {
	for _, payload := range []string{"[] \n", "null \n"} {
		t.Run(strings.TrimSpace(payload), func(t *testing.T) {
			api := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(*http.Request) *http.Response {
					return jsonResponse(200, `{"docs":[],"isDocsInline":false,"docsUrl":"https://storage.example/docs"}`)
				},
				func(*http.Request) *http.Response { return jsonResponse(200, payload) },
			}}
			result, err := lambdadb.New(lambdadb.WithClient(api), lambdadb.WithTransferClient(api)).Collection("articles").Query(context.Background(), lambdadb.QueryInput{})
			if err != nil || result == nil || len(result.Docs) != 0 {
				t.Fatalf("empty array decode: %v", err)
			}
			if (result.Docs == nil) != (strings.TrimSpace(payload) == "null") {
				t.Fatal("empty/null distinction lost")
			}
			api.assertDone()
		})
	}
}
