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
)

// Contract: lambdadb/docs@7be299af9cac4195b75feb6e38b5cd7f8989f5a0 reference/api/openapi.json
// (SHA-256 ec499bb6ad4fd7359c1854359e9ecfd4cd8609420f440f7dcbf75566ccda8f32, info.version 1.3.0),
// synced from lambdadb/lambdadb@f9aa63c79cbac2ac477a76ec45e5015ce665afd1.
// These tests exercise the public SDK wire boundary for the individual ref
// routes, branch schema updates, paginated ref listings and the ref fields
// added to every branch, tag and alias response.

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
