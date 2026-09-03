package lambdadb_test

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
)

func TestPublicAPI_VersioningLifecycle(t *testing.T) {
	const createdAt = int64(1788336000123)
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-versioning/collections/articles/branches")
				body := decodeJSONBody(t, req)
				if body["branchName"] != "candidate" {
					t.Fatalf("branchName = %v, want candidate", body["branchName"])
				}
				source := body["source"].(map[string]any)
				if source["kind"] != "branch" || source["name"] != "main" || source["asOf"] != float64(1788336000000) {
					t.Fatalf("branch source = %#v", source)
				}
				return jsonResponse(http.StatusCreated, `{"branch":{"name":"candidate","snapshotId":"snapshot-1","createdAt":1788336000123}}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-versioning/collections/articles/branches")
				return jsonResponse(http.StatusOK, `{"branches":[{"name":"candidate","snapshotId":"snapshot-1","createdAt":1788336000123}]}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodDelete, "https://api.example.com/projects/project-versioning/collections/articles/branches/candidate")
				return jsonResponse(http.StatusOK, `{"message":"Ref deleted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-versioning/collections/articles/tags")
				body := decodeJSONBody(t, req)
				if body["tagName"] != "validated-2026-09" {
					t.Fatalf("tagName = %v", body["tagName"])
				}
				return jsonResponse(http.StatusCreated, `{"tag":{"name":"validated-2026-09","snapshotId":"snapshot-1","createdAt":1788336000123}}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-versioning/collections/articles/tags")
				return jsonResponse(http.StatusOK, `{"tags":[{"name":"validated-2026-09","snapshotId":"snapshot-1","createdAt":1788336000123}]}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodDelete, "https://api.example.com/projects/project-versioning/collections/articles/tags/validated-2026-09")
				return jsonResponse(http.StatusOK, `{"message":"Ref deleted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-versioning/collections/articles/aliases")
				body := decodeJSONBody(t, req)
				target := body["target"].(map[string]any)
				if target["kind"] != "branch" || target["name"] != "candidate" {
					t.Fatalf("alias target = %#v", target)
				}
				return jsonResponse(http.StatusCreated, aliasResponse("candidate", "BRANCH", 1))
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-versioning/collections/articles/aliases")
				return jsonResponse(http.StatusOK, `{"aliases":[{"aliasId":"alias-1","aliasName":"production","targetKind":"BRANCH","targetName":"candidate","targetId":"branch-1","aliasRevision":1,"dangling":false,"createdAt":1788336000123}]}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPatch, "https://api.example.com/projects/project-versioning/collections/articles/aliases/production")
				body := decodeJSONBody(t, req)
				target := body["target"].(map[string]any)
				if target["kind"] != "tag" || target["name"] != "validated-2026-09" {
					t.Fatalf("retarget target = %#v", target)
				}
				return jsonResponse(http.StatusOK, aliasResponse("validated-2026-09", "TAG", 2))
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodDelete, "https://api.example.com/projects/project-versioning/collections/articles/aliases/production")
				return jsonResponse(http.StatusOK, `{"message":"Ref deleted"}`)
			},
		},
	}

	client := lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-versioning"),
		lambdadb.WithClient(mock),
	)
	collection := client.Collection("articles")

	branch, err := collection.Branches().Create(context.Background(), lambdadb.CreateBranchInput{
		BranchName: "candidate",
		Source: &lambdadb.RefSource{
			Kind: lambdadb.RefSourceKindBranch,
			Name: "main",
			AsOf: lambdadb.Int64(1788336000000),
		},
	})
	if err != nil {
		t.Fatalf("Branches().Create() error = %v", err)
	}
	if branch.Name != "candidate" || branch.SnapshotID == nil || *branch.SnapshotID != "snapshot-1" {
		t.Fatalf("created branch = %#v", branch)
	}
	if got := branch.CreatedAt.UnixMilli(); got != createdAt {
		t.Fatalf("branch createdAt = %d, want %d", got, createdAt)
	}

	branches, err := collection.Branches().List(context.Background())
	if err != nil || len(branches) != 1 {
		t.Fatalf("Branches().List() = %#v, %v", branches, err)
	}
	if _, err := collection.Branches().Delete(context.Background(), "candidate"); err != nil {
		t.Fatalf("Branches().Delete() error = %v", err)
	}

	tag, err := collection.Tags().Create(context.Background(), lambdadb.CreateTagInput{
		TagName: "validated-2026-09",
		Source:  &lambdadb.RefSource{Kind: lambdadb.RefSourceKindBranch, Name: "candidate"},
	})
	if err != nil || tag.Name != "validated-2026-09" {
		t.Fatalf("Tags().Create() = %#v, %v", tag, err)
	}
	tags, err := collection.Tags().List(context.Background())
	if err != nil || len(tags) != 1 {
		t.Fatalf("Tags().List() = %#v, %v", tags, err)
	}
	if _, err := collection.Tags().Delete(context.Background(), "validated-2026-09"); err != nil {
		t.Fatalf("Tags().Delete() error = %v", err)
	}

	alias, err := collection.Aliases().Create(context.Background(), lambdadb.CreateAliasInput{
		AliasName: "production",
		Target: lambdadb.AliasTarget{
			Kind: lambdadb.RefSourceKindBranch,
			Name: "candidate",
		},
	})
	if err != nil || alias.TargetName != "candidate" || alias.AliasRevision != 1 {
		t.Fatalf("Aliases().Create() = %#v, %v", alias, err)
	}
	aliases, err := collection.Aliases().List(context.Background())
	if err != nil || len(aliases) != 1 {
		t.Fatalf("Aliases().List() = %#v, %v", aliases, err)
	}
	alias, err = collection.Aliases().Retarget(context.Background(), "production", lambdadb.RetargetAliasInput{
		Target: lambdadb.AliasTarget{
			Kind: lambdadb.RefSourceKindTag,
			Name: "validated-2026-09",
		},
	})
	if err != nil || alias.TargetName != "validated-2026-09" || alias.AliasRevision != 2 {
		t.Fatalf("Aliases().Retarget() = %#v, %v", alias, err)
	}
	if _, err := collection.Aliases().Delete(context.Background(), "production"); err != nil {
		t.Fatalf("Aliases().Delete() error = %v", err)
	}
	mock.assertDone()
}

func TestPublicAPI_VersioningTypedError(t *testing.T) {
	mock := &publicAPIMockClient{
		t: t,
		handlers: []func(*http.Request) *http.Response{
			func(req *http.Request) *http.Response {
				return jsonResponse(http.StatusConflict, `{"message":"branch already exists"}`)
			},
		},
	}
	client := lambdadb.New(lambdadb.WithClient(mock))

	_, err := client.Collection("articles").Branches().Create(context.Background(), lambdadb.CreateBranchInput{BranchName: "candidate"})
	var conflict *apierrors.ResourceAlreadyExistsError
	if !errors.As(err, &conflict) {
		t.Fatalf("Branches().Create() error = %T %v, want ResourceAlreadyExistsError", err, err)
	}
	if conflict.Message == nil || *conflict.Message != "branch already exists" {
		t.Fatalf("conflict message = %v", conflict.Message)
	}
	mock.assertDone()
}

func aliasResponse(targetName, targetKind string, revision int) string {
	return `{"alias":{"aliasId":"alias-1","aliasName":"production","targetKind":"` + targetKind + `","targetName":"` + targetName + `","targetId":"target-1","aliasRevision":` + strconv.Itoa(revision) + `,"dangling":false,"createdAt":1788336000123}}`
}
