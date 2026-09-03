package lambdadb_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
)

func TestIntegrationDataVersioningSmoke(t *testing.T) {
	if os.Getenv("LAMBDADB_RUN_VERSIONING_SMOKE") != "1" {
		t.Skip("set LAMBDADB_RUN_VERSIONING_SMOKE=1 to run the live smoke test")
	}

	baseURL := normalizeIntegrationBaseURL(requireIntegrationEnv(t, "LAMBDADB_BASE_URL"))
	projectName := requireIntegrationEnv(t, "LAMBDADB_PROJECT_NAME")
	apiKey := requireIntegrationEnv(t, "LAMBDADB_PROJECT_API_KEY")

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	suffix := time.Now().UTC().Format("20060102-150405")
	collectionName := "go-sdk-versioning-" + suffix
	branchName := "candidate-" + suffix
	tagName := "validated-" + suffix
	aliasName := "production-" + suffix
	seedID := "seed-" + suffix
	docID := "doc-" + suffix
	bulkDocID := "bulk-" + suffix

	client := lambdadb.New(
		lambdadb.WithBaseURL(baseURL),
		lambdadb.WithProjectName(projectName),
		lambdadb.WithAPIKey(apiKey),
	)
	collection := client.Collection(collectionName)
	collectionDeleted := false
	t.Cleanup(func() {
		if collectionDeleted {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		if _, err := collection.Delete(cleanupCtx); err != nil {
			t.Logf("cleanup collection %q: %v", collectionName, err)
		}
	})

	t.Logf("temporary collection: %s", collectionName)
	createInput := lambdadb.CreateCollectionOptions{
		CollectionName: collectionName,
		IndexConfigs: map[string]components.IndexConfigsUnion{
			"title": components.CreateIndexConfigsUnionText(components.IndexConfigsText{
				Analyzers: []components.Analyzer{components.AnalyzerStandard},
			}),
		},
		Description:             lambdadb.String("Go SDK Data Versioning smoke test"),
		Tags:                    map[string]string{"purpose": "sdk-smoke"},
		SnapshotRetentionInDays: lambdadb.Int64(1),
	}
	created, err := client.Collections.Create(ctx, createInput)
	if err != nil && strings.Contains(err.Error(), "tags is not valid field") {
		t.Errorf("contract mismatch: create collection rejected tags: %v", err)
		createInput.Tags = nil
		created, err = client.Collections.Create(ctx, createInput)
	}
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if created == nil || created.CollectionName != collectionName {
		t.Fatalf("unexpected created collection: %#v", created)
	}
	if created.DefaultBranchName != "main" {
		t.Errorf("contract mismatch: defaultBranchName = %q, want main", created.DefaultBranchName)
	}
	if created.CreatedAt.Year() < 2020 {
		t.Errorf("contract mismatch: createdAt did not decode as epoch milliseconds: %v", created.CreatedAt.Time)
	}

	updateInput := lambdadb.UpdateCollectionOptions{
		Description:             lambdadb.String("Go SDK Data Versioning smoke test (updated)"),
		Tags:                    map[string]string{"purpose": "sdk-smoke", "state": "updated"},
		SnapshotRetentionInDays: lambdadb.Int64(2),
	}
	updated, err := collection.Update(ctx, updateInput)
	if err != nil && strings.Contains(err.Error(), "tags is not valid field") {
		t.Errorf("contract mismatch: update collection rejected tags: %v", err)
		updateInput.Tags = nil
		updated, err = collection.Update(ctx, updateInput)
	}
	if err != nil {
		t.Fatalf("update collection metadata: %v", err)
	}
	if updated == nil || updated.Description != "Go SDK Data Versioning smoke test (updated)" || updated.SnapshotRetentionInDays != 2 {
		t.Fatalf("unexpected updated collection: %#v", updated)
	}

	if _, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{
		Docs: []map[string]any{{"id": seedID, "title": "seed"}},
	}); err != nil {
		t.Fatalf("upsert seed on main: %v", err)
	}
	waitForIntegrationMainDoc(t, ctx, collection, seedID, "seed")
	waitForIntegrationBranchSnapshot(t, ctx, collection, "main")

	branch, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
		BranchName: branchName,
		Source: &lambdadb.RefSource{
			Kind: lambdadb.RefSourceKindBranch,
			Name: "main",
		},
	})
	if err != nil {
		t.Fatalf("create branch: %v", err)
	}
	if branch == nil || branch.Name != branchName || branch.CreatedAt.IsZero() {
		t.Fatalf("unexpected created branch: %#v", branch)
	}

	if _, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{
		Docs:   []map[string]any{{"id": docID, "title": "initial"}},
		Branch: lambdadb.String(branchName),
	}); err != nil {
		t.Fatalf("upsert on branch: %v", err)
	}
	refFetchSupported := checkIntegrationFetchRef(t, ctx, collection, branchName, seedID)
	if refFetchSupported {
		waitForIntegrationDoc(t, ctx, collection, branchName, docID, "initial")
	}

	if _, err := collection.Docs().Update(ctx, lambdadb.UpdateDocsInput{
		Docs:   []map[string]any{{"id": docID, "title": "updated"}},
		Branch: lambdadb.String(branchName),
	}); err != nil {
		t.Fatalf("update on branch: %v", err)
	}
	if refFetchSupported {
		waitForIntegrationDoc(t, ctx, collection, branchName, docID, "updated")
	}

	if _, err := collection.Docs().BulkUpsertDocuments(ctx, lambdadb.UpsertDocsInput{
		Docs:   []map[string]any{{"id": bulkDocID, "title": "bulk"}},
		Branch: lambdadb.String(branchName),
	}); err != nil {
		t.Fatalf("bulk upsert on branch: %v", err)
	}
	if refFetchSupported {
		waitForIntegrationDoc(t, ctx, collection, branchName, bulkDocID, "bulk")
	}

	tag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{
		TagName: tagName,
		Source: &lambdadb.RefSource{
			Kind: lambdadb.RefSourceKindBranch,
			Name: branchName,
		},
	})
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if tag == nil || tag.Name != tagName || tag.SnapshotID == nil {
		t.Fatalf("unexpected created tag: %#v", tag)
	}

	alias, err := collection.Aliases().Create(ctx, lambdadb.CreateAliasInput{
		AliasName: aliasName,
		Target: lambdadb.AliasTarget{
			Kind: lambdadb.RefSourceKindTag,
			Name: tagName,
		},
	})
	if err != nil {
		t.Fatalf("create alias: %v", err)
	}
	if alias == nil || alias.TargetKind != lambdadb.AliasTargetKindTag || alias.TargetName != tagName {
		t.Fatalf("unexpected created alias: %#v", alias)
	}

	branches, err := collection.Branches().List(ctx)
	if err != nil || !containsIntegrationRef(branches, branchName) {
		t.Fatalf("list branches: found=%v err=%v", containsIntegrationRef(branches, branchName), err)
	}
	tags, err := collection.Tags().List(ctx)
	if err != nil || !containsIntegrationRef(tags, tagName) {
		t.Fatalf("list tags: found=%v err=%v", containsIntegrationRef(tags, tagName), err)
	}
	aliases, err := collection.Aliases().List(ctx)
	if err != nil || !containsIntegrationAlias(aliases, aliasName) {
		t.Fatalf("list aliases: found=%v err=%v", containsIntegrationAlias(aliases, aliasName), err)
	}

	listed, err := collection.Docs().List(ctx, &lambdadb.ListDocsOpts{
		Ref: &lambdadb.RefContext{Kind: lambdadb.RefKindAlias, Name: aliasName},
	})
	if err != nil {
		if strings.Contains(err.Error(), "ref is not valid field") {
			t.Errorf("contract mismatch: list rejected ref: %v", err)
		} else {
			t.Fatalf("list through alias: %v", err)
		}
	} else if listed == nil || !containsIntegrationListDoc(listed.Docs, seedID) {
		t.Fatalf("alias list did not contain seed document: %#v", listed)
	}

	queried, err := collection.Query(ctx, lambdadb.QueryInput{
		Query: map[string]any{"queryString": map[string]any{"query": "*:*"}},
		Ref:   &lambdadb.RefContext{Kind: lambdadb.RefKindTag, Name: tagName},
	})
	if err != nil {
		if strings.Contains(err.Error(), "ref is not valid field") {
			t.Errorf("contract mismatch: query rejected ref: %v", err)
		} else {
			t.Fatalf("query through tag: %v", err)
		}
	} else if queried == nil || !containsIntegrationQueryDoc(queried.Docs, seedID) {
		t.Fatalf("tag query did not contain %q: %#v", seedID, queried)
	}

	alias, err = collection.Aliases().Retarget(ctx, aliasName, lambdadb.RetargetAliasInput{
		Target: lambdadb.AliasTarget{
			Kind: lambdadb.RefSourceKindBranch,
			Name: branchName,
		},
	})
	if err != nil {
		t.Fatalf("retarget alias: %v", err)
	}
	if alias == nil || alias.TargetKind != lambdadb.AliasTargetKindBranch || alias.TargetName != branchName {
		t.Fatalf("unexpected retargeted alias: %#v", alias)
	}

	if _, err := collection.Docs().Delete(ctx, lambdadb.DeleteDocsInput{
		Ids:    []string{bulkDocID},
		Branch: lambdadb.String(branchName),
	}); err != nil {
		t.Fatalf("delete on branch: %v", err)
	}
	if refFetchSupported {
		waitForIntegrationDocAbsent(t, ctx, collection, branchName, bulkDocID)
	}

	if _, err := collection.Aliases().Delete(ctx, aliasName); err != nil {
		t.Fatalf("delete alias: %v", err)
	}
	if _, err := collection.Tags().Delete(ctx, tagName); err != nil {
		t.Fatalf("delete tag: %v", err)
	}
	if _, err := collection.Branches().Delete(ctx, branchName); err != nil {
		t.Fatalf("delete branch: %v", err)
	}
	if _, err := collection.Delete(ctx); err != nil {
		t.Fatalf("delete collection: %v", err)
	}
	collectionDeleted = true
}

