package lambdadb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
)

// Contract: lambdadb/docs@1714009a2659a18b26cb68193dd7a071fd6efb42 reference/api/openapi.json
// (SHA-256 604e6fdcf0e0a5af2c9ac8714e60d282d0ebfb9847a49371ff0452fac48be070).
// Every IndexConfigs member accepts an optional description that the SDK forwards
// unchanged: nil omits the key, "" is sent to clear it, and reads decode it only when set.
func TestFieldDescriptionIndexConfigMembers(t *testing.T) {
	embedding := components.EmbeddingConfig{Provider: components.EmbeddingConfigProviderOpenai, Model: "text-embedding-3-small", SourceField: "body"}
	cases := map[string]struct {
		union func(*string) components.IndexConfigsUnion
		get   func(components.IndexConfigsUnion) *string
		base  string
	}{
		"text": {
			union: func(d *string) components.IndexConfigsUnion {
				return components.CreateIndexConfigsUnionText(components.IndexConfigsText{Analyzers: []components.Analyzer{components.AnalyzerStandard}, Description: d})
			},
			get:  func(u components.IndexConfigsUnion) *string { return u.IndexConfigsText.GetDescription() },
			base: `{"type":"text","analyzers":["standard"]}`,
		},
		"vector": {
			union: func(d *string) components.IndexConfigsUnion {
				return components.CreateIndexConfigsUnionVector(components.IndexConfigsVector{Dimensions: 128, Description: d})
			},
			get:  func(u components.IndexConfigsUnion) *string { return u.IndexConfigsVector.GetDescription() },
			base: `{"type":"vector","dimensions":128,"similarity":"cosine"}`,
		},
		"native_embedding_vector": {
			union: func(d *string) components.IndexConfigsUnion {
				return components.CreateIndexConfigsUnionNativeEmbeddingVector(components.IndexConfigsNativeEmbeddingVector{Embedding: embedding, Description: d})
			},
			get: func(u components.IndexConfigsUnion) *string {
				return u.IndexConfigsNativeEmbeddingVector.GetDescription()
			},
			base: `{"type":"vector","embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"body"}}`,
		},
		"managed_embedding_vector": {
			union: func(d *string) components.IndexConfigsUnion {
				return components.CreateIndexConfigsUnionManagedEmbeddingVector(components.IndexConfigsManagedEmbeddingVector{Embedding: embedding, Description: d})
			},
			get: func(u components.IndexConfigsUnion) *string {
				return u.IndexConfigsManagedEmbeddingVector.GetDescription()
			},
			base: `{"type":"vector","managedEmbedding":true,"embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"body"}}`,
		},
		"keyword": {
			union: func(d *string) components.IndexConfigsUnion {
				return components.CreateIndexConfigsUnionKeyword(components.IndexConfigs{Description: d})
			},
			get:  func(u components.IndexConfigsUnion) *string { return u.IndexConfigs.GetDescription() },
			base: `{"type":"keyword"}`,
		},
		"object": {
			union: func(d *string) components.IndexConfigsUnion {
				return components.CreateIndexConfigsUnionObject(components.IndexConfigsObject{ObjectIndexConfigs: map[string]any{"city": map[string]any{"type": "keyword"}}, Description: d})
			},
			get:  func(u components.IndexConfigsUnion) *string { return u.IndexConfigsObject.GetDescription() },
			base: `{"type":"object","objectIndexConfigs":{"city":{"type":"keyword"}}}`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for _, mode := range []struct {
				name string
				in   *string
				want string
			}{
				{"absent", nil, tc.base},
				{"set", lambdadb.String("Book title as printed on the cover"), withDescription(t, tc.base, "Book title as printed on the cover")},
				{"cleared", lambdadb.String(""), withDescription(t, tc.base, "")},
			} {
				t.Run(mode.name, func(t *testing.T) {
					raw, err := json.Marshal(tc.union(mode.in))
					if err != nil {
						t.Fatal(err)
					}
					assertJSONEqual(t, raw, mode.want)
					var decoded components.IndexConfigsUnion
					if err := json.Unmarshal([]byte(mode.want), &decoded); err != nil {
						t.Fatal(err)
					}
					if got := tc.get(decoded); !reflect.DeepEqual(got, mode.in) {
						t.Fatalf("decoded description = %v, want %v", deref(got), deref(mode.in))
					}
				})
			}
			t.Run("null", func(t *testing.T) {
				var decoded components.IndexConfigsUnion
				if err := json.Unmarshal([]byte(withNullDescription(t, tc.base)), &decoded); err != nil {
					t.Fatal(err)
				}
				if got := tc.get(decoded); got != nil {
					t.Fatalf("null description decoded to %q, want nil", *got)
				}
			})
		})
	}
}

