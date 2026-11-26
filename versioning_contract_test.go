package lambdadb_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/apierrors"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/models/operations"
	"github.com/lambdadb/go-lambdadb/retry"
)

// Contract: lambdadb/docs@9a1f563b9b53bc29e5278105f8fc8c25a17e0cb8 reference/api/openapi.json
// (SHA-256 269590b8addc8e25785c222938b84e737a9f9dee0a5e8b45c0dde4779bde1b6d, info.version 1.3.0),
// synced from lambdadb/lambdadb@16c6e46854b13cebd20c501e541bb030b33ca805.
// These tests exercise the public SDK wire boundary for the individual ref
// routes, branch schema updates, paginated ref listings, the ref fields of
// every branch, tag and alias response, branch rollback, and tags created
// before the first snapshot.

const (
	contractRefBase   = "https://api.example.com/projects/project-versioning/collections/articles"
	contractSchema    = `{"id":{"type":"keyword"},"title":{"type":"text","analyzers":["english"]}}`
	contractBranch    = `{"name":"candidate","parentBranch":{"branchId":"main-id","name":"main"},"headSnapshot":{"snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456},"parentSnapshot":{"snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456},"numDocs":1000,"dataUpdatedAt":1788335940000,"indexConfigs":` + contractSchema + `,"createdAt":1788336000123}`
	contractTag       = `{"name":"release","snapshotId":"snapshot-1","snapshotCommittedAt":1788335940456,"numDocs":1000,"dataUpdatedAt":1788335940000,"indexConfigs":` + contractSchema + `,"createdAt":1788336000123}`
	contractAlias     = `{"aliasId":"alias-1","aliasName":"latest","targetKind":"BRANCH","targetName":"main","targetId":"main-id","aliasRevision":1,"dangling":false,"indexConfigs":` + contractSchema + `,"createdAt":1788336000123}`
	contractDataMilli = int64(1788335940000)
)

func contractCollection(mock *publicAPIMockClient) *lambdadb.Collection {
	return lambdadb.New(
		lambdadb.WithAPIKey("public-key"),
		lambdadb.WithBaseURL("https://api.example.com"),
		lambdadb.WithProjectName("project-versioning"),
		lambdadb.WithClient(mock),
	).Collection("articles")
}

// assertBareRequest checks method and URL and that neither a query string nor
// a body is sent.
func assertBareRequest(t *testing.T, req *http.Request, method, url string) {
	t.Helper()
	assertRequest(t, req, method, url)
	if req.URL.RawQuery != "" {
		t.Fatalf("query = %q, want none", req.URL.RawQuery)
	}
	if req.ContentLength != 0 || req.Header.Get("Content-Type") != "" {
		t.Fatalf("unexpected body: length %d, content type %q", req.ContentLength, req.Header.Get("Content-Type"))
	}
}

func assertContractSchema(t *testing.T, schema map[string]components.IndexConfigsUnion) {
	t.Helper()
	title, ok := schema["title"]
	if len(schema) != 2 || !ok || title.IndexConfigsText == nil || !reflect.DeepEqual(title.IndexConfigsText.Analyzers, []components.Analyzer{components.AnalyzerEnglish}) {
		t.Fatalf("indexConfigs = %#v", schema)
	}
	if id := schema["id"]; id.Type != components.IndexConfigsUnionTypeKeyword {
		t.Fatalf("id field = %#v", id)
	}
}

