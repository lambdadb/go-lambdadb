package components

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestIndexConfigsUnionRejectsEmbeddingWithManagedEmbeddingFalse(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name: "embedding with managedEmbedding false",
			payload: `{
				"type": "vector",
				"managedEmbedding": false,
				"embedding": {
					"provider": "openai",
					"model": "text-embedding-3-small",
					"sourceField": "body"
				}
			}`,
			wantErr: "embedding is not allowed when managedEmbedding=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var union IndexConfigsUnion
			err := json.Unmarshal([]byte(tt.payload), &union)
			if err == nil {
				t.Fatal("json.Unmarshal() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("json.Unmarshal() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestIndexConfigsManagedEmbeddingVectorRejectsTopLevelVectorSettings(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name: "dimensions",
			payload: `{
				"type": "vector",
				"managedEmbedding": true,
				"dimensions": 1536,
				"embedding": {
					"provider": "openai",
					"model": "text-embedding-3-small",
					"sourceField": "body"
				}
			}`,
			wantErr: "Top-level dimensions are not allowed for managed embedding field",
		},
		{
			name: "similarity",
			payload: `{
				"type": "vector",
				"managedEmbedding": true,
				"similarity": "cosine",
				"embedding": {
					"provider": "openai",
					"model": "text-embedding-3-small",
					"sourceField": "body"
				}
			}`,
			wantErr: "Top-level similarity is not allowed for managed embedding field",
		},
		{
			name: "missing embedding",
			payload: `{
				"type": "vector",
				"managedEmbedding": true
			}`,
			wantErr: "embedding is required when managedEmbedding=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var union IndexConfigsUnion
			err := json.Unmarshal([]byte(tt.payload), &union)
			if err == nil {
				t.Fatal("json.Unmarshal() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("json.Unmarshal() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestIndexConfigsUnionRejectsInvalidUnmanagedVectorSettings(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name: "missing dimensions",
			payload: `{
				"type": "vector"
			}`,
			wantErr: "Dimensions is required field",
		},
		{
			name: "unsupported similarity",
			payload: `{
				"type": "vector",
				"dimensions": 1536,
				"similarity": "unsupported"
			}`,
			wantErr: "unsupported is not a supported similarity metric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var union IndexConfigsUnion
			err := json.Unmarshal([]byte(tt.payload), &union)
			if err == nil {
				t.Fatal("json.Unmarshal() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("json.Unmarshal() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestIndexConfigsManagedEmbeddingVectorMarshalForcesManagedEmbeddingTrue(t *testing.T) {
	union := CreateIndexConfigsUnionManagedEmbeddingVector(IndexConfigsManagedEmbeddingVector{
		ManagedEmbedding: false,
		Embedding: EmbeddingConfig{
			Provider:    EmbeddingConfigProviderOpenai,
			Model:       "text-embedding-3-small",
			SourceField: "body",
		},
	})

	data, err := json.Marshal(union)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("json.Unmarshal(marshaled) error = %v", err)
	}
	if got := body["managedEmbedding"]; got != true {
		t.Fatalf("managedEmbedding = %v, want true; body = %s", got, string(data))
	}
	if _, ok := body["dimensions"]; ok {
		t.Fatalf("top-level dimensions was marshaled: %s", string(data))
	}
	if _, ok := body["similarity"]; ok {
		t.Fatalf("top-level similarity was marshaled: %s", string(data))
	}
}

func TestIndexConfigsVectorMarshalRejectsInvalidUnmanagedSettings(t *testing.T) {
	tests := []struct {
		name    string
		vector  IndexConfigsVector
		wantErr string
	}{
		{
			name:    "missing dimensions",
			vector:  IndexConfigsVector{},
			wantErr: "Dimensions is required field",
		},
		{
			name: "unsupported similarity",
			vector: IndexConfigsVector{
				Dimensions: 1536,
				Similarity: func() *Similarity {
					similarity := Similarity("unsupported")
					return &similarity
				}(),
			},
			wantErr: "unsupported is not a supported similarity metric",
		},
		{
			name: "managed embedding true on unmanaged vector",
			vector: IndexConfigsVector{
				ManagedEmbedding: func() *bool {
					managedEmbedding := true
					return &managedEmbedding
				}(),
				Dimensions: 1536,
			},
			wantErr: "managedEmbedding=true requires IndexConfigsManagedEmbeddingVector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := json.Marshal(CreateIndexConfigsUnionVector(tt.vector))
			if err == nil {
				t.Fatal("json.Marshal() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("json.Marshal() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestNativeEmbeddingRoundTrip(t *testing.T) {
	for _, extra := range []string{"", `,"dimensions":512,"similarity":"dot_product"`} {
		raw := `{"type":"vector","embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"body"` + extra + `}}`
		var union IndexConfigsUnion
		if err := json.Unmarshal([]byte(raw), &union); err != nil {
			t.Fatal(err)
		}
		if union.IndexConfigsNativeEmbeddingVector == nil {
			t.Fatal("native input not selected")
		}
		encoded, err := json.Marshal(union)
		if err != nil {
			t.Fatal(err)
		}
		var got, want map[string]any
		if err := json.Unmarshal(encoded, &got); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(raw), &want); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("roundtrip = %s, want %s", encoded, raw)
		}
	}
	for _, flag := range []*bool{nil, func() *bool { v := true; return &v }()} {
		vector := CreateIndexConfigsUnionNativeEmbeddingVector(IndexConfigsNativeEmbeddingVector{ManagedEmbedding: flag, Embedding: EmbeddingConfig{Provider: EmbeddingConfigProviderOpenai, Model: "text-embedding-3-small", SourceField: "body"}})
		raw, err := json.Marshal(vector)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}
		if _, ok := got["managedEmbedding"]; ok != (flag != nil) {
			t.Fatalf("flag presence: %s", raw)
		}
	}
}

func TestNativeEmbeddingRejectsConflicts(t *testing.T) {
	for _, extra := range []string{`,"dimensions":1536`, `,"similarity":"cosine"`, `,"managedEmbedding":false`} {
		raw := `{"type":"vector","embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"body"}` + extra + `}`
		var union IndexConfigsUnion
		if err := json.Unmarshal([]byte(raw), &union); err == nil {
			t.Fatalf("accepted conflict: %s", raw)
		}
		var vector IndexConfigsNativeEmbeddingVector
		if err := json.Unmarshal([]byte(raw), &vector); err == nil {
			t.Fatalf("accepted direct conflict: %s", raw)
		}
	}
	no := false
	_, err := json.Marshal(CreateIndexConfigsUnionNativeEmbeddingVector(IndexConfigsNativeEmbeddingVector{ManagedEmbedding: &no}))
	if err == nil {
		t.Fatal("accepted explicit false")
	}
	var vector IndexConfigsNativeEmbeddingVector
	if err := json.Unmarshal([]byte(`{"type":"vector"}`), &vector); err == nil {
		t.Fatal("accepted missing embedding")
	}
}

func TestNativeEmbeddingReplacesReusedUnion(t *testing.T) {
	const native = `{"type":"vector","embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"body"}}`
	for name, previous := range map[string]string{
		"text":    `{"type":"text"}`,
		"vector":  `{"type":"vector","dimensions":2}`,
		"managed": `{"type":"vector","managedEmbedding":true,"embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"oldBody"}}`,
		"keyword": `{"type":"keyword"}`,
		"object":  `{"type":"object","objectIndexConfigs":{}}`,
	} {
		t.Run(name, func(t *testing.T) {
			var union IndexConfigsUnion
			if err := json.Unmarshal([]byte(previous), &union); err != nil {
				t.Fatal(err)
			}
			before := union
			invalid := `{"type":"vector","dimensions":2,"embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"body"}}`
			if err := json.Unmarshal([]byte(invalid), &union); err == nil {
				t.Fatal("accepted conflicting native dimensions")
			}
			if !reflect.DeepEqual(union, before) {
				t.Fatal("failed decode changed the existing union")
			}
			if err := json.Unmarshal([]byte(native), &union); err != nil {
				t.Fatal(err)
			}
			want := CreateIndexConfigsUnionNativeEmbeddingVector(IndexConfigsNativeEmbeddingVector{
				Embedding: EmbeddingConfig{Provider: EmbeddingConfigProviderOpenai, Model: "text-embedding-3-small", SourceField: "body"},
			})
			if !reflect.DeepEqual(union, want) {
				t.Fatal("native decode retained a stale union member")
			}
			raw, err := json.Marshal(union)
			if err != nil {
				t.Fatal(err)
			}
			var gotJSON, wantJSON map[string]any
			if err := json.Unmarshal(raw, &gotJSON); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(native), &wantJSON); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotJSON, wantJSON) {
				t.Fatalf("reused union serialized as %s", raw)
			}
		})
	}
}

func TestIndexConfigsUnionReplacesNativeWithOtherVariants(t *testing.T) {
	const native = `{"type":"vector","embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"body"}}`
	for name, next := range map[string]string{
		"text":         `{"type":"text"}`,
		"vector":       `{"type":"vector","dimensions":2,"similarity":"cosine"}`,
		"managed":      `{"type":"vector","managedEmbedding":true,"embedding":{"provider":"openai","model":"text-embedding-3-small","sourceField":"other"}}`,
		"keyword":      `{"type":"keyword"}`,
		"long":         `{"type":"long"}`,
		"double":       `{"type":"double"}`,
		"datetime":     `{"type":"datetime"}`,
		"boolean":      `{"type":"boolean"}`,
		"sparseVector": `{"type":"sparseVector"}`,
		"object":       `{"type":"object","objectIndexConfigs":{}}`,
	} {
		t.Run(name, func(t *testing.T) {
			var union IndexConfigsUnion
			if err := json.Unmarshal([]byte(native), &union); err != nil {
				t.Fatal(err)
			}
			before := union
			if err := json.Unmarshal([]byte(`{"type":"vector","dimensions":0}`), &union); err == nil {
				t.Fatal("accepted invalid vector dimensions")
			}
			if !reflect.DeepEqual(union, before) {
				t.Fatal("failed decode changed the existing native value")
			}
			if err := json.Unmarshal([]byte(next), &union); err != nil {
				t.Fatal(err)
			}
			if union.IndexConfigsNativeEmbeddingVector != nil {
				t.Error("stale native member retained")
			}
			raw, err := json.Marshal(union)
			if err != nil {
				t.Fatal(err)
			}
			var gotJSON, wantJSON map[string]any
			if err := json.Unmarshal(raw, &gotJSON); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(next), &wantJSON); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotJSON, wantJSON) {
				t.Fatalf("reused union = %s, want %s", raw, next)
			}
		})
	}
}
