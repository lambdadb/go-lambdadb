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

// Contract: lambdadb/lambdadb@9072a1bc8925954369a887f558f1eaf387b7ea0e.
// Opt in only after verifying deployment and authorizing temporary collections
// and provider calls. Keys stay in the process environment, never test output.
func TestIntegrationBayesianNativeEmbeddingSmoke(t *testing.T) {
	if os.Getenv("LAMBDADB_RUN_BAYESIAN_SMOKE") != "1" {
		t.Skip("set LAMBDADB_RUN_BAYESIAN_SMOKE=1 to run the live smoke test")
	}
	client := lambdadb.New(lambdadb.WithBaseURL(normalizeIntegrationBaseURL(requireIntegrationEnv(t, "LAMBDADB_BASE_URL"))), lambdadb.WithProjectName(requireIntegrationEnv(t, "LAMBDADB_PROJECT_NAME")), lambdadb.WithAPIKey(requireIntegrationEnv(t, "LAMBDADB_PROJECT_API_KEY")), lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}))
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy=%t", legacy), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			name := fmt.Sprintf("go-sdk-native-%d", time.Now().UnixNano())
			collection := client.Collection(name)
			embedding := components.EmbeddingConfig{Provider: components.EmbeddingConfigProviderOpenai, Model: "text-embedding-3-small", SourceField: "body"}
			vector := components.CreateIndexConfigsUnionNativeEmbeddingVector(components.IndexConfigsNativeEmbeddingVector{Embedding: embedding})
			if legacy {
				vector = components.CreateIndexConfigsUnionManagedEmbeddingVector(components.IndexConfigsManagedEmbeddingVector{Embedding: embedding})
			}
			schema := map[string]components.IndexConfigsUnion{"body": components.CreateIndexConfigsUnionText(components.IndexConfigsText{}), "vector": vector, "manual": components.CreateIndexConfigsUnionVector(components.IndexConfigsVector{Dimensions: 2})}
			// Register cleanup before create, including ambiguous network failures.
			t.Cleanup(func() {
				cleanup, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				defer cancel()
				_, err := collection.Delete(cleanup)
				var missing *apierrors.ResourceNotFoundError
				if err != nil && !errors.As(err, &missing) {
					t.Errorf("collection cleanup failed (%T)", err)
					return
				}
				_, err = collection.Get(cleanup)
				if !errors.As(err, &missing) {
					t.Errorf("collection absence not verified (%T)", err)
					return
				}
				t.Log("temporary collection deleted; GET verified ResourceNotFoundError")
			})
			if _, err := client.Collections.Create(ctx, lambdadb.CreateCollectionOptions{CollectionName: name, IndexConfigs: schema}); err != nil {
				t.Fatalf("create (%T)", err)
			}
			schema["category"] = components.CreateIndexConfigsUnionKeyword(components.IndexConfigs{})
			stored, err := collection.Update(ctx, lambdadb.UpdateCollectionOptions{IndexConfigs: schema})
			if err != nil {
				t.Fatalf("update (%T)", err)
			}
			native := stored.IndexConfigs["vector"].IndexConfigsManagedEmbeddingVector
			if native == nil || !native.ManagedEmbedding || native.Embedding.Dimensions == nil || *native.Embedding.Dimensions != 1536 || native.Embedding.Similarity == nil || *native.Embedding.Similarity != components.SimilarityCosine {
				t.Fatal("normalized native metadata missing")
			}
			if _, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{Docs: []map[string]any{
				{"id": "restore", "body": "Restore a previous collection version using a branch created from a tag.", "manual": []float64{1, 0}},
				{"id": "search", "body": "Search a collection using text and vector queries.", "manual": []float64{0.8, 0.2}},
				{"id": "weather", "body": "Weather forecasts describe rain and temperature.", "manual": []float64{0, 1}},
			}}); err != nil {
				t.Fatalf("document embedding upsert (%T)", err)
			}
			lexical := map[string]any{"queryString": map[string]any{"query": "*:*"}}
			input := lambdadb.QueryInput{Size: lambdadb.Int64(3), Query: lexical, ConsistentRead: lambdadb.Bool(true)}
			waitForIntegrationCondition(t, ctx, "native document query readiness", func() (bool, error) {
				result, err := collection.Query(ctx, input)
				var api *apierrors.APIError
				if errors.As(err, &api) && api.StatusCode == 503 {
					return false, nil
				}
				if err != nil {
					return false, fmt.Errorf("query readiness (%T)", err)
				}
				return result != nil && len(result.Docs) == 3, nil
			})
			knn := map[string]any{"knn": map[string]any{"field": "vector", "queryText": "restore a previous collection version", "k": 3}}
			input.Query = knn
			input.IncludeVectors = lambdadb.Bool(true)
			result, err := collection.Query(ctx, input)
			if err != nil {
				t.Fatalf("ordinary native KNN (%T)", err)
			}
			if len(result.Docs) != 3 {
				t.Fatal("ordinary KNN result count")
			}
			for _, doc := range result.Docs {
				values, ok := doc.Doc["vector"].([]any)
				if !ok || len(values) != 1536 {
					t.Fatal("document embeddings not generated")
				}
				nonzero := false
				for _, v := range values {
					if n, ok := v.(float64); !ok || math.IsNaN(n) || math.IsInf(n, 0) {
						t.Fatal("invalid embedding")
					} else if n != 0 {
						nonzero = true
					}
				}
				if !nonzero {
					t.Fatal("zero embedding")
				}
			}
			input.Query = map[string]any{"knn": map[string]any{"field": "manual", "queryVector": []float64{1, 0}, "k": 3}}
			manual, err := collection.Query(ctx, input)
			if err != nil || len(manual.Docs) != 3 || manual.Docs[0].Doc["id"] != "restore" {
				t.Fatalf("caller vector KNN (%T)", err)
			}
			signals := []any{lexical, knn}
			input.Query = map[string]any{"bayesian": signals}
			input.CandidateSize = lambdadb.Int64(3)
			input.Size = lambdadb.Int64(2)
			input.IncludeVectors = nil
			result, err = collection.Query(ctx, input)
			if err != nil || len(result.Docs) != 2 || result.Rerank != nil {
				t.Fatalf("Bayesian retrieval (%T)", err)
			}
			input.Rerank = optionalnullable.From[components.RerankConfig](nil)
			if _, err := collection.Query(ctx, input); err != nil {
				t.Fatalf("Bayesian null rerank (%T)", err)
			}
			t.Log("native create/update, normalized defaults, document/query embeddings, ordinary KNN and Bayesian passed")
			if legacy {
				return
			}
			// Invalid requests use caller vectors so server validation cannot call an embedding provider.
			manualKNN := map[string]any{"knn": map[string]any{"field": "manual", "queryVector": []float64{1, 0}, "k": 3}}
			signals = []any{lexical, manualKNN}
			input.Query = map[string]any{"bayesian": signals}
			input.Rerank = nil
			invalid := map[string]lambdadb.QueryInput{}
			missing := input
			missing.CandidateSize = nil
			invalid["missing_budget"] = missing
			for _, budget := range []int64{0, 1, 101} {
				bad := input
				bad.CandidateSize = lambdadb.Int64(budget)
				invalid[fmt.Sprintf("budget_%d", budget)] = bad
			}
			bad := input
			bad.Query = map[string]any{"bayesian": []any{lexical}}
			invalid["one_signal"] = bad
			bad = input
			bad.Query = map[string]any{"bayesian": []any{lexical, manualKNN, lexical}}
			invalid["three_signals"] = bad
			boosted := map[string]any{"queryString": map[string]any{"query": "*:*"}, "boost": 1}
			for _, signal := range []any{boosted, map[string]any{"bool": []any{boosted}}} {
				bad = input
				bad.Query = map[string]any{"bayesian": []any{signal, manualKNN}}
				invalid[fmt.Sprintf("boost_%d", len(invalid))] = bad
			}
			for _, fusion := range []string{"bayesian", "rrf", "mm", "l2"} {
				bad = input
				bad.Query = map[string]any{"bayesian": []any{map[string]any{"bool": []any{map[string]any{fusion: signals}}}, manualKNN}}
				invalid["nested_"+fusion] = bad
			}
			for _, query := range []map[string]any{lexical, manualKNN, {"rrf": signals}, {"mm": signals}, {"l2": signals}} {
				bad = input
				bad.Query = query
				invalid[fmt.Sprintf("ordinary_budget_%d", len(invalid))] = bad
				ordinary := bad
				ordinary.CandidateSize = nil
				if _, err := collection.Query(ctx, ordinary); err != nil {
					t.Fatalf("ordinary query without budget (%T)", err)
				}
			}
			config := components.RerankConfig{Provider: "typesafe", Model: "jev-1.13.0", QueryText: "How do I restore a previous collection version?", Fields: []string{"body"}}
			bad = input
			bad.Rerank = optionalnullable.From(&config)
			invalid["rerank_budget_conflict"] = bad
			for label, request := range invalid {
				t.Run(label, func(t *testing.T) {
					_, err := collection.Query(ctx, request)
					var bad *apierrors.BadRequestError
					if !errors.As(err, &bad) {
						t.Fatalf("expected BadRequestError, got %T", err)
					}
				})
			}
			input.CandidateSize = nil
			input.Size = lambdadb.Int64(2)
			for _, budget := range []*int64{nil, lambdadb.Int64(3)} {
				config.CandidateSize = budget
				input.Rerank = optionalnullable.From(&config)
				result, err := collection.Query(ctx, input)
				if err != nil {
					t.Fatalf("Bayesian rerank (%T)", err)
				}
				if result.Rerank == nil || result.Rerank.Status != "applied" || result.Rerank.CandidateCount != 3 || result.Rerank.ScoredCount != 3 || len(result.Docs) != 2 {
					t.Fatal("rerank metadata or output size")
				}
				previous := math.Inf(1)
				for _, doc := range result.Docs {
					if doc.RetrievalScore == nil || doc.Score == nil || *doc.Score < 0 || *doc.Score > 1 || *doc.Score > previous {
						t.Fatal("rerank score contract")
					}
					previous = *doc.Score
				}
			}
			t.Logf("%d invalid requests returned BadRequestError; Bayesian rerank default/explicit candidate budgets applied", len(invalid))
		})
	}
}