func TestVersioningContract_GetRefs(t *testing.T) {
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			assertBareRequest(t, req, http.MethodGet, contractRefBase+"/branches/candidate")
			return jsonResponse(http.StatusOK, `{"branch":`+contractBranch+`}`)
		},
		func(req *http.Request) *http.Response {
			assertBareRequest(t, req, http.MethodGet, contractRefBase+"/tags/release")
			return jsonResponse(http.StatusOK, `{"tag":`+contractTag+`}`)
		},
		func(req *http.Request) *http.Response {
			assertBareRequest(t, req, http.MethodGet, contractRefBase+"/aliases/latest")
			return jsonResponse(http.StatusOK, `{"alias":`+contractAlias+`}`)
		},
	}}
	collection := contractCollection(mock)

	branch, err := collection.Branches().Get(context.Background(), "candidate")
	if err != nil {
		t.Fatalf("Branches().Get() error = %v", err)
	}
	if branch.GetName() != "candidate" || branch.GetNumDocs() != 1000 || branch.DataUpdatedAt == nil || branch.GetDataUpdatedAt().UnixMilli() != contractDataMilli {
		t.Fatalf("branch = %#v", branch)
	}
	assertBranchSnapshots(t, branch, "snapshot-1", "snapshot-1")
	assertContractSchema(t, branch.GetIndexConfigs())

	tag, err := collection.Tags().Get(context.Background(), "release")
	if err != nil {
		t.Fatalf("Tags().Get() error = %v", err)
	}
	if tag.GetName() != "release" || tag.GetSnapshotID() != "snapshot-1" || tag.GetNumDocs() != 1000 || tag.GetDataUpdatedAt().UnixMilli() != contractDataMilli {
		t.Fatalf("tag = %#v", tag)
	}
	assertContractSchema(t, tag.GetIndexConfigs())

	alias, err := collection.Aliases().Get(context.Background(), "latest")
	if err != nil {
		t.Fatalf("Aliases().Get() error = %v", err)
	}
	if alias.GetAliasName() != "latest" || alias.GetTargetKind() != lambdadb.AliasTargetKindBranch || alias.GetDangling() {
		t.Fatalf("alias = %#v", alias)
	}
	assertContractSchema(t, alias.GetIndexConfigs())
	mock.assertDone()
}

func TestVersioningContract_GetRefNotFound(t *testing.T) {
	for _, kind := range []string{"branches", "tags", "aliases"} {
		t.Run(kind, func(t *testing.T) {
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					assertBareRequest(t, req, http.MethodGet, contractRefBase+"/"+kind+"/missing")
					return jsonResponse(http.StatusNotFound, `{"message":"Resource not found"}`)
				},
			}}
			collection := contractCollection(mock)
			var err error
			switch kind {
			case "branches":
				_, err = collection.Branches().Get(context.Background(), "missing")
			case "tags":
				_, err = collection.Tags().Get(context.Background(), "missing")
			default:
				_, err = collection.Aliases().Get(context.Background(), "missing")
			}
			var notFound *apierrors.ResourceNotFoundError
			if !errors.As(err, &notFound) {
				t.Fatalf("error = %T %v, want ResourceNotFoundError", err, err)
			}
			mock.assertDone()
		})
	}
}

