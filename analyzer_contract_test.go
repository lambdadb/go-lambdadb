package lambdadb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
)

// Analyzer contract: lambdadb/docs@3bda642f2e7f4f26432f1dfdcb076f656d50f873,
// reference/api/openapi.json, and lambdadb/lambdadb PR #417 head
// 410154abcdf5275add1df47dcf23c170ed0e0efd. These tests verify SDK wire behavior,
// not deployment or server-side acceptance of arbitrary analyzer strings.
var knownAnalyzers = []struct {
	value components.Analyzer
	wire  string
}{
	{components.AnalyzerStandard, "standard"},
	{components.AnalyzerEnglish, "english"},
	{components.AnalyzerKorean, "korean"},
	{components.AnalyzerJapanese, "japanese"},
	{components.AnalyzerChinese, "chinese"},
	{components.AnalyzerCjk, "cjk"},
	{components.AnalyzerArabic, "arabic"},
	{components.AnalyzerFrench, "french"},
	{components.AnalyzerGerman, "german"},
	{components.AnalyzerHindi, "hindi"},
	{components.AnalyzerIndonesian, "indonesian"},
	{components.AnalyzerItalian, "italian"},
	{components.AnalyzerPortuguese, "portuguese"},
	{components.AnalyzerRussian, "russian"},
	{components.AnalyzerSpanish, "spanish"},
	{components.AnalyzerTurkish, "turkish"},
}

func TestAnalyzerIsExact(t *testing.T) {
	for _, tc := range knownAnalyzers {
		t.Run(tc.wire, func(t *testing.T) {
			if string(tc.value) != tc.wire {
				t.Fatalf("constant = %q, want %q", tc.value, tc.wire)
			}
			if !tc.value.IsExact() {
				t.Fatalf("IsExact(%q) = false, want true", tc.value)
			}
		})
	}
	for _, value := range []components.Analyzer{"future_analyzer", "English", "CHINESE", "CJK", "", " chinese "} {
		if value.IsExact() {
			t.Errorf("IsExact(%q) = true, want false", value)
		}
	}
	var missing *components.Analyzer
	if missing.IsExact() {
		t.Error("nil IsExact() = true, want false")
	}
}

func TestPublicAPI_AnalyzerJSONCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		analyzers []components.Analyzer
		fieldJSON string
	}{
		{"omitted", nil, `{"type":"text"}`},
		{"empty", []components.Analyzer{}, `{"type":"text","analyzers":[]}`},
		{"duplicates", []components.Analyzer{components.AnalyzerStandard, components.AnalyzerStandard}, `{"type":"text","analyzers":["standard","standard"]}`},
		{"chinese_and_cjk", []components.Analyzer{components.AnalyzerChinese, components.AnalyzerCjk}, `{"type":"text","analyzers":["chinese","cjk"]}`},
		{"unknown", []components.Analyzer{"future_analyzer"}, `{"type":"text","analyzers":["future_analyzer"]}`},
		{"case_preserved", []components.Analyzer{"English", "CHINESE", "CJK"}, `{"type":"text","analyzers":["English","CHINESE","CJK"]}`},
		{"empty_name", []components.Analyzer{""}, `{"type":"text","analyzers":[""]}`},
	}
	for _, tc := range knownAnalyzers {
		tests = append(tests, struct {
			name      string
			analyzers []components.Analyzer
			fieldJSON string
		}{tc.wire, []components.Analyzer{tc.value}, `{"type":"text","analyzers":["` + tc.wire + `"]}`})
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var wantBody map[string]any
			if err := json.Unmarshal([]byte(`{"collectionName":"articles","indexConfigs":{"content":`+tc.fieldJSON+`}}`), &wantBody); err != nil {
				t.Fatal(err)
			}
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					assertRequest(t, req, http.MethodPost, "https://api.lambdadb.ai/projects/playground/collections")
					if body := decodeJSONBody(t, req); !reflect.DeepEqual(body, wantBody) {
						t.Fatalf("create body = %#v, want %#v", body, wantBody)
					}
					return jsonResponse(http.StatusCreated, `{"collection":{"collectionName":"articles","defaultBranchName":"main"}}`)
				},
				func(req *http.Request) *http.Response {
					assertRequest(t, req, http.MethodGet, "https://api.lambdadb.ai/projects/playground/collections/articles")
					// An independently specified response exercises union deserialization.
					return jsonResponse(http.StatusOK, `{"collection":{"collectionName":"articles","projectName":"playground","indexConfigs":{"content":`+tc.fieldJSON+`},"description":"","tags":{},"numPartitions":1,"numDocs":0,"defaultBranchName":"main","snapshotRetentionInDays":30,"createdAt":1790380800000,"updatedAt":1790380800000}}`)
				},
			}}
			client := lambdadb.New(lambdadb.WithClient(mock))
			_, err := client.Collections.Create(context.Background(), lambdadb.CreateCollectionOptions{
				CollectionName: "articles",
				IndexConfigs: map[string]components.IndexConfigsUnion{
					"content": components.CreateIndexConfigsUnionText(components.IndexConfigsText{
						Type: components.TypeTextText, Analyzers: tc.analyzers,
					}),
				},
			})
			if err != nil {
				t.Fatalf("Collections.Create() error = %v", err)
			}
			response, err := client.Collection("articles").Get(context.Background())
			if err != nil {
				t.Fatalf("Collection.Get() error = %v", err)
			}
			if response == nil {
				t.Fatal("Collection.Get() returned nil")
			}
			field := response.IndexConfigs["content"]
			if field.Type != components.IndexConfigsUnionTypeText || field.IndexConfigsText == nil {
				t.Fatalf("unexpected text union: %#v", field)
			}
			if got := field.IndexConfigsText.Analyzers; !reflect.DeepEqual(got, tc.analyzers) {
				t.Fatalf("decoded analyzers = %#v, want %#v", got, tc.analyzers)
			}
			mock.assertDone()
		})
	}
}
