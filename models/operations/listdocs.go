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
	// Set to true to include vector values in the response. Defaults to false.
	IncludeVectors *bool `default:"false" queryParam:"style=form,explode=true,name=includeVectors"`
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

func (l *ListDocsRequest) GetIncludeVectors() *bool {
	if l == nil {
		return nil
	}
	return l.IncludeVectors
}

type ListDocsExtendedRequestBody struct {
	// Max number of documents to return at once.
	Size *int64 `json:"size,omitzero"`
	// Next page token.
	PageToken *string `json:"pageToken,omitzero"`
	// Filter applied before pagination.
	Filter map[string]any `json:"filter,omitzero"`
	// Restricts the request to matching partition values.
	PartitionFilter *components.PartitionFilter `json:"partitionFilter,omitzero"`
	// An object to specify a list of field names to include and/or exclude in the result.
	Fields *components.FieldsSelectorUnion `json:"fields,omitzero"`
	// Set to true to include vector values in the response. Defaults to false.
	IncludeVectors *bool `default:"false" json:"includeVectors"`
}

func (l ListDocsExtendedRequestBody) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(l, "", false)
}

func (l *ListDocsExtendedRequestBody) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &l, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (l *ListDocsExtendedRequestBody) GetSize() *int64 {
	if l == nil {
		return nil
	}
	return l.Size
}

func (l *ListDocsExtendedRequestBody) GetPageToken() *string {
	if l == nil {
		return nil
	}
	return l.PageToken
}

func (l *ListDocsExtendedRequestBody) GetFilter() map[string]any {
	if l == nil {
		return nil
	}
	return l.Filter
}

func (l *ListDocsExtendedRequestBody) GetPartitionFilter() *components.PartitionFilter {
	if l == nil {
		return nil
	}
	return l.PartitionFilter
}

func (l *ListDocsExtendedRequestBody) GetFields() *components.FieldsSelectorUnion {
	if l == nil {
		return nil
	}
	return l.Fields
}

func (l *ListDocsExtendedRequestBody) GetIncludeVectors() *bool {
	if l == nil {
		return nil
	}
	return l.IncludeVectors
}

type ListDocsExtendedRequest struct {
	// Collection name.
	CollectionName string                      `pathParam:"style=simple,explode=false,name=collectionName"`
	Body           ListDocsExtendedRequestBody `request:"mediaType=application/json"`
}

func (l *ListDocsExtendedRequest) GetCollectionName() string {
	if l == nil {
		return ""
	}
	return l.CollectionName
}

func (l *ListDocsExtendedRequest) GetBody() ListDocsExtendedRequestBody {
	if l == nil {
		return ListDocsExtendedRequestBody{}
	}
	return l.Body
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
	NextPageToken *string       `json:"nextPageToken,omitzero"`
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