// An empty branch has an explicit null dataUpdatedAt and a dangling alias an
// explicit null indexConfigs; both decode to nil and encode back as null.
func TestVersioningContract_NullableRefFields(t *testing.T) {
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"branch":{"name":"empty","parentBranch":{"branchId":"main-id","name":"main"},"headSnapshot":null,"parentSnapshot":null,"numDocs":0,"dataUpdatedAt":null,"indexConfigs":`+contractSchema+`,"createdAt":1788336000123}}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"alias":{"aliasId":"alias-1","aliasName":"latest","targetKind":"BRANCH","targetName":"gone","targetId":"gone-id","aliasRevision":3,"dangling":true,"indexConfigs":null,"createdAt":1788336000123}}`)
		},
	}}
	collection := contractCollection(mock)

	branch, err := collection.Branches().Get(context.Background(), "empty")
	if err != nil {
		t.Fatalf("Branches().Get() error = %v", err)
	}
	if branch.DataUpdatedAt != nil || !branch.GetDataUpdatedAt().IsZero() || branch.GetNumDocs() != 0 || branch.GetHeadSnapshot() != nil {
		t.Fatalf("empty branch = %#v", branch)
	}
	assertContractSchema(t, branch.GetIndexConfigs())

	alias, err := collection.Aliases().Get(context.Background(), "latest")
	if err != nil {
		t.Fatalf("Aliases().Get() error = %v", err)
	}
	if !alias.GetDangling() || alias.IndexConfigs != nil || alias.GetIndexConfigs() != nil || alias.GetAliasRevision() != 3 {
		t.Fatalf("dangling alias = %#v", alias)
	}

	for name, value := range map[string]any{"dataUpdatedAt": branch, "indexConfigs": alias} {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var body map[string]any
		if err := json.Unmarshal(encoded, &body); err != nil {
			t.Fatal(err)
		}
		if got, exists := body[name]; !exists || got != nil {
			t.Fatalf("%s must encode as explicit null: %s", name, encoded)
		}
	}

	var nilBranch *lambdadb.BranchDetails
	var nilTag *lambdadb.TagDetails
	var nilAlias *lambdadb.AliasDetails
	if nilBranch.GetNumDocs() != 0 || !nilBranch.GetDataUpdatedAt().IsZero() || nilBranch.GetIndexConfigs() != nil ||
		nilTag.GetNumDocs() != 0 || !nilTag.GetDataUpdatedAt().IsZero() || nilTag.GetIndexConfigs() != nil ||
		nilAlias.GetIndexConfigs() != nil {
		t.Fatal("new getters must be nil-safe")
	}
	mock.assertDone()
}

func TestVersioningContract_UpdateBranch(t *testing.T) {
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			assertRequest(t, req, http.MethodPatch, contractRefBase+"/branches/candidate")
			if req.URL.RawQuery != "" {
				t.Fatalf("query = %q, want none", req.URL.RawQuery)
			}
			var want map[string]any
			if err := json.Unmarshal([]byte(`{"indexConfigs":{"title":{"type":"text","analyzers":["english"]},"author":{"type":"keyword"}}}`), &want); err != nil {
				t.Fatal(err)
			}
			// The body is strict: indexConfigs is its only key.
			if body := decodeJSONBody(t, req); !reflect.DeepEqual(body, want) {
				t.Fatalf("body = %#v, want %#v", body, want)
			}
			return jsonResponse(http.StatusOK, `{"branch":`+contractBranch+`}`)
		},
		func(req *http.Request) *http.Response {
			assertRequest(t, req, http.MethodPatch, contractRefBase+"/branches/release")
			return jsonResponse(http.StatusBadRequest, `{"message":"release is not a branch"}`)
		},
	}}
	collection := contractCollection(mock)
	input := lambdadb.UpdateBranchInput{IndexConfigs: map[string]components.IndexConfigsUnion{
		"title":  components.CreateIndexConfigsUnionText(components.IndexConfigsText{Analyzers: []components.Analyzer{components.AnalyzerEnglish}}),
		"author": components.CreateIndexConfigsUnionKeyword(components.IndexConfigs{}),
	}}

	branch, err := collection.Branches().Update(context.Background(), "candidate", input)
	if err != nil {
		t.Fatalf("Branches().Update() error = %v", err)
	}
	if branch.GetName() != "candidate" || branch.GetNumDocs() != 1000 {
		t.Fatalf("updated branch = %#v", branch)
	}
	assertContractSchema(t, branch.GetIndexConfigs())

	// Only a branch can be updated; the server's 400 for a tag or alias name is preserved.
	_, err = collection.Branches().Update(context.Background(), "release", input)
	var badRequest *apierrors.BadRequestError
	if !errors.As(err, &badRequest) || badRequest.Message == nil || *badRequest.Message != "release is not a branch" {
		t.Fatalf("error = %T %v, want BadRequestError", err, err)
	}
	mock.assertDone()
}

// Create, retarget and list entries carry the same ref fields as single reads.
func TestVersioningContract_RefFieldsInEveryResponse(t *testing.T) {
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusCreated, `{"branch":`+contractBranch+`}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusCreated, `{"tag":`+contractTag+`}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusCreated, `{"alias":`+contractAlias+`}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"alias":`+contractAlias+`}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"branches":[`+contractBranch+`]}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"tags":[`+contractTag+`]}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"aliases":[`+contractAlias+`]}`)
		},
	}}
	collection := contractCollection(mock)
	ctx := context.Background()

	branch, err := collection.Branches().Create(ctx, lambdadb.CreateBranchInput{BranchName: "candidate"})
	if err != nil {
		t.Fatal(err)
	}
	tag, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{TagName: "release"})
	if err != nil {
		t.Fatal(err)
	}
	alias, err := collection.Aliases().Create(ctx, lambdadb.CreateAliasInput{AliasName: "latest", Target: lambdadb.BranchTarget("main")})
	if err != nil {
		t.Fatal(err)
	}
	retargeted, err := collection.Aliases().Retarget(ctx, "latest", lambdadb.RetargetAliasInput{Target: lambdadb.BranchTarget("main")})
	if err != nil {
		t.Fatal(err)
	}
	branches, err := collection.Branches().List(ctx)
	if err != nil || len(branches) != 1 {
		t.Fatalf("branches = %#v, %v", branches, err)
	}
	tags, err := collection.Tags().List(ctx)
	if err != nil || len(tags) != 1 {
		t.Fatalf("tags = %#v, %v", tags, err)
	}
	aliases, err := collection.Aliases().List(ctx)
	if err != nil || len(aliases) != 1 {
		t.Fatalf("aliases = %#v, %v", aliases, err)
	}

	for _, got := range []*lambdadb.BranchDetails{branch, &branches[0]} {
		if got.GetNumDocs() != 1000 || got.GetDataUpdatedAt().UnixMilli() != contractDataMilli {
			t.Fatalf("branch = %#v", got)
		}
		assertContractSchema(t, got.GetIndexConfigs())
	}
	for _, got := range []*lambdadb.TagDetails{tag, &tags[0]} {
		if got.GetNumDocs() != 1000 || got.GetDataUpdatedAt().UnixMilli() != contractDataMilli {
			t.Fatalf("tag = %#v", got)
		}
		assertContractSchema(t, got.GetIndexConfigs())
	}
	for _, got := range []*lambdadb.AliasDetails{alias, retargeted, &aliases[0]} {
		assertContractSchema(t, got.GetIndexConfigs())
	}
	mock.assertDone()
}

// refListCall runs one ListPage call for a ref kind and returns the number of
// entries and the next page token.
func refListCall(t *testing.T, collection *lambdadb.Collection, kind string, listOpts *lambdadb.ListRefsOpts) (int, *string) {
	t.Helper()
	ctx := context.Background()
	switch kind {
	case "branches":
		page, err := collection.Branches().ListPage(ctx, listOpts)
		if err != nil {
			t.Fatalf("Branches().ListPage() error = %v", err)
		}
		return len(page.Branches), page.NextPageToken
	case "tags":
		page, err := collection.Tags().ListPage(ctx, listOpts)
		if err != nil {
			t.Fatalf("Tags().ListPage() error = %v", err)
		}
		return len(page.Tags), page.NextPageToken
	default:
		page, err := collection.Aliases().ListPage(ctx, listOpts)
		if err != nil {
			t.Fatalf("Aliases().ListPage() error = %v", err)
		}
		return len(page.Aliases), page.NextPageToken
	}
}

func refListBody(kind, nextPageToken string) string {
	entry := map[string]string{"branches": contractBranch, "tags": contractTag, "aliases": contractAlias}[kind]
	body := `{"` + kind + `":[` + entry + `]`
	if nextPageToken != "" {
		body += `,"nextPageToken":"` + nextPageToken + `"`
	}
	return body + `}`
}

// The server rejects unknown query parameters and an empty pageToken on the
// ref listings, so unset values and an empty token must not be sent at all.
func TestVersioningContract_ListPageQueryParameters(t *testing.T) {
	for _, kind := range []string{"branches", "tags", "aliases"} {
		for _, tc := range []struct {
			name      string
			listOpts  *lambdadb.ListRefsOpts
			wantQuery string
			nextToken string
		}{
			{"nil_opts", nil, "", ""},
			{"zero_opts", &lambdadb.ListRefsOpts{}, "", ""},
			{"empty_page_token", &lambdadb.ListRefsOpts{PageToken: lambdadb.String("")}, "", ""},
			{"size_only", &lambdadb.ListRefsOpts{Size: lambdadb.Int64(100)}, "size=100", "page-2"},
			{"page_token_only", &lambdadb.ListRefsOpts{PageToken: lambdadb.String("page 2/+")}, "pageToken=page+2%2F%2B", ""},
			{"size_and_page_token", &lambdadb.ListRefsOpts{Size: lambdadb.Int64(1), PageToken: lambdadb.String("page-2")}, "pageToken=page-2&size=1", "page-3"},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
					func(req *http.Request) *http.Response {
						assertRequest(t, req, http.MethodGet, contractRefBase+"/"+kind)
						if req.URL.RawQuery != tc.wantQuery {
							t.Fatalf("query = %q, want %q", req.URL.RawQuery, tc.wantQuery)
						}
						return jsonResponse(http.StatusOK, refListBody(kind, tc.nextToken))
					},
				}}
				count, next := refListCall(t, contractCollection(mock), kind, tc.listOpts)
				if count != 1 {
					t.Fatalf("entries = %d, want 1", count)
				}
				if tc.nextToken == "" {
					if next != nil {
						t.Fatalf("nextPageToken = %q, want nil on the last page", *next)
					}
				} else if next == nil || *next != tc.nextToken {
					t.Fatalf("nextPageToken = %v, want %q", next, tc.nextToken)
				}
				mock.assertDone()
			})
		}
	}
}

// List keeps returning every ref: it follows nextPageToken with the server's
// default page size and sends nothing on the first request.
func TestVersioningContract_ListFollowsEveryPage(t *testing.T) {
	for _, kind := range []string{"branches", "tags", "aliases"} {
		t.Run(kind, func(t *testing.T) {
			wantQueries := []string{"", "pageToken=page-2", "pageToken=page-3"}
			nextTokens := []string{"page-2", "page-3", ""}
			var handlers []func(*http.Request) *http.Response
			for i := range wantQueries {
				handlers = append(handlers, func(req *http.Request) *http.Response {
					assertRequest(t, req, http.MethodGet, contractRefBase+"/"+kind)
					if req.URL.RawQuery != wantQueries[i] {
						t.Fatalf("request %d query = %q, want %q", i+1, req.URL.RawQuery, wantQueries[i])
					}
					return jsonResponse(http.StatusOK, refListBody(kind, nextTokens[i]))
				})
			}
			mock := &publicAPIMockClient{t: t, handlers: handlers}
			collection := contractCollection(mock)
			var count int
			var err error
			switch kind {
			case "branches":
				var refs []lambdadb.BranchDetails
				refs, err = collection.Branches().List(context.Background())
				count = len(refs)
			case "tags":
				var refs []lambdadb.TagDetails
				refs, err = collection.Tags().List(context.Background())
				count = len(refs)
			default:
				var refs []lambdadb.AliasDetails
				refs, err = collection.Aliases().List(context.Background())
				count = len(refs)
			}
			if err != nil || count != 3 {
				t.Fatalf("List() = %d entries, %v; want 3", count, err)
			}
			mock.assertDone()
		})
	}
}

func TestVersioningContract_ListIteratorKeepsSize(t *testing.T) {
	wantQueries := []string{"size=2", "pageToken=page-2&size=2"}
	nextTokens := []string{"page-2", ""}
	var handlers []func(*http.Request) *http.Response
	for i := range wantQueries {
		handlers = append(handlers, func(req *http.Request) *http.Response {
			if req.URL.RawQuery != wantQueries[i] {
				t.Fatalf("request %d query = %q, want %q", i+1, req.URL.RawQuery, wantQueries[i])
			}
			return jsonResponse(http.StatusOK, refListBody("tags", nextTokens[i]))
		})
	}
	mock := &publicAPIMockClient{t: t, handlers: handlers}
	it := contractCollection(mock).Tags().ListIterator(context.Background(), &lambdadb.ListRefsOpts{Size: lambdadb.Int64(2)})
	pages := 0
	for {
		page, err := it.Next(context.Background())
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		if page == nil {
			break
		}
		if len(page.Tags) != 1 {
			t.Fatalf("page = %#v", page)
		}
		pages++
	}
	if page, err := it.Next(context.Background()); page != nil || err != nil {
		t.Fatalf("Next() after the last page = %#v, %v", page, err)
	}
	if pages != 2 {
		t.Fatalf("pages = %d, want 2", pages)
	}
	mock.assertDone()
}

func TestVersioningContract_ListPageBadRequest(t *testing.T) {
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			if req.URL.RawQuery != "size=101" {
				t.Fatalf("query = %q, want size=101", req.URL.RawQuery)
			}
			return jsonResponse(http.StatusBadRequest, `{"message":"size must be between 1 and 100"}`)
		},
	}}
	// The SDK does not clamp size; the server's validation and 400 are preserved.
	_, err := contractCollection(mock).Branches().ListPage(context.Background(), &lambdadb.ListRefsOpts{Size: lambdadb.Int64(101)})
	var badRequest *apierrors.BadRequestError
	if !errors.As(err, &badRequest) {
		t.Fatalf("error = %T %v, want BadRequestError", err, err)
	}
	mock.assertDone()
}

// A tag source with asOf is invalid; the SDK forwards it and preserves the
// server's 400 instead of validating locally.
func TestVersioningContract_BranchFromTagWithAsOf(t *testing.T) {
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			assertRequest(t, req, http.MethodPost, contractRefBase+"/branches")
			want := map[string]any{"kind": "tag", "name": "release", "asOf": float64(1788335940456)}
			if source := decodeJSONBody(t, req)["source"]; !reflect.DeepEqual(source, want) {
				t.Fatalf("source = %#v, want %#v", source, want)
			}
			return jsonResponse(http.StatusBadRequest, `{"message":"asOf is not allowed for a tag source"}`)
		},
	}}
	_, err := contractCollection(mock).Branches().Create(context.Background(), lambdadb.CreateBranchInput{
		BranchName: "hotfix",
		Source:     &lambdadb.RefSource{Kind: lambdadb.RefSourceKindTag, Name: "release", AsOf: lambdadb.Int64(1788335940456)},
	})
	var badRequest *apierrors.BadRequestError
	if !errors.As(err, &badRequest) {
		t.Fatalf("error = %T %v, want BadRequestError", err, err)
	}
	mock.assertDone()
}

// GET /collections rejects an empty pageToken and declares a 400 response.
func TestVersioningContract_ListCollectionsPageToken(t *testing.T) {
	for _, tc := range []struct {
		name      string
		listOpts  *lambdadb.ListCollectionsOpts
		wantQuery string
	}{
		{"nil_opts", nil, ""},
		{"zero_opts", &lambdadb.ListCollectionsOpts{}, ""},
		{"empty_page_token", &lambdadb.ListCollectionsOpts{PageToken: lambdadb.String("")}, ""},
		{"size_with_empty_page_token", &lambdadb.ListCollectionsOpts{Size: lambdadb.Int64(5), PageToken: lambdadb.String("")}, "size=5"},
		{"page_token", &lambdadb.ListCollectionsOpts{PageToken: lambdadb.String("page-2")}, "pageToken=page-2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					assertRequest(t, req, http.MethodGet, "https://api.example.com/projects/project-versioning/collections")
					if req.URL.RawQuery != tc.wantQuery {
						t.Fatalf("query = %q, want %q", req.URL.RawQuery, tc.wantQuery)
					}
					return jsonResponse(http.StatusOK, `{"collections":[]}`)
				},
			}}
			client := lambdadb.New(lambdadb.WithBaseURL("https://api.example.com"), lambdadb.WithProjectName("project-versioning"), lambdadb.WithClient(mock))
			res, err := client.Collections.List(context.Background(), tc.listOpts)
			if err != nil || res.NextPageToken != nil {
				t.Fatalf("Collections.List() = %#v, %v", res, err)
			}
			mock.assertDone()
		})
	}

	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusBadRequest, `{"message":"Invalid pageToken"}`)
		},
	}}
	_, err := lambdadb.New(lambdadb.WithClient(mock)).Collections.List(context.Background(), &lambdadb.ListCollectionsOpts{PageToken: lambdadb.String("tampered")})
	var badRequest *apierrors.BadRequestError
	if !errors.As(err, &badRequest) || badRequest.Message == nil || *badRequest.Message != "Invalid pageToken" {
		t.Fatalf("error = %T %v, want BadRequestError", err, err)
	}
	mock.assertDone()
}

// POST /collections returns the same shape as GET /collections/{collectionName}.
func TestVersioningContract_CreateCollectionReturnsFullShape(t *testing.T) {
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			assertRequest(t, req, http.MethodPost, "https://api.example.com/projects/project-versioning/collections")
			return jsonResponse(http.StatusCreated, `{"collection":{"projectName":"project-versioning","collectionName":"articles","indexConfigs":`+contractSchema+`,"description":"Product catalog","tags":{"environment":"production"},"numPartitions":1,"numDocs":0,"defaultBranchName":"main","snapshotRetentionInDays":30,"createdAt":1788336000000,"updatedAt":1788336000000}}`)
		},
	}}
	client := lambdadb.New(lambdadb.WithBaseURL("https://api.example.com"), lambdadb.WithProjectName("project-versioning"), lambdadb.WithClient(mock))
	created, err := client.Collections.Create(context.Background(), lambdadb.CreateCollectionOptions{
		CollectionName: "articles",
		IndexConfigs: map[string]components.IndexConfigsUnion{
			"title": components.CreateIndexConfigsUnionText(components.IndexConfigsText{Analyzers: []components.Analyzer{components.AnalyzerEnglish}}),
		},
	})
	if err != nil {
		t.Fatalf("Collections.Create() error = %v", err)
	}
	if created.GetProjectName() != "project-versioning" || created.GetNumPartitions() != 1 || created.GetNumDocs() != 0 ||
		created.GetUpdatedAt().UnixMilli() != 1788336000000 || !created.GetDataUpdatedAt().IsZero() {
		t.Fatalf("created collection = %#v", created)
	}
	assertContractSchema(t, created.GetIndexConfigs())
	mock.assertDone()
}

// The rollback body is strict: tag is its only key. The response head is a
// snapshot, or explicit null when the tag had no published snapshot.
func TestVersioningContract_RollbackBranch(t *testing.T) {
	assertRollbackRequest := func(req *http.Request) {
		assertRequest(t, req, http.MethodPost, contractRefBase+"/branches/main/rollback")
		if req.URL.RawQuery != "" {
			t.Fatalf("query = %q, want none", req.URL.RawQuery)
		}
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content type = %q, want application/json", got)
		}
		if body := decodeJSONBody(t, req); !reflect.DeepEqual(body, map[string]any{"tag": "release-42"}) {
			t.Fatalf("body = %#v, want only the tag", body)
		}
	}
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			assertRollbackRequest(req)
			return jsonResponse(http.StatusOK, `{"branchId":"main-branch-id","headSnapshot":{"snapshotId":"rollback-snapshot-id","snapshotCommittedAt":1788336000000}}`)
		},
		func(req *http.Request) *http.Response {
			assertRollbackRequest(req)
			return jsonResponse(http.StatusOK, `{"branchId":"main-branch-id","headSnapshot":null}`)
		},
	}}
	collection := contractCollection(mock)
	input := lambdadb.RollbackBranchInput{Tag: "release-42"}

	restored, err := collection.Branches().Rollback(context.Background(), "main", input)
	if err != nil {
		t.Fatalf("Branches().Rollback() error = %v", err)
	}
	head := restored.GetHeadSnapshot()
	if restored.GetBranchID() != "main-branch-id" || head == nil || head.GetSnapshotID() != "rollback-snapshot-id" || head.GetSnapshotCommittedAt().UnixMilli() != 1788336000000 {
		t.Fatalf("rollback = %#v", restored)
	}

	empty, err := collection.Branches().Rollback(context.Background(), "main", input)
	if err != nil {
		t.Fatalf("Branches().Rollback() to a tag without a snapshot error = %v", err)
	}
	if empty.GetBranchID() != "main-branch-id" || empty.HeadSnapshot != nil || empty.GetHeadSnapshot() != nil {
		t.Fatalf("rollback without a snapshot = %#v, want nil HeadSnapshot", empty)
	}
	encoded, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"branchId":"main-branch-id","headSnapshot":null}` {
		t.Fatalf("encoded rollback = %s, want explicit null headSnapshot", encoded)
	}

	var nilResult *lambdadb.RollbackBranchResponse
	if nilResult.GetBranchID() != "" || nilResult.GetHeadSnapshot() != nil {
		t.Fatal("rollback getters must be nil-safe")
	}
	mock.assertDone()
}

