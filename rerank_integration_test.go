package lambdadb_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/optionalnullable"
	"github.com/lambdadb/go-lambdadb/retry"
)

// This opt-in test creates and deletes one temporary collection and invokes the
// server-managed provider five times. It checks SDK contracts, not ranking quality,
// load, injected provider failures, usage accounting, or production availability.
func TestIntegrationManagedRerankingSmoke(t *testing.T) {
	if os.Getenv("LAMBDADB_RUN_RERANK_SMOKE") != "1" {
		t.Skip("set LAMBDADB_RUN_RERANK_SMOKE=1 to run the live smoke test")
	}
	client := lambdadb.New(
		lambdadb.WithBaseURL(normalizeIntegrationBaseURL(requireIntegrationEnv(t, "LAMBDADB_BASE_URL"))),
		lambdadb.WithProjectName(requireIntegrationEnv(t, "LAMBDADB_PROJECT_NAME")),
		lambdadb.WithAPIKey(requireIntegrationEnv(t, "LAMBDADB_PROJECT_API_KEY")),
		lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	name := fmt.Sprintf("go-sdk-rerank-%d", time.Now().UnixNano())
	schema := map[string]components.IndexConfigsUnion{
		"body": components.CreateIndexConfigsUnionText(components.IndexConfigsText{}),
	}
	collection := client.Collection(name)
	if _, err := client.Collections.Create(ctx, lambdadb.CreateCollectionOptions{CollectionName: name, IndexConfigs: schema}); err != nil {
		t.Fatalf("create temporary collection: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, err := collection.Delete(cleanupCtx); err != nil {
			t.Errorf("delete temporary collection %s: %v", name, err)
		} else {
			t.Log("temporary collection deleted")
		}
	})
	docs := []map[string]any{
		{"id": "restore", "body": "Restore a previous collection version using a branch created from a tag."},
		{"id": "search", "body": "Search a collection using a query string and vector search."},
		{"id": "weather", "body": "A weather forecast describes rain and temperature."},
	}
	if _, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{Docs: docs}); err != nil {
		t.Fatal(err)
	}
	selector := components.CreateFieldsSelectorUnionFieldsSelector1(components.FieldsSelector1{Include: []string{"id"}})
	input := lambdadb.QueryInput{Size: lambdadb.Int64(3), Query: map[string]any{"queryString": map[string]any{"query": "*:*"}}, ConsistentRead: lambdadb.Bool(true), Fields: &selector}
	var baseline *lambdadb.QueryResult
	waitForIntegrationCondition(t, ctx, "rerank baseline readiness", func() (bool, error) {
		result, err := collection.Query(ctx, input)
		var apiError *apierrors.APIError
		if errors.As(err, &apiError) && apiError.StatusCode == 503 {
			return false, nil // Wait for query readiness before any paid stage.
		}
		if err != nil {
			return false, err
		}
		baseline = result
		return result != nil && len(result.Docs) == 3, nil
	})
	if baseline.Rerank != nil {
		t.Fatal("legacy query returned rerank metadata")
	}
	original := map[string]float64{}
	for _, doc := range baseline.Docs {
		if doc.Score == nil || doc.RetrievalScore != nil {
			t.Fatal("legacy score contract changed")
		}
		original[doc.Doc["id"].(string)] = *doc.Score
	}
	input.Rerank = optionalnullable.From[components.RerankConfig](nil)
	legacyNull, err := collection.Query(ctx, input)
	if err != nil || legacyNull.Rerank != nil || len(legacyNull.Docs) != 3 {
		t.Fatalf("null rerank: %v", err)
	}
	for _, doc := range legacyNull.Docs {
		if doc.RetrievalScore != nil || doc.Score == nil || *doc.Score != original[doc.Doc["id"].(string)] {
			t.Fatal("null rerank changed retrieval scores")
		}
	}
	cases := []struct {
		name     string
		criteria optionalnullable.OptionalNullable[[]string]
	}{{"default", nil}, {"null_criteria", optionalnullable.From[[]string](nil)}}
	for _, levels := range []int{2, 3, 10} {
		criteria := make([]string, levels)
		for i := range criteria {
			criteria[i] = fmt.Sprintf("Relevance level %d of %d: from unrelated to a complete explanation of collection restoration.", i+1, levels)
		}
		cases = append(cases, struct {
			name     string
			criteria optionalnullable.OptionalNullable[[]string]
		}{fmt.Sprintf("custom%d", levels), optionalnullable.From(&criteria)})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := components.RerankConfig{Provider: "typesafe", Model: "jev-1.13.0", QueryText: "How do I restore a previous collection version?", Fields: []string{"body"}, Criteria: tc.criteria}
			request := input
			request.Size = lambdadb.Int64(2)
			request.Rerank = optionalnullable.From(&config)
			result, err := collection.Query(ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			r := result.Rerank
			version := "default-relevance-v1"
			if tc.criteria.IsSet() && !tc.criteria.IsNull() {
				version = "custom"
			}
			if r == nil || r.Status != "applied" || r.Provider != "typesafe" || r.Model != "jev-1.13.0" || r.CandidateCount != 3 || r.ScoredCount != 3 || r.CriteriaVersion == nil || *r.CriteriaVersion != version || r.Reason != nil {
				t.Fatalf("applied metadata mismatch: %#v", r)
			}
			if len(result.Docs) != 2 || result.Total != 2 || result.MaxScore == nil {
				t.Fatal("final size or maxScore missing")
			}
			previous := math.Inf(1)
			for _, doc := range result.Docs {
				id, ok := doc.Doc["id"].(string)
				if !ok || len(doc.Doc) != 1 {
					t.Fatal("public projection changed")
				}
				retrieval, exists := original[id]
				if !exists || doc.RetrievalScore == nil || *doc.RetrievalScore != retrieval {
					t.Fatal("original retrieval score lost")
				}
				if doc.Score == nil || math.IsNaN(*doc.Score) || math.IsInf(*doc.Score, 0) || *doc.Score < 0 || *doc.Score > 1 || *doc.Score > previous {
					t.Fatal("final scores invalid or unsorted")
				}
				previous = *doc.Score
			}
			if *result.MaxScore != *result.Docs[0].Score {
				t.Fatal("maxScore differs from final scores")
			}
		})
	}
	config := components.RerankConfig{Provider: "typesafe", Model: "jev-1.13.0", QueryText: "restore", Fields: []string{"body"}}
	input.Rerank = optionalnullable.From(&config)
	input.Query = map[string]any{"queryString": map[string]any{"query": "body:unmatchableSdkSmokeTerm"}}
	empty, err := collection.Query(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Rerank == nil || empty.Rerank.Status != "skipped" || empty.Rerank.Reason == nil || *empty.Rerank.Reason != "noCandidates" || empty.Rerank.CandidateCount != 0 || empty.Rerank.ScoredCount != 0 || empty.Rerank.CriteriaVersion != nil || empty.MaxScore != nil || len(empty.Docs) != 0 {
		t.Fatal("empty result contract changed")
	}
	criteria := []string{"Only one criterion"}
	config.Criteria = optionalnullable.From(&criteria)
	input.Rerank = optionalnullable.From(&config)
	_, err = collection.Query(ctx, input)
	requireIntegrationBadRequest(t, err, "invalid criteria before empty results")
}

