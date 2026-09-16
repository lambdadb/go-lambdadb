package lambdadb_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
)

const integrationConditionTimeout = 2 * time.Minute

func TestIntegrationDataVersioningSmoke(t *testing.T) {
	if os.Getenv("LAMBDADB_RUN_VERSIONING_SMOKE") != "1" {
		t.Skip("set LAMBDADB_RUN_VERSIONING_SMOKE=1 to run the live smoke test")
	}

	baseURL := normalizeIntegrationBaseURL(requireIntegrationEnv(t, "LAMBDADB_BASE_URL"))
	projectName := requireIntegrationEnv(t, "LAMBDADB_PROJECT_NAME")
	apiKey := requireIntegrationEnv(t, "LAMBDADB_PROJECT_API_KEY")

	// Allow separate indexing commits for the seed, writes, and both bulk paths.
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	suffix := time.Now().UTC().Format("20060102-150405")
	collectionName := "go-sdk-versioning-" + suffix
	branchName := "candidate-" + suffix
	defaultBranchName := "default-" + suffix
	asOfBranchName := "asof-" + suffix
	tagName := "validated-" + suffix
	defaultTagName := "default-tag-" + suffix
	aliasName := "production-" + suffix
	seedID := "seed-" + suffix
	docID := "doc-" + suffix
	bulkDocID := "bulk-" + suffix
	manualBulkDocID := "manual-bulk-" + suffix

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
		} else {
			t.Logf("cleanup deleted temporary collection %q", collectionName)
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
	cleared, err := collection.Update(ctx, lambdadb.UpdateCollectionOptions{
		Description: lambdadb.String(""),
		Tags:        map[string]string{},
	})
	if err != nil {
		t.Fatalf("clear collection metadata: %v", err)
	}
	if cleared == nil || cleared.Description != "" || len(cleared.Tags) != 0 || cleared.SnapshotRetentionInDays != 2 {
		t.Fatalf("metadata clearing or omitted retention mismatch: %#v", cleared)
	}
	t.Log("metadata clearing and omitted retention verified")

	// Parent identity exists independently of snapshot materialization.
	emptyBranch, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{BranchName: "empty-" + suffix})
	if err != nil {
		t.Fatalf("create branch from empty main: %v", err)
	}
	requireIntegrationParentBranch(t, emptyBranch, "main")
	if emptyBranch.HeadSnapshot != nil || emptyBranch.ParentSnapshot != nil {
		t.Fatalf("expected empty branch snapshots: %#v", emptyBranch)
	}
	emptyBranches, err := collection.Branches().List(ctx)
	if err != nil || len(emptyBranches) != 2 {
		t.Fatalf("list empty branches: %#v, %v", emptyBranches, err)
	}
	for _, listed := range emptyBranches {
		if listed.Name == "main" {
			if listed.ParentBranch != nil {
				t.Fatalf("main has parent: %#v", listed.ParentBranch)
			}
		} else {
			requireIntegrationParentBranch(t, &listed, "main")
			if *listed.ParentBranch != *emptyBranch.ParentBranch {
				t.Fatalf("parent changed between Create and List: %#v", listed.ParentBranch)
			}
		}
	}
	if _, err := collection.Branches().Delete(ctx, emptyBranch.Name); err != nil {
		t.Fatalf("delete empty branch: %v", err)
	}

	if _, err := collection.Docs().Upsert(ctx, lambdadb.UpsertDocsInput{
		Docs: []map[string]any{{"id": seedID, "title": "seed"}},
	}); err != nil {
		t.Fatalf("upsert seed on main: %v", err)
	}
	waitForIntegrationMainDoc(t, ctx, collection, seedID, "seed")
	waitForIntegrationBranchSnapshot(t, ctx, collection, "main")

	defaultBranch, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
		BranchName: defaultBranchName,
	})
	if err != nil {
		t.Fatalf("create branch with omitted source: %v", err)
	}
	if defaultBranch == nil || defaultBranch.Name != defaultBranchName || defaultBranch.HeadSnapshot == nil || defaultBranch.ParentSnapshot == nil {
		t.Fatalf("unexpected default-source branch: %#v", defaultBranch)
	}
	requireIntegrationParentBranch(t, defaultBranch, "main")
	_, err = collection.Branches().Create(ctx, lambdadb.CreateBranchInput{BranchName: defaultBranchName})
	requireIntegrationAlreadyExists(t, err, "create duplicate branch")

	asOfBranch, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{
		BranchName: asOfBranchName,
		Source: &lambdadb.RefSource{
			Kind: lambdadb.RefSourceKindBranch,
			Name: "main",
			AsOf: lambdadb.Int64(time.Now().Add(time.Minute).UnixMilli()),
		},
	})
	if err != nil {
		t.Fatalf("create branch with asOf source: %v", err)
	}
	if asOfBranch == nil || asOfBranch.Name != asOfBranchName || asOfBranch.HeadSnapshot == nil || asOfBranch.ParentSnapshot == nil {
		t.Fatalf("unexpected asOf branch: %#v", asOfBranch)
	}

	requireIntegrationParentBranch(t, asOfBranch, "main")

	defaultTag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{TagName: defaultTagName})
	if err != nil {
		t.Fatalf("create tag with omitted source: %v", err)
	}
	if defaultTag == nil || defaultTag.Name != defaultTagName || defaultTag.SnapshotID == "" || defaultTag.GetSnapshotCommittedAt().IsZero() {
		t.Fatalf("unexpected default-source tag: %#v", defaultTag)
	}

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
	// Exercise the manual completion path with Type omitted, in addition to the
	// one-step helper above, which passes the returned Type explicitly.
	info, err := collection.Docs().GetBulkUpsertInfoForBranch(ctx, branchName)
	if err != nil {
		t.Fatalf("get manual bulk upload info: %v", err)
	}
	if info == nil || info.Type == nil || info.HTTPMethod == nil {
		t.Fatalf("manual bulk upload info is incomplete")
	}
	payload, err := json.Marshal(map[string]any{"docs": []map[string]any{{"id": manualBulkDocID, "title": "manual bulk"}}})
	if err != nil {
		t.Fatal(err)
	}
	if info.SizeLimitBytes != nil && int64(len(payload)) > *info.SizeLimitBytes {
		t.Fatal("manual bulk payload exceeds returned size limit")
	}
	upload, err := http.NewRequestWithContext(ctx, string(*info.HTTPMethod), info.URL, bytes.NewReader(payload))
	if err != nil {
		t.Fatal("create manual bulk upload request")
	}
	upload.Header.Set("Content-Type", string(*info.Type))
	for key, value := range info.Headers {
		upload.Header.Set(key, value)
	}
	response, err := http.DefaultClient.Do(upload)
	if err != nil {
		// Do not log a presigned URL contained in a transport error.
		t.Fatal("manual bulk upload transport failure")
	}
	_, _ = io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("manual bulk upload status: %d", response.StatusCode)
	}
	if _, err := collection.Docs().BulkUpsert(ctx, lambdadb.BulkUpsertInput{
		ObjectKey: info.ObjectKey,
		Branch:    lambdadb.String(branchName),
	}); err != nil {
		t.Fatalf("bulk completion with omitted Type: %v", err)
	}
	if refFetchSupported {
		waitForIntegrationDoc(t, ctx, collection, branchName, manualBulkDocID, "manual bulk")
	}
	t.Log("helper and omitted-Type bulk imports verified with committed reads")

	branchDocs, err := collection.Docs().ListAll(ctx, &lambdadb.ListDocsOpts{
		Size: lambdadb.Int64(1),
		Ref:  &lambdadb.RefContext{Kind: lambdadb.RefKindBranch, Name: branchName},
	})
	if err != nil {
		t.Fatalf("list all paginated docs through branch: %v", err)
	}
	if !containsIntegrationRawDoc(branchDocs, seedID) ||
		!containsIntegrationRawDoc(branchDocs, docID) ||
		!containsIntegrationRawDoc(branchDocs, bulkDocID) ||
		!containsIntegrationRawDoc(branchDocs, manualBulkDocID) {
		t.Fatalf("paginated branch list did not contain expected documents: %#v", branchDocs)
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
	if tag == nil || tag.Name != tagName || tag.SnapshotID == "" || tag.GetSnapshotCommittedAt().IsZero() {
		t.Fatalf("unexpected created tag: %#v", tag)
	}

	copiedTag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{
		TagName: "copy-" + suffix,
		Source:  lambdadb.TagSource(tagName),
	})
	if err != nil || copiedTag == nil || copiedTag.SnapshotID != tag.SnapshotID {
		t.Fatalf("create tag from tag: %#v, %v", copiedTag, err)
	}
	if _, err := collection.Tags().Delete(ctx, copiedTag.Name); err != nil {
		t.Fatalf("delete copied tag: %v", err)
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

	_, err = collection.Tags().Delete(ctx, tagName)
	requireIntegrationAlreadyExists(t, err, "delete alias-referenced tag")

	_, err = collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{
		Ids:            []string{seedID},
		ConsistentRead: lambdadb.Bool(true),
		Ref:            &lambdadb.RefContext{Kind: lambdadb.RefKindTag, Name: tagName},
	})
	requireIntegrationBadRequest(t, err, "fetch tag with consistentRead")

	_, err = collection.Query(ctx, lambdadb.QueryInput{
		Query:          map[string]any{"queryString": map[string]any{"query": "*:*"}},
		ConsistentRead: lambdadb.Bool(true),
		Ref:            &lambdadb.RefContext{Kind: lambdadb.RefKindAlias, Name: aliasName},
	})
	requireIntegrationBadRequest(t, err, "query alias with consistentRead")

	_, err = collection.Branches().Delete(ctx, "main")
	requireIntegrationBadRequest(t, err, "delete main branch")

	branches, err := collection.Branches().List(ctx)
	if err != nil || !containsIntegrationBranch(branches, branchName) {
		t.Fatalf("list branches: found=%v err=%v", containsIntegrationBranch(branches, branchName), err)
	}
	tags, err := collection.Tags().List(ctx)
	if err != nil || !containsIntegrationTag(tags, tagName) {
		t.Fatalf("list tags: found=%v err=%v", containsIntegrationTag(tags, tagName), err)
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

	_, err = collection.Branches().Delete(ctx, branchName)
	requireIntegrationAlreadyExists(t, err, "delete alias-referenced branch")

	// Retargeting released the old tag, so deletion can now succeed.
	if _, err := collection.Tags().Delete(ctx, tagName); err != nil {
		t.Fatalf("delete tag after retargeting alias: %v", err)
	}

	_, err = collection.Docs().Fetch(ctx, lambdadb.FetchDocsInput{
		Ids: []string{seedID},
		Ref: lambdadb.BranchRef("missing-" + suffix),
	})
	requireIntegrationResourceNotFound(t, err, "fetch through missing ref")

	if _, err := collection.Aliases().Delete(ctx, aliasName); err != nil {
		t.Fatalf("delete alias: %v", err)
	}
	if _, err := collection.Branches().Delete(ctx, branchName); err != nil {
		t.Fatalf("delete branch after deleting alias: %v", err)
	}
	if _, err := collection.Tags().Delete(ctx, defaultTagName); err != nil {
		t.Fatalf("delete default-source tag: %v", err)
	}
	if _, err := collection.Branches().Delete(ctx, asOfBranchName); err != nil {
		t.Fatalf("delete asOf branch: %v", err)
	}
	if _, err := collection.Branches().Delete(ctx, defaultBranchName); err != nil {
		t.Fatalf("delete default-source branch: %v", err)
	}
	if _, err := collection.Delete(ctx); err != nil {
		t.Fatalf("delete collection: %v", err)
	}
	collectionDeleted = true
	t.Log("temporary collection deleted")
}