// A repeated rollback can undo intervening writes, so Rollback sends exactly
// one request whatever the client or per-call retry configuration says, and
// preserves the server's error.
func TestVersioningContract_RollbackBranchIsNotRetried(t *testing.T) {
	backoff := retry.Config{
		Strategy:              "backoff",
		Backoff:               &retry.BackoffStrategy{InitialInterval: 1, MaxInterval: 2, Exponent: 1.1, MaxElapsedTime: 1000},
		RetryConnectionErrors: true,
	}
	for _, tc := range []struct {
		name    string
		status  int
		perCall bool
	}{
		{"conflict", http.StatusConflict, false},
		{"rate_limited", http.StatusTooManyRequests, false},
		{"unavailable", http.StatusServiceUnavailable, false},
		{"unavailable_with_per_call_retries", http.StatusServiceUnavailable, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
				func(req *http.Request) *http.Response {
					assertRequest(t, req, http.MethodPost, contractRefBase+"/branches/main/rollback")
					return jsonResponse(tc.status, `{"message":"server context"}`)
				},
			}}
			collection := lambdadb.New(
				lambdadb.WithBaseURL("https://api.example.com"),
				lambdadb.WithProjectName("project-versioning"),
				lambdadb.WithClient(mock),
				lambdadb.WithRetryConfig(backoff),
			).Collection("articles")
			opts := make([]operations.Option, 0, 4)
			if tc.perCall {
				opts = append(opts, operations.WithRetries(backoff))
			}
			_, err := collection.Branches().Rollback(context.Background(), "main", lambdadb.RollbackBranchInput{Tag: "release-42"}, opts...)
			if spare := opts[:cap(opts)]; spare[len(opts)] != nil {
				t.Fatal("Rollback wrote into the caller's option slice")
			}
			var conflict *apierrors.ResourceAlreadyExistsError
			var tooMany *apierrors.TooManyRequestsError
			var internal *apierrors.InternalServerError
			switch tc.status {
			case http.StatusConflict:
				if !errors.As(err, &conflict) {
					t.Fatalf("error = %T %v, want ResourceAlreadyExistsError", err, err)
				}
			case http.StatusTooManyRequests:
				if !errors.As(err, &tooMany) {
					t.Fatalf("error = %T %v, want TooManyRequestsError", err, err)
				}
			default:
				if !errors.As(err, &internal) || internal.HTTPMeta.Response.StatusCode != tc.status {
					t.Fatalf("error = %T %v, want InternalServerError with status %d", err, err, tc.status)
				}
			}
			mock.assertDone()
		})
	}
}

