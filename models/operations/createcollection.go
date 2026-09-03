package operations

import (
	"github.com/lambdadb/go-lambdadb/internal/utils"
	"github.com/lambdadb/go-lambdadb/models/components"
)

type CreateCollectionRequest struct {
	// Collection name must be unique within a project and the supported maximum length is 52.
	CollectionName string                                  `json:"collectionName"`
	IndexConfigs   map[string]components.IndexConfigsUnion `json:"indexConfigs"`
	// Optional collection description.
	Description *string `json:"description,omitzero"`
	// Collection metadata tags. The API accepts up to five entries.
	Tags            map[string]string           `json:"tags,omitzero"`
	PartitionConfig *components.PartitionConfig `json:"partitionConfig,omitzero"`
	// Number of days to retain committed snapshots. Defaults to 30.
	SnapshotRetentionInDays *int64 `json:"snapshotRetentionInDays,omitzero"`
}

func (c CreateCollectionRequest) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(c, "", false)
}

func (c *CreateCollectionRequest) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &c, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (c *CreateCollectionRequest) GetCollectionName() string {
	if c == nil {
		return ""
	}
	return c.CollectionName
}

func (c *CreateCollectionRequest) GetIndexConfigs() map[string]components.IndexConfigsUnion {
	if c == nil {
		return nil
	}
	return c.IndexConfigs
}

func (c *CreateCollectionRequest) GetPartitionConfig() *components.PartitionConfig {
	if c == nil {
		return nil
	}
	return c.PartitionConfig
}

func (c *CreateCollectionRequest) GetDescription() *string {
	if c == nil {
		return nil
	}
	return c.Description
}

func (c *CreateCollectionRequest) GetTags() map[string]string {
	if c == nil {
		return nil
	}
	return c.Tags
}

func (c *CreateCollectionRequest) GetSnapshotRetentionInDays() *int64 {
	if c == nil {
		return nil
	}
	return c.SnapshotRetentionInDays
}

// CreateCollectionResponseBody - Created collection
type CreateCollectionResponseBody struct {
	Collection components.CollectionResponse `json:"collection"`
}

func (c *CreateCollectionResponseBody) GetCollection() components.CollectionResponse {
	if c == nil {
		return components.CollectionResponse{}
	}
	return c.Collection
}

type CreateCollectionResponse struct {
	HTTPMeta components.HTTPMetadata `json:"-"`
	// Created collection
	Object *CreateCollectionResponseBody
}

func (c CreateCollectionResponse) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(c, "", false)
}

func (c *CreateCollectionResponse) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &c, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (c *CreateCollectionResponse) GetHTTPMeta() components.HTTPMetadata {
	if c == nil {
		return components.HTTPMetadata{}
	}
	return c.HTTPMeta
}

func (c *CreateCollectionResponse) GetObject() *CreateCollectionResponseBody {
	if c == nil {
		return nil
	}
	return c.Object
}
