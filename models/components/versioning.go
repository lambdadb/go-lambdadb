package components

import (
	"time"

	"github.com/lambdadb/go-lambdadb/types"
)

// RefKind identifies a branch, tag, or alias used for a read.
type RefKind string

const (
	RefKindBranch RefKind = "branch"
	RefKindTag    RefKind = "tag"
	RefKindAlias  RefKind = "alias"
)

// RefContext selects the collection ref used for a read operation.
type RefContext struct {
	Kind RefKind `json:"kind"`
	Name string  `json:"name"`
}

func (r *RefContext) GetKind() RefKind {
	if r == nil {
		return ""
	}
	return r.Kind
}

func (r *RefContext) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

// RefSourceKind identifies a branch or tag used as the source of a new ref.
type RefSourceKind string

const (
	RefSourceKindBranch RefSourceKind = "branch"
	RefSourceKindTag    RefSourceKind = "tag"
)

// RefSource selects the branch or tag from which a branch or tag is created.
// AsOf is valid only when Kind is RefSourceKindBranch.
type RefSource struct {
	Kind RefSourceKind `json:"kind"`
	Name string        `json:"name"`
	AsOf *int64        `json:"asOf,omitempty"`
}

func (r *RefSource) GetKind() RefSourceKind {
	if r == nil {
		return ""
	}
	return r.Kind
}

func (r *RefSource) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *RefSource) GetAsOf() *int64 {
	if r == nil {
		return nil
	}
	return r.AsOf
}

// RefDetails describes a branch or tag and its committed snapshot.
type RefDetails struct {
	Name       string              `json:"name"`
	SnapshotID *string             `json:"snapshotId"`
	CreatedAt  types.UnixMilliTime `json:"createdAt"`
}

func (r *RefDetails) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *RefDetails) GetSnapshotID() *string {
	if r == nil {
		return nil
	}
	return r.SnapshotID
}

// GetCreatedAt returns the ref creation time.
func (r *RefDetails) GetCreatedAt() time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.CreatedAt.Time
}

// AliasTarget selects a branch or tag for an alias.
type AliasTarget struct {
	Kind RefSourceKind `json:"kind"`
	Name string        `json:"name"`
}

func (a *AliasTarget) GetKind() RefSourceKind {
	if a == nil {
		return ""
	}
	return a.Kind
}

func (a *AliasTarget) GetName() string {
	if a == nil {
		return ""
	}
	return a.Name
}

// AliasTargetKind is the resolved target kind returned by the API.
type AliasTargetKind string

const (
	AliasTargetKindBranch AliasTargetKind = "BRANCH"
	AliasTargetKindTag    AliasTargetKind = "TAG"
)

// AliasDetails describes an alias and its current resolved target.
type AliasDetails struct {
	AliasID       string              `json:"aliasId"`
	AliasName     string              `json:"aliasName"`
	TargetKind    AliasTargetKind     `json:"targetKind"`
	TargetName    string              `json:"targetName"`
	TargetID      string              `json:"targetId"`
	AliasRevision int64               `json:"aliasRevision"`
	Dangling      bool                `json:"dangling"`
	CreatedAt     types.UnixMilliTime `json:"createdAt"`
}

func (a *AliasDetails) GetAliasID() string {
	if a == nil {
		return ""
	}
	return a.AliasID
}

func (a *AliasDetails) GetAliasName() string {
	if a == nil {
		return ""
	}
	return a.AliasName
}

func (a *AliasDetails) GetTargetKind() AliasTargetKind {
	if a == nil {
		return ""
	}
	return a.TargetKind
}

func (a *AliasDetails) GetTargetName() string {
	if a == nil {
		return ""
	}
	return a.TargetName
}

func (a *AliasDetails) GetTargetID() string {
	if a == nil {
		return ""
	}
	return a.TargetID
}

func (a *AliasDetails) GetAliasRevision() int64 {
	if a == nil {
		return 0
	}
	return a.AliasRevision
}

func (a *AliasDetails) GetDangling() bool {
	if a == nil {
		return false
	}
	return a.Dangling
}

// GetCreatedAt returns the alias creation time.
func (a *AliasDetails) GetCreatedAt() time.Time {
	if a == nil {
		return time.Time{}
	}
	return a.CreatedAt.Time
}