func checkIntegrationFetchRef(t *testing.T, ctx context.Context, collection *lambdadb.Collection, branchName, id string) bool {
	t.Helper()
	_, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{
		Ids:            []string{id},
		ConsistentRead: lambdadb.Bool(true),
		Ref:            &lambdadb.RefContext{Kind: lambdadb.RefKindBranch, Name: branchName},
	})
	if err == nil {
		return true
	}
	if strings.Contains(err.Error(), "ref is not valid field") {
		t.Errorf("contract mismatch: fetch rejected ref: %v", err)
		return false
	}
	t.Fatalf("fetch through branch: %v", err)
	return false
}

func requireIntegrationEnv(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Fatalf("required environment variable %s is not set", name)
	}
	return value
}

func normalizeIntegrationBaseURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if !strings.Contains(baseURL, "://") {
		return "https://" + baseURL
	}
	return baseURL
}

func waitForIntegrationDoc(t *testing.T, ctx context.Context, collection *lambdadb.Collection, branchName, id, title string) {
	t.Helper()
	waitForIntegrationCondition(t, ctx, "document "+id, func() (bool, error) {
		result, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{
			Ids:            []string{id},
			ConsistentRead: lambdadb.Bool(true),
			Ref:            &lambdadb.RefContext{Kind: lambdadb.RefKindBranch, Name: branchName},
		})
		if err != nil {
			return false, err
		}
		for _, doc := range result.Docs {
			if doc.Doc["id"] == id && doc.Doc["title"] == title {
				return true, nil
			}
		}
		return false, nil
	})
}