func requireIntegrationBadRequest(t *testing.T, err error, operation string) {
	t.Helper()
	var target *apierrors.BadRequestError
	if !errors.As(err, &target) {
		t.Fatalf("%s error = %T %v, want BadRequestError", operation, err, err)
	}
}

func requireIntegrationResourceNotFound(t *testing.T, err error, operation string) {
	t.Helper()
	var target *apierrors.ResourceNotFoundError
	if !errors.As(err, &target) {
		t.Fatalf("%s error = %T %v, want ResourceNotFoundError", operation, err, err)
	}
}

func requireIntegrationAlreadyExists(t *testing.T, err error, operation string) {
	t.Helper()
	var target *apierrors.ResourceAlreadyExistsError
	if !errors.As(err, &target) {
		t.Fatalf("%s error = %T %v, want ResourceAlreadyExistsError", operation, err, err)
	}
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
			ConsistentRead: lambdadb.Bool(false),
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
			ConsistentRead: lambdadb.Bool(false),
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
			ConsistentRead: lambdadb.Bool(false),
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
			if branch.Name == branchName && branch.HeadSnapshot != nil && branch.HeadSnapshot.SnapshotID != "" {
				return true, nil
			}
		}
		return false, nil
	})
}

func waitForIntegrationCondition(t *testing.T, ctx context.Context, description string, condition func() (bool, error)) {
	t.Helper()
	deadline := time.NewTimer(integrationConditionTimeout)
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
			t.Fatalf("timed out after %s waiting for %s", integrationConditionTimeout, description)
		case <-ticker.C:
		}
	}
}

func containsIntegrationBranch(refs []lambdadb.BranchDetails, name string) bool {
	for _, ref := range refs {
		if ref.Name == name {
			return true
		}
	}
	return false
}

func containsIntegrationTag(refs []lambdadb.TagDetails, name string) bool {
	for _, ref := range refs {
		if ref.Name == name {
			return true
		}
	}
	return false
}

func containsIntegrationRawDoc(docs []map[string]any, id string) bool {
	for _, doc := range docs {
		if doc["id"] == id {
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

func requireIntegrationParentBranch(t *testing.T, branch *lambdadb.BranchDetails, name string) {
	t.Helper()
	parent := branch.GetParentBranch()
	if parent == nil || parent.BranchID == "" || parent.Name != name {
		t.Fatalf("parentBranch = %#v, want source %q with nonempty identity", parent, name)
	}
}
