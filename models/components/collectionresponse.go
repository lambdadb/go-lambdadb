package components

import (
	"time"

	"github.com/lambdadb/go-lambdadb/internal/utils"
	"github.com/lambdadb/go-lambdadb/types"
)

type CollectionResponse struct {
	// Project name.
	ProjectName string `json:"projectName"`
	// Collection name.
	CollectionName  string                       `json:"collectionName"`
	IndexConfigs    map[string]IndexConfigsUnion `json:"indexConfigs"`
	PartitionConfig *PartitionConfig             `json:"partitionConfig,omitzero"`
	// Total number of partitions including the default partition.
	NumPartitions int64 `json:"numPartitions"`
	// Total number of documents.
	NumDocs int64 `json:"numDocs"`
	// Source project name.
	SourceProjectName *string `json:"sourceProjectName,omitzero"`
	// Source collection name.
	SourceCollectionName *string `json:"sourceCollectionName,omitzero"`
	// Source collection version.
	SourceCollectionVersionID *string `json:"sourceCollectionVersionId,omitzero"`
	// Status
	CollectionStatus Status `json:"collectionStatus"`
	// Collection creation time (seconds since Unix epoch in API; exposed as time).
	CreatedAt types.UnixTime `json:"createdAt"`
	// Collection last update time (seconds since Unix epoch in API; exposed as time).
	UpdatedAt types.UnixTime `json:"updatedAt"`
	// Collection data last update time (seconds since Unix epoch in API; exposed as time).
	DataUpdatedAt types.UnixTime `json:"dataUpdatedAt"`
}

func (c CollectionResponse) MarshalJSON() ([]byte, error) {
	return utils.MarshalJSON(c, "", false)
}

func (c *CollectionResponse) UnmarshalJSON(data []byte) error {
	if err := utils.UnmarshalJSON(data, &c, "", false, nil); err != nil {
		return err
	}
	return nil
}

func (c *CollectionResponse) GetProjectName() string {
	if c == nil {
		return ""
	}
	return c.ProjectName
}

func (c *CollectionResponse) GetCollectionName() string {
	if c == nil {
		return ""
	}
	return c.CollectionName
}

func (c *CollectionResponse) GetIndexConfigs() map[string]IndexConfigsUnion {
	if c == nil {
		return map[string]IndexConfigsUnion{}
	}
	return c.IndexConfigs
}

func (c *CollectionResponse) GetPartitionConfig() *PartitionConfig {
	if c == nil {
		return nil
	}
	return c.PartitionConfig
}

func (c *CollectionResponse) GetNumPartitions() int64 {
	if c == nil {
		return 0
	}
	return c.NumPartitions
}

func (c *CollectionResponse) GetNumDocs() int64 {
	if c == nil {
		return 0
	}
	return c.NumDocs
}

func (c *CollectionResponse) GetSourceProjectName() *string {
	if c == nil {
		return nil
	}
	return c.SourceProjectName
}

func (c *CollectionResponse) GetSourceCollectionName() *string {
	if c == nil {
		return nil
	}
	return c.SourceCollectionName
}

func (c *CollectionResponse) GetSourceCollectionVersionID() *string {
	if c == nil {
		return nil
	}
	return c.SourceCollectionVersionID
}

func (c *CollectionResponse) GetCollectionStatus() Status {
	if c == nil {
		return Status("")
	}
	return c.CollectionStatus
}

// GetCreatedAt returns the collection creation time.
func (c *CollectionResponse) GetCreatedAt() time.Time {
	if c == nil {
		return time.Time{}
	}
	return c.CreatedAt.Time
}

// GetUpdatedAt returns the collection last update time.
func (c *CollectionResponse) GetUpdatedAt() time.Time {
	if c == nil {
		return time.Time{}
	}
	return c.UpdatedAt.Time
}

// GetDataUpdatedAt returns the collection data last update time.
func (c *CollectionResponse) GetDataUpdatedAt() time.Time {
	if c == nil {
		return time.Time{}
	}
	return c.DataUpdatedAt.Time
}
