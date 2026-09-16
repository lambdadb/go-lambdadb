package lambdadb_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
)

func TestPublicAPI_RefAndSourceConstructors(t *testing.T) {
	cutoff := time.Date(2026, time.September, 3, 12, 0, 0, 123_000_000, time.UTC)

	tests := []struct {
		name string
		ref  *lambdadb.RefContext
		kind lambdadb.RefKind
	}{
		{name: "candidate", ref: lambdadb.BranchRef("candidate"), kind: lambdadb.RefKindBranch},
		{name: "validated", ref: lambdadb.TagRef("validated"), kind: lambdadb.RefKindTag},
		{name: "production", ref: lambdadb.AliasRef("production"), kind: lambdadb.RefKindAlias},
	}
	for _, test := range tests {
		if test.ref == nil || test.ref.Kind != test.kind || test.ref.Name != test.name {
			t.Errorf("ref constructor = %#v, want %s/%s", test.ref, test.kind, test.name)
		}
	}

	branchSource := lambdadb.BranchSource("main")
	if branchSource.Kind != lambdadb.RefSourceKindBranch || branchSource.Name != "main" || branchSource.AsOf != nil {
		t.Fatalf("BranchSource() = %#v", branchSource)
	}
	branchSourceAt := lambdadb.BranchSourceAt("main", cutoff)
	if branchSourceAt.Kind != lambdadb.RefSourceKindBranch || branchSourceAt.AsOf == nil || *branchSourceAt.AsOf != cutoff.UnixMilli() {
		t.Fatalf("BranchSourceAt() = %#v, want %d", branchSourceAt, cutoff.UnixMilli())
	}
	tagSource := lambdadb.TagSource("validated")
	if tagSource.Kind != lambdadb.RefSourceKindTag || tagSource.Name != "validated" || tagSource.AsOf != nil {
		t.Fatalf("TagSource() = %#v", tagSource)
	}
	if target := lambdadb.BranchTarget("candidate"); target.Kind != lambdadb.RefSourceKindBranch || target.Name != "candidate" {
		t.Fatalf("BranchTarget() = %#v", target)
	}
	if target := lambdadb.TagTarget("validated"); target.Kind != lambdadb.RefSourceKindTag || target.Name != "validated" {
		t.Fatalf("TagTarget() = %#v", target)
	}
}

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
				return jsonResponse(http.StatusCreated, `{"branch":{"name":"candidate","headSnapshot":{"snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456},"parentSnapshot":{"snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456},"createdAt":1788336000123}}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-versioning/collections/articles/branches")
				return jsonResponse(http.StatusOK, `{"branches":[{"name":"candidate","headSnapshot":{"snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456},"parentSnapshot":{"snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456},"createdAt":1788336000123}]}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-versioning/collections/articles/tags")
				body := decodeJSONBody(t, req)
				if body["tagName"] != "validated-2026-09" {
					t.Fatalf("tagName = %v", body["tagName"])
				}
				return jsonResponse(http.StatusCreated, `{"tag":{"name":"validated-2026-09","snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456,"createdAt":1788336000123}}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-versioning/collections/articles/tags")
				return jsonResponse(http.StatusOK, `{"tags":[{"name":"validated-2026-09","snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456,"createdAt":1788336000123}]}`)
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
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodDelete, "https://api.example.com/projects/project-versioning/collections/articles/branches/candidate")
				return jsonResponse(http.StatusOK, `{"message":"Ref deleted"}`)
			},
			func(req *http.Request) *http.Response {
				assertRequest(t, req, http.MethodDelete, "https://api.example.com/projects/project-versioning/collections/articles/tags/validated-2026-09")
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
	if branch.Name != "candidate" || branch.HeadSnapshot == nil || branch.HeadSnapshot.SnapshotID != "snapshot-1" {
		t.Fatalf("created branch = %#v", branch)
	}
	if got := branch.CreatedAt.UnixMilli(); got != createdAt {
		t.Fatalf("branch createdAt = %d, want %d", got, createdAt)
	}

	branches, err := collection.Branches().List(context.Background())
	if err != nil || len(branches) != 1 {
		t.Fatalf("Branches().List() = %#v, %v", branches, err)
	}
	assertBranchSnapshots(t, branch, "snapshot-1", "snapshot-1")
	assertBranchSnapshots(t, &branches[0], "snapshot-1", "snapshot-1")

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
	for _, got := range []*lambdadb.TagDetails{tag, &tags[0]} {
		if got.GetSnapshotID() != "snapshot-1" || got.GetSnapshotCommittedAt().UnixMilli() != 1788335940456 || got.GetCreatedAt().UnixMilli() != createdAt {
			t.Fatalf("tag snapshot metadata = %#v", got)
		}
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
	if _, err := collection.Branches().Delete(context.Background(), "candidate"); err != nil {
		t.Fatalf("Branches().Delete() error = %v", err)
	}
	if _, err := collection.Tags().Delete(context.Background(), "validated-2026-09"); err != nil {
		t.Fatalf("Tags().Delete() error = %v", err)
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

func assertBranchSnapshots(t *testing.T, branch *lambdadb.BranchDetails, headID, parentID string) {
	t.Helper()
	if branch.GetName() != "candidate" || branch.GetCreatedAt().UnixMilli() != 1788336000123 {
		t.Fatalf("branch metadata = %#v", branch)
	}
	for name, snapshot := range map[string]*lambdadb.SnapshotDetails{"head": branch.GetHeadSnapshot(), "parent": branch.GetParentSnapshot()} {
		wantID := headID
		if name == "parent" {
			wantID = parentID
		}
		if wantID == "" {
			if snapshot != nil {
				t.Fatalf("%s = %#v, want nil", name, snapshot)
			}
			continue
		}
		wantCommittedAt := int64(1788335940456)
		if wantID == "head-2" {
			wantCommittedAt = 1788336060789
		}
		if snapshot == nil || snapshot.GetSnapshotID() != wantID || snapshot.GetSnapshotCommittedAt().UnixMilli() != wantCommittedAt {
			t.Fatalf("%s snapshot = %#v, want %s with millisecond commit time", name, snapshot, wantID)
		}
	}
}

func TestPublicAPI_BranchSnapshotStates(t *testing.T) {
	for _, tc := range []struct {
		name, head, parent, headID, parentID string
		parentBranch                         *lambdadb.ParentBranchDetails
	}{
		{"empty", `null`, `null`, "", "", &lambdadb.ParentBranchDetails{BranchID: "main-id", Name: "main"}},
		{"committed_from_empty", `{"snapshotId":"head-2","snapshotCommittedAt":1788336060789}`, `null`, "head-2", "", &lambdadb.ParentBranchDetails{BranchID: "main-id", Name: "main"}},
		{"advanced_head", `{"snapshotId":"head-2","snapshotCommittedAt":1788336060789}`, `{"snapshotId":"source-1","snapshotCommittedAt":1788335940456}`, "head-2", "source-1", &lambdadb.ParentBranchDetails{BranchID: "dev-id", Name: "dev"}},
		{"no_recorded_parent", `null`, `null`, "", "", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parentJSON, err := json.Marshal(tc.parentBranch)
			if err != nil {
				t.Fatal(err)
			}
			createdPayload := `{"parentBranch":` + string(parentJSON) + `,"name":"candidate","createdAt":1788336000123,"headSnapshot":` + tc.parent + `,"parentSnapshot":` + tc.parent + `}`
			payload := `{"parentBranch":` + string(parentJSON) + `,"name":"candidate","createdAt":1788336000123,"headSnapshot":` + tc.head + `,"parentSnapshot":` + tc.parent + `}`
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					return jsonResponse(http.StatusCreated, `{"branch":`+createdPayload+`}`)
				},
				func(req *http.Request) *http.Response {
					return jsonResponse(http.StatusOK, `{"branches":[`+payload+`]}`)
				},
			}}
			collection := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles")
			branch, err := collection.Branches().Create(context.Background(), lambdadb.CreateBranchInput{BranchName: "candidate"})
			if err != nil {
				t.Fatal(err)
			}
			assertBranchSnapshots(t, branch, tc.parentID, tc.parentID)
			branches, err := collection.Branches().List(context.Background())
			if err != nil || len(branches) != 1 {
				t.Fatalf("branches = %#v, %v", branches, err)
			}
			assertBranchSnapshots(t, &branches[0], tc.headID, tc.parentID)
			for _, got := range []*lambdadb.BranchDetails{branch, &branches[0]} {
				if !reflect.DeepEqual(got.GetParentBranch(), tc.parentBranch) {
					t.Fatalf("parentBranch = %#v, want %#v", got.ParentBranch, tc.parentBranch)
				}
			}
			encoded, err := json.Marshal(&branches[0])
			if err != nil {
				t.Fatal(err)
			}
			var body map[string]any
			if err := json.Unmarshal(encoded, &body); err != nil {
				t.Fatal(err)
			}
			if _, exists := body["snapshotId"]; exists {
				t.Fatalf("legacy top-level snapshotId: %s", encoded)
			}
			for _, key := range []string{"headSnapshot", "parentSnapshot", "parentBranch"} {
				value, exists := body[key]
				if !exists {
					t.Fatalf("missing required nullable %s: %s", key, encoded)
				}
				if (key == "headSnapshot" && tc.headID == "" || key == "parentSnapshot" && tc.parentID == "" || key == "parentBranch" && tc.parentBranch == nil) && value != nil {
					t.Fatalf("%s must encode as null: %s", key, encoded)
				}
			}
			if tc.parentBranch != nil {
				parent := body["parentBranch"].(map[string]any)
				if parent["branchId"] != tc.parentBranch.GetBranchID() || parent["name"] != tc.parentBranch.GetName() || len(parent) != 2 {
					t.Fatalf("parentBranch JSON = %#v", parent)
				}
			}
			mock.assertDone()
		})
	}
}

func TestPublicAPI_RefDeletionConflictDoesNotRetry(t *testing.T) {
	for _, kind := range []string{"branches", "tags"} {
		t.Run(kind, func(t *testing.T) {
			const message = "target is referenced by an alias"
			response := jsonResponse(http.StatusConflict, `{"message":"`+message+`"}`)
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					if req.Method != http.MethodDelete || !strings.HasSuffix(req.URL.Path, "/"+kind+"/candidate") {
						t.Fatalf("unexpected request: %s %s", req.Method, req.URL)
					}
					return response
				},
			}}
			collection := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles")
			var err error
			if kind == "branches" {
				_, err = collection.Branches().Delete(context.Background(), "candidate")
			} else {
				_, err = collection.Tags().Delete(context.Background(), "candidate")
			}
			var conflict *apierrors.ResourceAlreadyExistsError
			if !errors.As(err, &conflict) || conflict.Message == nil || *conflict.Message != message || conflict.HTTPMeta.Response != response {
				t.Fatalf("lost conflict status/message/metadata: %T %v", err, err)
			}
			mock.assertDone() // Default retries must not retry an in-use target.
		})
	}
}

