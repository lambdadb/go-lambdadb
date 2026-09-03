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
	CollectionName string                       `json:"collectionName"`
	IndexConfigs   map[string]IndexConfigsUnion `json:"indexConfigs"`
	// Collection description.
	Description string `json:"description"`
	// Collection metadata tags.
	Tags            map[string]string `json:"tags"`
	PartitionConfig *PartitionConfig  `json:"partitionConfig,omitzero"`
	// Total number of partitions including the default partition.
	NumPartitions int64 `json:"numPartitions"`
	// Total number of documents.
	NumDocs int64 `json:"numDocs"`
	// Default writable branch. The current API always returns main.
	DefaultBranchName string `json:"defaultBranchName"`
	// Number of days committed snapshots are retained.
	SnapshotRetentionInDays int64 `json:"snapshotRetentionInDays"`
	// Collection creation time as Unix epoch milliseconds, exposed as time.
	CreatedAt types.UnixMilliTime `json:"createdAt"`
	// Collection last update time as Unix epoch milliseconds, exposed as time.
	UpdatedAt types.UnixMilliTime `json:"updatedAt"`
	// Collection data last update time as Unix epoch milliseconds, exposed as time.
	DataUpdatedAt types.UnixMilliTime `json:"dataUpdatedAt,omitzero"`
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

func (c *CollectionResponse) GetDescription() string {
	if c == nil {
		return ""
	}
	return c.Description
}

func (c *CollectionResponse) GetTags() map[string]string {
	if c == nil {
		return nil
	}
	return c.Tags
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

func (c *CollectionResponse) GetDefaultBranchName() string {
	if c == nil {
		return ""
	}
	return c.DefaultBranchName
}

func (c *CollectionResponse) GetSnapshotRetentionInDays() int64 {
	if c == nil {
		return 0
	}
	return c.SnapshotRetentionInDays
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