// A tag created before the first snapshot has explicit null snapshot fields
// and numDocs 0 in create, get and list responses. They decode to nil, the
// getters return the zero values, and they encode back as null.
func TestVersioningContract_TagBeforeFirstSnapshot(t *testing.T) {
	const pendingTag = `{"name":"pending","snapshotId":null,"snapshotCommittedAt":null,"numDocs":0,"dataUpdatedAt":null,"indexConfigs":` + contractSchema + `,"createdAt":1788336000123}`
	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusCreated, `{"tag":`+pendingTag+`}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"tag":`+pendingTag+`}`)
		},
		func(req *http.Request) *http.Response {
			return jsonResponse(http.StatusOK, `{"tags":[`+pendingTag+`,`+contractTag+`]}`)
		},
	}}
	collection := contractCollection(mock)
	ctx := context.Background()

	created, err := collection.Tags().Create(ctx, lambdadb.CreateTagInput{TagName: "pending"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := collection.Tags().Get(ctx, "pending")
	if err != nil {
		t.Fatal(err)
	}
	tags, err := collection.Tags().List(ctx)
	if err != nil || len(tags) != 2 {
		t.Fatalf("tags = %#v, %v", tags, err)
	}

	for _, tag := range []*lambdadb.TagDetails{created, got, &tags[0]} {
		if tag.SnapshotID != nil || tag.SnapshotCommittedAt != nil || tag.DataUpdatedAt != nil ||
			tag.GetSnapshotID() != "" || !tag.GetSnapshotCommittedAt().IsZero() || !tag.GetDataUpdatedAt().IsZero() ||
			tag.GetNumDocs() != 0 || tag.GetCreatedAt().UnixMilli() != 1788336000123 {
			t.Fatalf("tag before the first snapshot = %#v", tag)
		}
		assertContractSchema(t, tag.GetIndexConfigs())
	}
	published := &tags[1]
	if published.SnapshotID == nil || published.GetSnapshotID() != "snapshot-1" || published.SnapshotCommittedAt == nil ||
		published.GetSnapshotCommittedAt().UnixMilli() != 1788335940456 || published.GetDataUpdatedAt().UnixMilli() != contractDataMilli {
		t.Fatalf("published tag = %#v", published)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"snapshotId", "snapshotCommittedAt", "dataUpdatedAt"} {
		if value, exists := body[name]; !exists || value != nil {
			t.Fatalf("%s must encode as explicit null: %s", name, encoded)
		}
	}

	var nilTag *lambdadb.TagDetails
	if nilTag.GetSnapshotID() != "" || !nilTag.GetSnapshotCommittedAt().IsZero() || !nilTag.GetDataUpdatedAt().IsZero() {
		t.Fatal("tag getters must be nil-safe")
	}
	mock.assertDone()
}
