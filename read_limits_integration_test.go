package lambdadb_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/retry"
)

// Contract: docs@ce2d8f7bd8ee2ab4f240be022085aab37cca0b91 and
// server@0a54b6ecb483ce7df9677ba2071a491f9f9879c9. This opt-in smoke creates
// and deletes temporary data. It makes no embedding or reranking provider calls.
func TestIntegrationReadLimitsSmoke(t *testing.T) {
	if os.Getenv("LAMBDADB_RUN_READ_LIMITS_SMOKE") != "1" {
		t.Skip("set LAMBDADB_RUN_READ_LIMITS_SMOKE=1 to run the live smoke test")
	}
	client := lambdadb.New(
		lambdadb.WithBaseURL(normalizeIntegrationBaseURL(requireIntegrationEnv(t, "LAMBDADB_BASE_URL"))),
		lambdadb.WithProjectName(requireIntegrationEnv(t, "LAMBDADB_PROJECT_NAME")),
		lambdadb.WithAPIKey(requireIntegrationEnv(t, "LAMBDADB_PROJECT_API_KEY")),
		lambdadb.WithRetryConfig(retry.Config{Strategy: "none"}),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	name := fmt.Sprintf("go-sdk-read-limits-%d", time.Now().UnixNano())
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
		if _, err = collection.Get(cleanup); !errors.As(err, &missing) {
			t.Errorf("collection absence not verified (%T)", err)
			return
		}
		t.Log("temporary collection deleted; GET verified ResourceNotFoundError")
	})
	schema := map[string]components.IndexConfigsUnion{"body": components.CreateIndexConfigsUnionText(components.IndexConfigsText{})}
	if _, err := client.Collections.Create(ctx, lambdadb.CreateCollectionOptions{CollectionName: name, IndexConfigs: schema}); err != nil {
		t.Fatalf("create (%T)", err)
	}
	docs := make([]map[string]any, 1000)
	ids := make([]string, 1000)
	for i := range docs {
		ids[i] = fmt.Sprintf("doc-%04d", i)
		docs[i] = map[string]any{"id": ids[i], "body": "read limit smoke"}
	}
	if _, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{Docs: docs}); err != nil {
		t.Fatalf("upsert (%T)", err)
	}
	waitForIntegrationCondition(t, ctx, "committed document list readiness", func() (bool, error) {
		result, err := collection.Docs().List(ctx, &lambdadb.ListDocsOpts{Size: lambdadb.Int64(1000)})
		var api *apierrors.APIError
		if errors.As(err, &api) && api.StatusCode == 503 {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("list readiness (%T)", err)
		}
		return result != nil && len(result.Docs) == 1000, nil
	})
	for _, size := range []int64{101, 1000} {
		result, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{Ids: ids[:size]})
		if err != nil || result == nil || len(result.Docs) != int(size) {
			t.Fatalf("fetch %d (%T)", size, err)
		}
		seen := map[string]bool{}
		for _, doc := range result.Docs {
			id, ok := doc.Doc["id"].(string)
			if !ok || seen[id] {
				t.Fatal("fetch lost or duplicated document IDs")
			}
			seen[id] = true
		}
		for _, id := range ids[:size] {
			if !seen[id] {
				t.Fatal("fetch omitted a requested ID")
			}
		}
		query, err := collection.Query(ctx, lambdadb.QueryInput{Size: lambdadb.Int64(size)})
		if err != nil || query == nil || len(query.Docs) != int(size) {
			t.Fatalf("query %d (%T)", size, err)
		}
	}
	for _, extended := range []bool{false, true} {
		for _, size := range []*int64{nil, lambdadb.Int64(101), lambdadb.Int64(1000), lambdadb.Int64(1001)} {
			options := &lambdadb.ListDocsOpts{Size: size}
			if extended {
				options.Ref = lambdadb.BranchRef("main")
			}
			result, err := collection.Docs().List(ctx, options)
			if size != nil && *size == 1001 {
				requireIntegrationBadRequest(t, err, "list above limit")
				continue
			}
			want := 100
			if size != nil {
				want = int(*size)
			}
			if err != nil || result == nil || len(result.Docs) != want {
				t.Fatalf("list extended=%t size=%d (%T)", extended, want, err)
			}
		}
	}
	_, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{Ids: append(ids, "one-too-many")})
	requireIntegrationBadRequest(t, err, "fetch above limit")
	_, err = collection.Query(ctx, lambdadb.QueryInput{Size: lambdadb.Int64(1001)})
	requireIntegrationBadRequest(t, err, "query above limit")
	t.Log("query, fetch and GET/POST list accepted 101/1000, list defaulted to 100, and 1001 returned BadRequestError")
}
