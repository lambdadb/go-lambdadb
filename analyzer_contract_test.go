package lambdadb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
)

// Analyzer wire shape/default: lambdadb/docs@961561c379acb079aec20191e13b89809ef096e9,
// reference/api/openapi.json (still lists the original 16 names).
// The 49 fixed presets are pinned to lambdadb/lambdadb PR #437 merge
// 55d888299fee44466326a9db8016af9811ade13b, core/IndexingConstants.java.
// These tests verify SDK wire behavior, not deployment or server-side acceptance.
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
	{components.AnalyzerArmenian, "armenian"},
	{components.AnalyzerBasque, "basque"},
	{components.AnalyzerBengali, "bengali"},
	{components.AnalyzerBrazilian, "brazilian"},
	{components.AnalyzerBulgarian, "bulgarian"},
	{components.AnalyzerCatalan, "catalan"},
	{components.AnalyzerCzech, "czech"},
	{components.AnalyzerDanish, "danish"},
	{components.AnalyzerDutch, "dutch"},
	{components.AnalyzerEstonian, "estonian"},
	{components.AnalyzerFinnish, "finnish"},
	{components.AnalyzerGalician, "galician"},
	{components.AnalyzerGreek, "greek"},
	{components.AnalyzerHungarian, "hungarian"},
	{components.AnalyzerIrish, "irish"},
	{components.AnalyzerLatvian, "latvian"},
	{components.AnalyzerLithuanian, "lithuanian"},
	{components.AnalyzerNorwegian, "norwegian"},
	{components.AnalyzerPersian, "persian"},
	{components.AnalyzerRomanian, "romanian"},
	{components.AnalyzerSerbian, "serbian"},
	{components.AnalyzerSorani, "sorani"},
	{components.AnalyzerSwedish, "swedish"},
	{components.AnalyzerThai, "thai"},
	{components.AnalyzerSimple, "simple"},
	{components.AnalyzerWhitespace, "whitespace"},
	{components.AnalyzerStop, "stop"},
	{components.AnalyzerKeyword, "keyword"},
	{components.AnalyzerPattern, "pattern"},
	{components.AnalyzerFingerprint, "fingerprint"},
	{components.AnalyzerNepali, "nepali"},
	{components.AnalyzerTamil, "tamil"},
	{components.AnalyzerTelugu, "telugu"},
}

func TestAnalyzerIsExact(t *testing.T) {
	if len(knownAnalyzers) != 49 {
		t.Fatalf("contract has %d analyzers, want 49", len(knownAnalyzers))
	}
	seen := make(map[string]bool)
	for _, tc := range knownAnalyzers {
		if seen[tc.wire] {
			t.Fatalf("duplicate contract analyzer %q", tc.wire)
		}
		seen[tc.wire] = true
	}
	for _, tc := range knownAnalyzers {
		t.Run(tc.wire, func(t *testing.T) {
			if string(tc.value) != tc.wire {
				t.Fatalf("constant = %q, want %q", tc.value, tc.wire)
			}
			for _, variant := range []components.Analyzer{
				components.Analyzer(strings.ToUpper(tc.wire)),
				components.Analyzer(" " + tc.wire + " "),
			} {
				if variant.IsExact() {
					t.Errorf("IsExact(%q) = true, want false", variant)
				}
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
		{"new_presets_order_duplicates_and_case", []components.Analyzer{components.AnalyzerTelugu, components.AnalyzerKeyword, "Nepali", components.AnalyzerTelugu, components.AnalyzerPattern}, `{"type":"text","analyzers":["telugu","keyword","Nepali","telugu","pattern"]}`},
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