func waitForIntegrationMainDoc(t *testing.T, ctx context.Context, collection *lambdadb.Collection, id, title string) {
	t.Helper()
	waitForIntegrationCondition(t, ctx, "main document "+id, func() (bool, error) {
		result, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{
			Ids:            []string{id},
			ConsistentRead: lambdadb.Bool(true),
		})
		if err != nil {
			return false, err
		}
		for _, doc := range result.Docs {
			if doc.Doc["id"] == id && doc.Doc["title"] == title {
				return true, nil
			}
		}
		return false, nil
	})
}

func waitForIntegrationDocAbsent(t *testing.T, ctx context.Context, collection *lambdadb.Collection, branchName, id string) {
	t.Helper()
	waitForIntegrationCondition(t, ctx, "document deletion "+id, func() (bool, error) {
		result, err := collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{
			Ids:            []string{id},
			ConsistentRead: lambdadb.Bool(true),
			Ref:            &lambdadb.RefContext{Kind: lambdadb.RefKindBranch, Name: branchName},
		})
		if err != nil {
			return false, err
		}
		for _, doc := range result.Docs {
			if doc.Doc["id"] == id {
				return false, nil
			}
		}
		return true, nil
	})
}

func waitForIntegrationBranchSnapshot(t *testing.T, ctx context.Context, collection *lambdadb.Collection, branchName string) {
	t.Helper()
	waitForIntegrationCondition(t, ctx, "branch snapshot "+branchName, func() (bool, error) {
		branches, err := collection.Branches().List(ctx)
		if err != nil {
			return false, err
		}
		for _, branch := range branches {
			if branch.Name == branchName && branch.SnapshotID != nil && *branch.SnapshotID != "" {
				return true, nil
			}
		}
		return false, nil
	})
}

func waitForIntegrationCondition(t *testing.T, ctx context.Context, description string, condition func() (bool, error)) {
	t.Helper()
	deadline := time.NewTimer(90 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		ok, err := condition()
		if err != nil {
			t.Fatalf("wait for %s: %v", description, err)
		}
		if ok {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("wait for %s: %v", description, ctx.Err())
		case <-deadline.C:
			t.Fatalf("timed out waiting for %s", description)
		case <-ticker.C:
		}
	}
}

func containsIntegrationRef(refs []lambdadb.RefDetails, name string) bool {
	for _, ref := range refs {
		if ref.Name == name {
			return true
		}
	}
	return false
}

func containsIntegrationAlias(aliases []lambdadb.AliasDetails, name string) bool {
	for _, alias := range aliases {
		if alias.AliasName == name {
			return true
		}
	}
	return false
}

func containsIntegrationListDoc(docs []operations.ListDocsDoc, id string) bool {
	for _, doc := range docs {
		if doc.Doc["id"] == id {
			return true
		}
	}
	return false
}

func containsIntegrationQueryDoc(docs []operations.QueryCollectionDoc, id string) bool {
	for _, doc := range docs {
		if doc.Doc["id"] == id {
			return true
		}
	}
	return false
}