// Contract: lambdadb/docs@c44180406c05b1a9043d8516e7c7f60df91fc9a7.
func TestPublicAPI_BranchSourceValidation(t *testing.T) {
	for _, kind := range []lambdadb.RefSourceKind{lambdadb.RefSourceKindTag, "alias", "", "unknown"} {
		t.Run(string(kind), func(t *testing.T) {
			// No handlers: invalid sources must not reach the HTTP client.
			mock := &publicAPIMockClient{t: t}
			collection := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles")
			branch, err := collection.Branches().Create(context.Background(), lambdadb.CreateBranchInput{
				BranchName: "candidate",
				Source:     &lambdadb.RefSource{Kind: kind, Name: "source"},
			})
			if branch != nil || err == nil || err.Error() != "branch source must be a branch" {
				t.Fatalf("Create = %#v, %v", branch, err)
			}
			mock.assertDone()
		})
	}
}

func TestPublicAPI_CreateSourceWireContract(t *testing.T) {
	cutoff := time.UnixMilli(1788335940456)
	for _, resource := range []string{"branches", "tags"} {
		for _, tc := range []struct {
			name   string
			source *lambdadb.RefSource
			want   string
		}{
			{"default_main", nil, `null`},
			{"branch", lambdadb.BranchSource("dev"), `{"kind":"branch","name":"dev"}`},
			{"branch_asof", lambdadb.BranchSourceAt("dev", cutoff), `{"kind":"branch","name":"dev","asOf":1788335940456}`},
			{"tag", lambdadb.TagSource("release"), `{"kind":"tag","name":"release"}`},
		} {
			if resource == "branches" && tc.name == "tag" {
				continue // Covered by local validation above.
			}
			t.Run(resource+"/"+tc.name, func(t *testing.T) {
				before, err := json.Marshal(tc.source)
				if err != nil {
					t.Fatal(err)
				}
				mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
					func(req *http.Request) *http.Response {
						if req.Method != http.MethodPost || !strings.HasSuffix(req.URL.Path, "/"+resource) {
							t.Fatalf("unexpected request: %s %s", req.Method, req.URL)
						}
						body := decodeJSONBody(t, req)
						var want any
						if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
							t.Fatal(err)
						}
						if !reflect.DeepEqual(body["source"], want) {
							t.Fatalf("source = %#v, want %#v", body["source"], want)
						}
						if tc.source == nil {
							if _, exists := body["source"]; exists {
								t.Fatal("nil source must be omitted")
							}
						}
						if resource == "tags" {
							return jsonResponse(http.StatusCreated, `{"tag":{"name":"candidate","snapshotId":"snapshot-main","snapshotCommittedAt":1788335940456,"createdAt":1788336000123}}`)
						}
						parent := `{"branchId":"dev-id","name":"dev"}`
						if tc.source == nil {
							parent = `{"branchId":"main-id","name":"main"}`
						}
						return jsonResponse(http.StatusCreated, `{"branch":{"name":"candidate","parentBranch":`+parent+`,"headSnapshot":{"snapshotId":"snapshot-main","snapshotCommittedAt":1788335940456},"parentSnapshot":{"snapshotId":"snapshot-main","snapshotCommittedAt":1788335940456},"createdAt":1788336000123}}`)
					},
				}}
				collection := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles")
				if resource == "tags" {
					tag, err := collection.Tags().Create(context.Background(), lambdadb.CreateTagInput{TagName: "candidate", Source: tc.source})
					if err != nil || tag == nil || tag.SnapshotID != "snapshot-main" {
						t.Fatalf("tag = %#v, %v", tag, err)
					}
				} else {
					branch, err := collection.Branches().Create(context.Background(), lambdadb.CreateBranchInput{BranchName: "candidate", Source: tc.source})
					if err != nil {
						t.Fatal(err)
					}
					wantName := "dev"
					if tc.source == nil {
						wantName = "main"
					}
					// The source is dev even when asOf selects a snapshot originating on main.
					if branch.GetParentBranch().GetName() != wantName || branch.ParentBranch.GetBranchID() != wantName+"-id" {
						t.Fatalf("parentBranch = %#v", branch.ParentBranch)
					}
				}
				after, err := json.Marshal(tc.source)
				if err != nil || string(before) != string(after) {
					t.Fatalf("source mutated: %s => %s, %v", before, after, err)
				}
				mock.assertDone()
			})
		}
	}
}

func TestPublicAPI_BranchListParentCompatibility(t *testing.T) {
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"branches":[
				{"name":"main","parentBranch":null,"headSnapshot":null,"parentSnapshot":null,"createdAt":1788336000123},
				{"name":"legacy","parentBranch":null,"headSnapshot":null,"parentSnapshot":null,"createdAt":1788336000123},
				{"name":"older-server","headSnapshot":null,"parentSnapshot":null,"createdAt":1788336000123}
			]}`)
		},
	}}
	branches, err := lambdadb.New(lambdadb.WithClient(mock)).Collection("articles").Branches().List(context.Background())
	if err != nil || len(branches) != 3 {
		t.Fatalf("branches = %#v, %v", branches, err)
	}
	for _, branch := range branches {
		if branch.GetParentBranch() != nil {
			t.Fatalf("unexpected parent: %#v", branch)
		}
	}
	var branch *lambdadb.BranchDetails
	var parent *lambdadb.ParentBranchDetails
	if branch.GetParentBranch() != nil || parent.GetBranchID() != "" || parent.GetName() != "" {
		t.Fatal("parent getters must be nil-safe")
	}
	mock.assertDone()
}
