package components

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestIndexConfigsUnionRejectsEmbeddingWithoutManagedEmbeddingTrue(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr string
	}{
		{
			name: "embedding without managedEmbedding",
			payload: `{
				"type": "vector",
				"embedding": {
					"provider": "openai",
					"model": "text-embedding-3-small",
					"sourceField": "body"
				}
			}`,
			wantErr: "managedEmbedding=true is required when embedding config is provided",
		},
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
