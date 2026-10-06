package lambdadb_test

import (
	"context"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/optionalnullable"
)

func ExampleCollection_Query_bayesian() {
	collection := lambdadb.New(lambdadb.WithAPIKey("<YOUR_PROJECT_API_KEY>")).Collection("articles")
	query := map[string]any{"bayesian": []any{
		map[string]any{"queryString": map[string]any{"query": "body:restore"}},
		map[string]any{"knn": map[string]any{"field": "vector", "queryText": "restore a previous version", "k": 30}},
	}}
	_, _ = collection.Query(context.Background(), lambdadb.QueryInput{Query: query, Size: lambdadb.Int64(10), CandidateSize: lambdadb.Int64(30)})
	config := components.RerankConfig{Provider: "typesafe", Model: "jev-1.13.0", QueryText: "How do I restore a previous version?", Fields: []string{"body"}}
	_, _ = collection.Query(context.Background(), lambdadb.QueryInput{Query: query, Size: lambdadb.Int64(10), Rerank: optionalnullable.From(&config)})
}

func ExampleProjectCollections_Create_nativeEmbedding() {
	client := lambdadb.New(lambdadb.WithAPIKey("<YOUR_PROJECT_API_KEY>"))
	schema := map[string]components.IndexConfigsUnion{
		"body": components.CreateIndexConfigsUnionText(components.IndexConfigsText{}),
		"vector": components.CreateIndexConfigsUnionNativeEmbeddingVector(components.IndexConfigsNativeEmbeddingVector{
			Embedding: components.EmbeddingConfig{Provider: components.EmbeddingConfigProviderOpenai, Model: "text-embedding-3-small", SourceField: "body"},
		}),
	}
	_, _ = client.Collections.Create(context.Background(), lambdadb.CreateCollectionOptions{CollectionName: "articles", IndexConfigs: schema})
	_, _ = client.Collection("articles").Update(context.Background(), lambdadb.UpdateCollectionOptions{IndexConfigs: schema})
}
