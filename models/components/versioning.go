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

// RefContext selects the collection ref used for a read operation. A ref that
// does not exist returns ResourceNotFoundError; an alias whose target is
// dangling returns BadRequestError.
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

// RefSource selects a source in the same collection. Branch creation accepts
// only RefSourceKindBranch; tag creation accepts a branch or tag source.
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

// SnapshotDetails identifies an immutable committed snapshot.
type SnapshotDetails struct {
	SnapshotID string `json:"snapshotId"`
	// Snapshot commit time, independent of ref creation time and collection dataUpdatedAt.
	SnapshotCommittedAt types.UnixMilliTime `json:"snapshotCommittedAt"`
}

func (r *SnapshotDetails) GetSnapshotID() string {
	if r == nil {
		return ""
	}
	return r.SnapshotID
}

func (r *SnapshotDetails) GetSnapshotCommittedAt() time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.SnapshotCommittedAt.Time
}

// ParentBranchDetails records the direct source branch at creation time.
// This historical identity survives parent deletion or reuse of its name.
type ParentBranchDetails struct {
	BranchID string `json:"branchId"`
	Name     string `json:"name"`
}

func (r *ParentBranchDetails) GetBranchID() string {
	if r == nil {
		return ""
	}
	return r.BranchID
}

func (r *ParentBranchDetails) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

// BranchDetails describes a writable branch and its snapshot lineage.
type BranchDetails struct {
	Name string `json:"name"`
	// Direct source branch, even for an empty source or an ancestor snapshot
	// selected through AsOf. Nil for main or when no parent was recorded.
	ParentBranch *ParentBranchDetails `json:"parentBranch"`
	// Current committed head; nil for an empty branch.
	HeadSnapshot *SnapshotDetails `json:"headSnapshot"`
	// Fixed creation source, not the previous head. Nil for main and branches
	// created from an empty source, even after their head advances. This metadata
	// does not extend snapshot retention.
	ParentSnapshot *SnapshotDetails    `json:"parentSnapshot"`
	CreatedAt      types.UnixMilliTime `json:"createdAt"`
}

func (r *BranchDetails) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *BranchDetails) GetParentBranch() *ParentBranchDetails {
	if r == nil {
		return nil
	}
	return r.ParentBranch
}

func (r *BranchDetails) GetHeadSnapshot() *SnapshotDetails {
	if r == nil {
		return nil
	}
	return r.HeadSnapshot
}

func (r *BranchDetails) GetParentSnapshot() *SnapshotDetails {
	if r == nil {
		return nil
	}
	return r.ParentSnapshot
}

func (r *BranchDetails) GetCreatedAt() time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.CreatedAt.Time
}

// TagDetails describes an immutable tag pinning a nonempty committed snapshot.
type TagDetails struct {
	Name       string `json:"name"`
	SnapshotID string `json:"snapshotId"`
	// Pinned snapshot commit time, independent of tag creation time.
	SnapshotCommittedAt types.UnixMilliTime `json:"snapshotCommittedAt"`
	CreatedAt           types.UnixMilliTime `json:"createdAt"`
}

func (r *TagDetails) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *TagDetails) GetSnapshotID() string {
	if r == nil {
		return ""
	}
	return r.SnapshotID
}

func (r *TagDetails) GetSnapshotCommittedAt() time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.SnapshotCommittedAt.Time
}

func (r *TagDetails) GetCreatedAt() time.Time {
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