// This opt-in test checks all fixed analyzer names at the server schema and
// indexing boundaries. It does not establish language-quality equivalence.
func TestIntegrationAnalyzerPresetsSmoke(t *testing.T) {
	if os.Getenv("LAMBDADB_RUN_ANALYZER_SMOKE") != "1" {
		t.Skip("set LAMBDADB_RUN_ANALYZER_SMOKE=1 to run the live smoke test")
	}
	client := lambdadb.New(
		lambdadb.WithBaseURL(normalizeIntegrationBaseURL(requireIntegrationEnv(t, "LAMBDADB_BASE_URL"))),
		lambdadb.WithProjectName(requireIntegrationEnv(t, "LAMBDADB_PROJECT_NAME")),
		lambdadb.WithAPIKey(requireIntegrationEnv(t, "LAMBDADB_PROJECT_API_KEY")),
		lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	name := fmt.Sprintf("go-sdk-analyzers-%d", time.Now().UnixNano())
	schema := map[string]components.IndexConfigsUnion{
		"body": components.CreateIndexConfigsUnionText(components.IndexConfigsText{}),
	}
	// One independently specified wire-name list covers all server presets.
	names := []string{
		"standard", "english", "korean", "japanese", "chinese", "cjk", "arabic", "french", "german", "hindi", "indonesian", "italian", "portuguese", "russian", "spanish", "turkish",
		"armenian", "basque", "bengali", "brazilian", "bulgarian", "catalan", "czech", "danish", "dutch", "estonian", "finnish", "galician", "greek", "hungarian", "irish", "latvian", "lithuanian", "norwegian", "persian", "romanian", "serbian", "sorani", "swedish", "thai", "simple", "whitespace", "stop", "keyword", "pattern", "fingerprint", "nepali", "tamil", "telugu",
	}
	for _, preset := range names {
		schema["preset_"+preset] = components.CreateIndexConfigsUnionText(components.IndexConfigsText{Analyzers: []components.Analyzer{components.Analyzer(preset)}})
	}
	collection := client.Collection(name)
	if _, err := client.Collections.Create(ctx, lambdadb.CreateCollectionOptions{CollectionName: name, IndexConfigs: schema}); err != nil {
		t.Fatalf("create temporary collection: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, err := collection.Delete(cleanupCtx); err != nil {
			t.Errorf("delete temporary collection %s: %v", name, err)
		} else {
			t.Log("temporary collection deleted")
		}
	})
	metadata, err := collection.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, preset := range names {
		field := metadata.IndexConfigs["preset_"+preset].IndexConfigsText
		if field == nil || len(field.Analyzers) != 1 || string(field.Analyzers[0]) != preset {
			t.Fatalf("analyzer %s did not round trip", preset)
		}
	}
	t.Log("all 49 analyzer presets accepted and preserved")

	doc := map[string]any{"id": "probe", "body": "Restore collection version"}
	for _, preset := range names {
		doc["preset_"+preset] = doc["body"]
	}
	if _, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{Docs: []map[string]any{doc}}); err != nil {
		t.Fatal(err)
	}
	input := lambdadb.QueryInput{
		Query:          map[string]any{"queryString": map[string]any{"query": "*:*"}},
		ConsistentRead: lambdadb.Bool(true),
	}
	waitForIntegrationCondition(t, ctx, "analyzer indexing readiness", func() (bool, error) {
		result, err := collection.Query(ctx, input)
		var apiError *apierrors.APIError
		if errors.As(err, &apiError) && apiError.StatusCode == 503 {
			return false, nil // Newly created collections may not be query-ready yet.
		}
		if err != nil {
			return false, err
		}
		return result != nil && len(result.Docs) == 1, nil
	})
	t.Log("analyzer indexing probe passed")
}