// Descriptions pass through the public create, read and update boundaries, including
// nested object fields, and a returned schema can be re-sent with descriptions intact.
func TestFieldDescriptionCollectionWire(t *testing.T) {
	const schema = `{
		"id":{"type":"keyword"},
		"title":{"type":"text","analyzers":["standard"],"description":"Book title as printed on the cover"},
		"year":{"type":"long"},
		"profile":{"type":"object","description":"Author profile","objectIndexConfigs":{
			"name":{"type":"keyword","description":"Author display name"},
			"address":{"type":"object","objectIndexConfigs":{"city":{"type":"keyword","description":"City of residence"}}}
		}}
	}`
	var want map[string]any
	if err := json.Unmarshal([]byte(schema), &want); err != nil {
		t.Fatal(err)
	}
	collection := `{"collection":{"collectionName":"books","indexConfigs":` + schema + `}}`
	cleared := map[string]any{"title": map[string]any{"type": "text", "analyzers": []any{"standard"}, "description": ""}}

	mock := &publicAPIMockClient{t: t, handlers: []func(*http.Request) *http.Response{
		func(req *http.Request) *http.Response {
			if got := decodeJSONBody(t, req)["indexConfigs"]; req.Method != http.MethodPost || !reflect.DeepEqual(got, want) {
				t.Fatalf("create %s indexConfigs = %#v, want %#v", req.Method, got, want)
			}
			return jsonResponse(201, collection)
		},
		func(req *http.Request) *http.Response { return jsonResponse(200, collection) },
		func(req *http.Request) *http.Response {
			if got := decodeJSONBody(t, req)["indexConfigs"]; req.Method != http.MethodPatch || !reflect.DeepEqual(got, want) {
				t.Fatalf("update %s indexConfigs = %#v, want %#v", req.Method, got, want)
			}
			return jsonResponse(200, collection)
		},
		func(req *http.Request) *http.Response {
			if got := decodeJSONBody(t, req)["indexConfigs"]; req.Method != http.MethodPatch || !reflect.DeepEqual(got, cleared) {
				t.Fatalf("clear %s indexConfigs = %#v, want %#v", req.Method, got, cleared)
			}
			return jsonResponse(200, `{"collection":{"collectionName":"books","indexConfigs":{"id":{"type":"keyword"},"title":{"type":"text","analyzers":["standard"]}}}}`)
		},
	}}
	client := lambdadb.New(lambdadb.WithClient(mock))
	ctx := context.Background()

	created := map[string]components.IndexConfigsUnion{
		"id":    components.CreateIndexConfigsUnionKeyword(components.IndexConfigs{}),
		"title": components.CreateIndexConfigsUnionText(components.IndexConfigsText{Analyzers: []components.Analyzer{components.AnalyzerStandard}, Description: lambdadb.String("Book title as printed on the cover")}),
		"year":  components.CreateIndexConfigsUnionLong(components.IndexConfigs{}),
		"profile": components.CreateIndexConfigsUnionObject(components.IndexConfigsObject{
			Description: lambdadb.String("Author profile"),
			ObjectIndexConfigs: map[string]any{
				"name":    map[string]any{"type": "keyword", "description": "Author display name"},
				"address": map[string]any{"type": "object", "objectIndexConfigs": map[string]any{"city": map[string]any{"type": "keyword", "description": "City of residence"}}},
			},
		}),
	}
	if _, err := client.Collections.Create(ctx, lambdadb.CreateCollectionOptions{CollectionName: "books", IndexConfigs: created}); err != nil {
		t.Fatal(err)
	}

	stored, err := client.Collection("books").Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := stored.IndexConfigs["title"].IndexConfigsText.GetDescription(); got == nil || *got != "Book title as printed on the cover" {
		t.Fatalf("title description = %v", deref(got))
	}
	if got := stored.IndexConfigs["year"].IndexConfigs.GetDescription(); got != nil {
		t.Fatalf("year description = %q, want nil", *got)
	}
	if got := stored.IndexConfigs["id"].IndexConfigs.GetDescription(); got != nil {
		t.Fatalf("id description = %q, want nil", *got)
	}
	profile := stored.IndexConfigs["profile"].IndexConfigsObject
	if got := profile.GetDescription(); got == nil || *got != "Author profile" {
		t.Fatalf("profile description = %v", deref(got))
	}
	name, _ := profile.ObjectIndexConfigs["name"].(map[string]any)
	if name["description"] != "Author display name" {
		t.Fatalf("nested name config = %#v", profile.ObjectIndexConfigs["name"])
	}
	if _, err := client.Collection("books").Update(ctx, lambdadb.UpdateCollectionOptions{IndexConfigs: stored.IndexConfigs}); err != nil {
		t.Fatal(err)
	}

	clearing := map[string]components.IndexConfigsUnion{
		"title": components.CreateIndexConfigsUnionText(components.IndexConfigsText{Analyzers: []components.Analyzer{components.AnalyzerStandard}, Description: lambdadb.String("")}),
	}
	updated, err := client.Collection("books").Update(ctx, lambdadb.UpdateCollectionOptions{IndexConfigs: clearing})
	if err != nil {
		t.Fatal(err)
	}
	if got := updated.IndexConfigs["title"].IndexConfigsText.GetDescription(); got != nil {
		t.Fatalf("cleared description decoded to %q, want nil", *got)
	}
	mock.assertDone()
}

func withDescription(t *testing.T, base, description string) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(base), &m); err != nil {
		t.Fatal(err)
	}
	m["description"] = description
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func withNullDescription(t *testing.T, base string) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(base), &m); err != nil {
		t.Fatal(err)
	}
	m["description"] = nil
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func assertJSONEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("invalid JSON %s: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("JSON = %s, want %s", got, want)
	}
}

func deref(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}
