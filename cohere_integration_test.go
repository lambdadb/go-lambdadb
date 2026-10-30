package lambdadb_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
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

// Contract: lambdadb/lambdadb@1c3e7d453cca296a7b2af85a375b4c5016504101 api/openapi.json.
// Opt in only after verifying deployment and authorizing a temporary collection and
// Cohere embedding/reranking calls. Images are inline data URIs, so no external fetch.
func TestIntegrationCohereNativeSmoke(t *testing.T) {
	if os.Getenv("LAMBDADB_RUN_COHERE_SMOKE") != "1" {
		t.Skip("set LAMBDADB_RUN_COHERE_SMOKE=1 to run the live smoke test")
	}
	client := lambdadb.New(lambdadb.WithBaseURL(normalizeIntegrationBaseURL(requireIntegrationEnv(t, "LAMBDADB_BASE_URL"))), lambdadb.WithProjectName(requireIntegrationEnv(t, "LAMBDADB_PROJECT_NAME")), lambdadb.WithAPIKey(requireIntegrationEnv(t, "LAMBDADB_PROJECT_API_KEY")), lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	name := fmt.Sprintf("go-sdk-cohere-%d", time.Now().UnixNano())
	collection := client.Collection(name)
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

	const dims = 256
	sources := []components.EmbeddingSource{{Field: "title", Type: components.EmbeddingSourceTypeText}, {Field: "images", Type: components.EmbeddingSourceTypeImage}}
	schema := map[string]components.IndexConfigsUnion{
		"title": components.CreateIndexConfigsUnionText(components.IndexConfigsText{}),
		"vector": components.CreateIndexConfigsUnionNativeEmbeddingVector(components.IndexConfigsNativeEmbeddingVector{Embedding: components.EmbeddingConfig{
			Provider: components.EmbeddingConfigProviderCohere, Model: "embed-v5.0-fast", Sources: sources, Dimensions: lambdadb.Int64(dims),
		}}),
	}
	if _, err := client.Collections.Create(ctx, lambdadb.CreateCollectionOptions{CollectionName: name, IndexConfigs: schema}); err != nil {
		t.Fatalf("create (%T): %v", err, err)
	}
	// Re-send the server's own schema unchanged, including the server-managed id
	// field: Sources must survive decode/encode without an empty sourceField.
	stored, err := collection.Get(ctx)
	if err != nil {
		t.Fatalf("get (%T)", err)
	}
	returned := stored.IndexConfigs
	if id := returned["id"].IndexConfigs; id == nil || id.Type != components.TypeKeyword {
		t.Fatalf("server-managed id config = %#v", returned["id"])
	}
	returned["category"] = components.CreateIndexConfigsUnionKeyword(components.IndexConfigs{})
	updated, err := collection.Update(ctx, lambdadb.UpdateCollectionOptions{IndexConfigs: returned})
	if err != nil {
		t.Fatalf("update with returned schema (%T): %v", err, err)
	}
	native := updated.IndexConfigs["vector"].IndexConfigsManagedEmbeddingVector
	if native == nil || native.Embedding.Provider != components.EmbeddingConfigProviderCohere || native.Embedding.SourceField != "" ||
		fmt.Sprint(native.Embedding.Sources) != fmt.Sprint(sources) || native.Embedding.Dimensions == nil || *native.Embedding.Dimensions != dims {
		t.Fatalf("normalized Cohere metadata = %#v", native)
	}
	t.Log("Cohere sources config created, returned and re-sent unchanged")

	red, blue := integrationPNG(t, color.RGBA{R: 220, A: 255}), integrationPNG(t, color.RGBA{B: 220, A: 255})
	direct := make([]float64, dims)
	direct[0] = 1
	documents := []map[string]any{
		{"id": "text", "title": "Red running shoes for trail races"},
		{"id": "mixed", "title": "Blue canvas sneakers", "images": []string{blue}},
		{"id": "image", "images": red},
		{"id": "direct", "title": "Direct vector skips inference", "vector": direct},
	}
	if _, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{Docs: documents}); err != nil {
		t.Fatalf("upsert (%T): %v", err, err)
	}
	waitForIntegrationCondition(t, ctx, "Cohere document readiness", func() (bool, error) {
		result, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{Ids: []string{"text", "mixed", "image", "direct"}, ConsistentRead: lambdadb.Bool(false), IncludeVectors: lambdadb.Bool(true)})
		if err != nil {
			return false, fmt.Errorf("fetch (%T)", err)
		}
		return len(result.Docs) == 4, nil
	})
	fetched, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{Ids: []string{"text", "mixed", "image", "direct"}, ConsistentRead: lambdadb.Bool(true), IncludeVectors: lambdadb.Bool(true)})
	if err != nil {
		t.Fatalf("fetch vectors (%T)", err)
	}
	for _, doc := range fetched.Docs {
		values, ok := doc.Doc["vector"].([]any)
		if !ok || len(values) != dims {
			t.Fatalf("document %v has no %d-dimensional vector", doc.Doc["id"], dims)
		}
		if doc.Doc["id"] == "direct" {
			for i, v := range values {
				if n, _ := v.(float64); math.Abs(n-direct[i]) > 1e-6 {
					t.Fatal("direct vector was not preserved")
				}
			}
		}
	}
	t.Log("text, mixed text/image, image-only and direct-vector documents stored 256-dimensional vectors")

	input := lambdadb.QueryInput{Size: lambdadb.Int64(4), ConsistentRead: lambdadb.Bool(true)}
	waitForIntegrationCondition(t, ctx, "Cohere query readiness", func() (bool, error) {
		input.Query = map[string]any{"queryString": map[string]any{"query": "*:*"}}
		result, err := collection.Query(ctx, input)
		var api *apierrors.APIError
		if errors.As(err, &api) && api.StatusCode == 503 {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("query readiness (%T)", err)
		}
		return len(result.Docs) == 4, nil
	})
	for label, knn := range map[string]map[string]any{
		"queryText":   {"queryText": "red shoes"},
		"queryInputs": {"queryInputs": []map[string]any{{"type": "text", "value": "shoes"}, {"type": "image", "value": red}}},
		"queryVector": {"queryVector": direct},
	} {
		knn["field"], knn["k"] = "vector", 4
		input.Query = map[string]any{"knn": knn}
		result, err := collection.Query(ctx, input)
		if err != nil || len(result.Docs) != 4 {
			t.Fatalf("%s KNN (%T): %v", label, err, err)
		}
		if label == "queryVector" && result.Docs[0].Doc["id"] != "direct" {
			t.Fatal("direct query vector did not rank its own document first")
		}
	}
	t.Log("queryText, mixed queryInputs and direct queryVector KNN passed")

	config := components.RerankConfig{Provider: "cohere", Model: "rerank-v5.0-fast", QueryText: "red running shoes", Fields: []string{"title"}}
	input.Query = map[string]any{"queryString": map[string]any{"query": "title:shoes OR title:sneakers"}}
	input.Size = lambdadb.Int64(2)
	input.Rerank = optionalnullable.From(&config)
	result := queryCohereRerank(t, ctx, collection, input)
	meta := result.Rerank
	if meta == nil || meta.Status != "applied" || meta.Provider != "cohere" || meta.Model != "rerank-v5.0-fast" || meta.CriteriaVersion == nil || *meta.CriteriaVersion != "cohere-relevance-v1" || meta.CandidateCount != 2 || meta.ScoredCount != 2 {
		t.Fatalf("Cohere rerank metadata = %#v", meta)
	}
	for _, doc := range result.Docs {
		if doc.Score == nil || *doc.Score < 0 || *doc.Score > 1 || doc.RetrievalScore == nil {
			t.Fatal("Cohere rerank scores missing or out of range")
		}
	}
	for _, criteria := range [][]string{{}, {"Not relevant.", "Relevant."}} {
		bad := config
		bad.Criteria = optionalnullable.From(&criteria)
		invalid := input
		invalid.Rerank = optionalnullable.From(&bad)
		_, err := collection.Query(ctx, invalid)
		requireIntegrationBadRequest(t, err, fmt.Sprintf("Cohere rerank with %d criteria", len(criteria)))
	}
	t.Log("Cohere rerank applied with cohere-relevance-v1; explicit criteria rejected")

	if _, err := collection.Docs().BulkUpsertDocuments(ctx, lambdadb.UpsertDocsInput{Docs: []map[string]any{{"id": "bulk", "title": "Bulk with vector", "vector": direct}, {"id": "bulk-empty", "images": []string{}, "vector": direct}}}); err != nil {
		t.Fatalf("bulk with supplied vectors (%T): %v", err, err)
	}
	_, err = collection.Docs().BulkUpsertDocuments(ctx, lambdadb.UpsertDocsInput{Docs: []map[string]any{{"id": "bulk-missing", "title": "No vector"}}})
	requireIntegrationBadRequest(t, err, "bulk source without vector")
	waitForIntegrationMainDoc(t, ctx, collection, "bulk", "Bulk with vector")
	t.Log("bulk accepted supplied vectors and rejected a source without a vector")
}

