package operations

import (
	"github.com/lambdadb/go-lambdadb/internal/utils"
	"github.com/lambdadb/go-lambdadb/models/components"
)

type ListDocsRequest struct {
	// Collection name.
	CollectionName string `pathParam:"style=simple,explode=false,name=collectionName"`
	// Max number of documents to return at once.
	Size *int64 `queryParam:"style=form,explode=true,name=size"`
	// Next page token.
	PageToken *string `queryParam:"style=form,explode=true,name=pageToken"`
}

func (l *ListDocsRequest) GetCollectionName() string {
	if l == nil {
		return ""
	}
	return l.CollectionName
}

func (l *ListDocsRequest) GetSize() *int64 {
	if l == nil {
		return nil
	}
	return l.Size
}

func (l *ListDocsRequest) GetPageToken() *string {
	if l == nil {
		return nil
	}
	return l.PageToken
}

// ListDocsDoc - A single document in a list response.
type ListDocsDoc struct {
	Collection string         `json:"collection"`
	Doc        map[string]any `json:"doc"`
}

func (l *ListDocsDoc) GetCollection() string {
	if l == nil {
		return ""
	}
	return l.Collection
}

func (l *ListDocsDoc) GetDoc() map[string]any {
	if l == nil {
		return map[string]any{}
	}
	return l.Doc
}

// ListDocsResponseBody - Documents list.
type ListDocsResponseBody struct {
	Total int64 `json:"total"`
	// A list of documents.
	Docs          []ListDocsDoc `json:"docs"`
	NextPageToken *string          `json:"nextPageToken,omitzero"`
	// Whether the list of documents is included in the response.
	IsDocsInline bool `json:"isDocsInline"`
	// Download URL for the list of documents when not inline.
	DocsURL *string `json:"docsUrl,omitzero"`
}

func (l *ListDocsResponseBody) GetTotal() int64 {
	if l == nil {
		return 0
	}
	return l.Total
}

func (l *ListDocsResponseBody) GetDocs() []ListDocsDoc {
	if l == nil {
		return []ListDocsDoc{}
	}
	return l.Docs
}

func (l *ListDocsResponseBody) GetNextPageToken() *string {
	if l == nil {
		return nil
	}
	return l.NextPageToken
}

func (l *ListDocsResponseBody) GetIsDocsInline() bool {
	if l == nil {
		return false
	}
	return l.IsDocsInline
}

func (l *ListDocsResponseBody) GetDocsURL() *string {
	if l == nil {
		return nil
	}
	return l.DocsURL
}

type ListDocsResponse struct {
	HTTPMeta components.HTTPMetadata `json:"-"`
	// Documents list.
	Object *ListDocsResponseBody
}

func (l ListDocsResponse) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(l, "", false)
}

func (l *ListDocsResponse) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &l, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (l *ListDocsResponse) GetHTTPMeta() components.HTTPMetadata {
	if l == nil {
		return components.HTTPMetadata{}
	}
	return l.HTTPMeta
}

func (l *ListDocsResponse) GetObject() *ListDocsResponseBody {
	if l == nil {
		return nil
	}
	return l.Object
}
