package operations

import (
	"github.com/lambdadb/go-lambdadb/internal/utils"
	"github.com/lambdadb/go-lambdadb/models/components"
)

// ListCollectionsRequest holds optional query parameters for listing collections.
type ListCollectionsRequest struct {
	// Max number of collections to return at once.
	Size *int64 `queryParam:"style=form,explode=true,name=size"`
	// Next page token.
	PageToken *string `queryParam:"style=form,explode=true,name=pageToken"`
}

// ListCollectionsResponseBody - A list of collections matched with a projectName.
type ListCollectionsResponseBody struct {
	Collections   []components.CollectionResponse `json:"collections"`
	NextPageToken *string                        `json:"nextPageToken,omitzero"`
}

func (l *ListCollectionsResponseBody) GetCollections() []components.CollectionResponse {
	if l == nil {
		return []components.CollectionResponse{}
	}
	return l.Collections
}

func (l *ListCollectionsResponseBody) GetNextPageToken() *string {
	if l == nil {
		return nil
	}
	return l.NextPageToken
}

type ListCollectionsResponse struct {
	HTTPMeta components.HTTPMetadata `json:"-"`
	// A list of collections matched with a projectName.
	Object *ListCollectionsResponseBody
}

func (l ListCollectionsResponse) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(l, "", false)
}

func (l *ListCollectionsResponse) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &l, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (l *ListCollectionsResponse) GetHTTPMeta() components.HTTPMetadata {
	if l == nil {
		return components.HTTPMetadata{}
	}
	return l.HTTPMeta
}

func (l *ListCollectionsResponse) GetObject() *ListCollectionsResponseBody {
	if l == nil {
		return nil
	}
	return l.Object
}
