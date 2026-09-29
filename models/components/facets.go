package components

// FacetRequest selects up to Size keyword buckets (default 10, range 1..100).
type FacetRequest struct {
	Size *int64 `json:"size,omitempty"`
}

// FacetBucket counts matching documents containing an indexed keyword value.
type FacetBucket struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// FacetResult contains buckets ordered by count descending, then value ascending.
type FacetResult struct {
	Buckets []FacetBucket `json:"buckets"`
}
