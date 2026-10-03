package lambdadb_test

import (
	"context"

	lambdadb "github.com/lambdadb/go-lambdadb"
	"github.com/lambdadb/go-lambdadb/models/components"
	"github.com/lambdadb/go-lambdadb/optionalnullable"
)

// This example requires a collection with managed bodyEmbedding and stored title/body fields.
func ExampleCollection_Query_managedReranking() {
	collection := lambdadb.New().Collection("articles")
	config := components.RerankConfig{
		Provider: "typesafe", Model: "jev-1.13.0",
		QueryText: "How do I restore a previous collection version?",
		Fields:    []string{"title", "body"},
	}
	_, _ = collection.Query(context.Background(), lambdadb.QueryInput{
		Size: lambdadb.Int64(10),
		Query: map[string]any{"knn": map[string]any{
			"field": "bodyEmbedding", "queryText": config.QueryText, "k": 50,
		}},
		Rerank: optionalnullable.From(&config),
	})
}

func ExampleCollection_Query_customCriteria() {
	collection := lambdadb.New().Collection("articles")
	criteria := []string{
		"Does not explain how to restore a collection version.",
		"Explains part of the restore procedure but leaves required steps missing.",
		"Explains all steps needed to restore the requested collection version.",
	}
	config := components.RerankConfig{
		Provider: "typesafe", Model: "jev-1.13.0", QueryText: "restore a collection version",
		Fields: []string{"body"}, CandidateSize: lambdadb.Int64(50),
		OnFailure: lambdadb.String("returnOriginal"), Criteria: optionalnullable.From(&criteria),
	}
	_, _ = collection.Query(context.Background(), lambdadb.QueryInput{
		Size:   lambdadb.Int64(10),
		Query:  map[string]any{"queryString": map[string]any{"query": "body:restore"}},
		Rerank: optionalnullable.From(&config),
	})
}
