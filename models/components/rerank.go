package components

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/lambdadb/go-lambdadb/optionalnullable"
)

// RerankConfig selects server-managed reranking for one scoring search request.
// The server validates the input and applies defaults; the SDK does not rewrite
// size, knn.k, fields, or criteria. No provider API key is supplied by the caller.
type RerankConfig struct {
	// Required provider: typesafe.
	Provider string `json:"provider"`
	// Required model: jev-1.13.0.
	Model string `json:"model"`
	// Nonblank evaluation query, at most 8 KiB UTF-8.
	QueryText string `json:"queryText"`
	// 1-8 unique stored scalar text paths (text or keyword schema fields).
	Fields []string `json:"fields"`
	// Defaults to max(50, size); size must be positive and <= CandidateSize <= 100.
	CandidateSize *int64 `json:"candidateSize,omitempty"`
	// error (default) or returnOriginal; only eligible provider errors can fall back.
	OnFailure *string `json:"onFailure,omitempty"`
	// Omitted/null uses default criteria. Otherwise 2-10 distinct nonblank
	// descriptions, low to high relevance, at most 2 KiB each and 8 KiB total UTF-8.
	Criteria optionalnullable.OptionalNullable[[]string] `json:"criteria,omitempty"`
}

// UnmarshalJSON rejects unsupported options instead of silently discarding them
// when a JSON request is loaded into the typed SDK model.
func (r *RerankConfig) UnmarshalJSON(data []byte) error {
	type plain RerankConfig
	var value plain
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err != nil {
			return err
		}
		return errors.New("rerank must contain exactly one JSON value")
	}
	*r = RerankConfig(value)
	return nil
}

// RerankResponse describes one whole reranking stage. Counts are diagnostics,
// not billing units. The server returns no metadata when reranking is unused.
type RerankResponse struct {
	// applied, skipped, or fallback.
	Status         string  `json:"status"`
	Provider       string  `json:"provider"`
	Model          string  `json:"model"`
	ResolvedModel  *string `json:"resolvedModel,omitempty"`
	CandidateCount int64   `json:"candidateCount"`
	// Equals CandidateCount when applied; zero on skip/fallback.
	ScoredCount int64 `json:"scoredCount"`
	// Rerank-stage elapsed time in milliseconds.
	Took int64 `json:"took"`
	// Applied only: default-relevance-v1 or custom. custom is not a content hash.
	CriteriaVersion *string `json:"criteriaVersion,omitempty"`
	// Skipped: noCandidates. Fallback: timeout, rateLimit, unavailable,
	// invalidResponse, or credentials.
	Reason *string `json:"reason,omitempty"`
}