// Provider timeouts, rate limits and outages are transient under shared load; with
// onFailure returnOriginal they arrive as a fallback naming the reason. Retry only
// those, a bounded number of times. Credential, invalid-response and any other
// fallback fails at once, and a fallback never counts as applied reranking.
var transientCohereRerankReasons = map[string]bool{"timeout": true, "rateLimit": true, "unavailable": true}

const cohereRerankAttempts = 3

// queryCohereRerank sends input with onFailure returnOriginal and returns the first
// response that is not a transient fallback. Errors fail the test as before.
func queryCohereRerank(t *testing.T, ctx context.Context, collection *lambdadb.Collection, input lambdadb.QueryInput) *lambdadb.QueryResult {
	t.Helper()
	config, _ := input.Rerank.Get()
	withFallback := *config
	withFallback.OnFailure = lambdadb.String("returnOriginal")
	input.Rerank = optionalnullable.From(&withFallback)
	for attempt := 1; ; attempt++ {
		result, err := collection.Query(ctx, input)
		if err != nil {
			t.Fatalf("Cohere rerank (%T): %v", err, err)
		}
		meta := result.Rerank
		if meta == nil || meta.Status != "fallback" {
			return result
		}
		reason := "<missing>"
		if meta.Reason != nil {
			reason = *meta.Reason
		}
		// Documented fallback shape: retrieval order without rerank scores or criteria.
		if meta.ScoredCount != 0 || meta.CriteriaVersion != nil || int64(len(result.Docs)) > *input.Size {
			t.Fatalf("Cohere %s fallback (%s) shape: scoredCount=%d criteriaVersion=%t docs=%d (size %d)", config.Model, reason, meta.ScoredCount, meta.CriteriaVersion != nil, len(result.Docs), *input.Size)
		}
		for _, doc := range result.Docs {
			if doc.RetrievalScore != nil {
				t.Fatalf("Cohere %s fallback (%s) must omit retrievalScore", config.Model, reason)
			}
		}
		if !transientCohereRerankReasons[reason] || attempt == cohereRerankAttempts {
			t.Fatalf("Cohere %s fell back with reason %s on attempt %d/%d", config.Model, reason, attempt, cohereRerankAttempts)
		}
		t.Logf("Cohere %s fell back with reason %s; retrying (%d/%d)", config.Model, reason, attempt+1, cohereRerankAttempts)
		select {
		case <-ctx.Done():
			t.Fatalf("Cohere %s rerank retry: %v", config.Model, ctx.Err())
		case <-time.After(time.Duration(attempt) * 5 * time.Second):
		}
	}
}

func integrationPNG(t *testing.T, fill color.RGBA) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for x := 0; x < 32; x++ {
		for y := 0; y < 32; y++ {
			img.Set(x, y, fill)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
