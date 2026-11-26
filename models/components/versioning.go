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

// RefSource selects a source in the same collection. Branch and tag creation
// both accept a branch or tag source. AsOf is valid only when Kind is
// RefSourceKindBranch; the server rejects it on a tag source with 400.
type RefSource struct {
	Kind RefSourceKind `json:"kind"`
	Name string        `json:"name"`
	// Selects the latest retained committed snapshot at or before this Unix
	// epoch millisecond cutoff. Inherited and pending writes are excluded; this
	// is not an exact-time restore.
	AsOf *int64 `json:"asOf,omitempty"`
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

// BranchDetails describes a writable branch, its snapshot lineage, its index
// schema, and the data statistics of its committed head.
type BranchDetails struct {
	Name string `json:"name"`
	// Direct source branch, even for an empty source or an ancestor snapshot
	// selected through AsOf. Nil for main, for a branch created from a tag, or
	// when no parent was recorded.
	ParentBranch *ParentBranchDetails `json:"parentBranch"`
	// Current committed head; nil for an empty branch.
	HeadSnapshot *SnapshotDetails `json:"headSnapshot"`
	// Fixed creation source, not the previous head. Nil for main and branches
	// created from an empty source, even after their head advances. This metadata
	// does not extend snapshot retention.
	ParentSnapshot *SnapshotDetails `json:"parentSnapshot"`
	// Document count in the branch's committed head; 0 for an empty branch.
	NumDocs int64 `json:"numDocs"`
	// Last data update of the branch's committed head; nil for an empty branch.
	DataUpdatedAt *types.UnixMilliTime `json:"dataUpdatedAt"`
	// The branch's own index schema. A branch inherits it at creation and later
	// updates of other branches, including main, do not change it.
	IndexConfigs map[string]IndexConfigsUnion `json:"indexConfigs"`
	CreatedAt    types.UnixMilliTime          `json:"createdAt"`
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

func (r *BranchDetails) GetNumDocs() int64 {
	if r == nil {
		return 0
	}
	return r.NumDocs
}

// GetDataUpdatedAt returns the last data update of the branch's committed
// head, or the zero time for an empty branch.
func (r *BranchDetails) GetDataUpdatedAt() time.Time {
	if r == nil || r.DataUpdatedAt == nil {
		return time.Time{}
	}
	return r.DataUpdatedAt.Time
}

func (r *BranchDetails) GetIndexConfigs() map[string]IndexConfigsUnion {
	if r == nil {
		return nil
	}
	return r.IndexConfigs
}

func (r *BranchDetails) GetCreatedAt() time.Time {
	if r == nil {
		return time.Time{}
	}
	return r.CreatedAt.Time
}

// RollbackBranchResponse is the committed result of a branch rollback.
type RollbackBranchResponse struct {
	// Identity of the rolled-back branch, which keeps its identity.
	BranchID string `json:"branchId"`
	// New head committed on the branch by this rollback; nil when the tag had
	// no published snapshot, which leaves the branch head empty.
	HeadSnapshot *SnapshotDetails `json:"headSnapshot"`
}

func (r *RollbackBranchResponse) GetBranchID() string {
	if r == nil {
		return ""
	}
	return r.BranchID
}

func (r *RollbackBranchResponse) GetHeadSnapshot() *SnapshotDetails {
	if r == nil {
		return nil
	}
	return r.HeadSnapshot
}

// TagDetails describes a fixed, read-only tag: its index schema and the
// statistics of its current published snapshot. A tag can be created before
// its first snapshot is published; SnapshotID, SnapshotCommittedAt and
// DataUpdatedAt are then nil and NumDocs is 0.
type TagDetails struct {
	Name string `json:"name"`
	// Current published snapshot; nil before the first publication. It may
	// change as the writes the tag captured are published.
	SnapshotID *string `json:"snapshotId"`
	// Current snapshot commit time, independent of tag creation time; nil
	// before the first publication.
	SnapshotCommittedAt *types.UnixMilliTime `json:"snapshotCommittedAt"`
	// Document count in the current published snapshot; 0 before the first
	// publication.
	NumDocs int64 `json:"numDocs"`
	// Last data update in the current snapshot; nil before the first
	// publication.
	DataUpdatedAt *types.UnixMilliTime `json:"dataUpdatedAt"`
	// Index schema pinned by the tag. It does not change when branches are
	// updated afterwards.
	IndexConfigs map[string]IndexConfigsUnion `json:"indexConfigs"`
	CreatedAt    types.UnixMilliTime          `json:"createdAt"`
}

func (r *TagDetails) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

// GetSnapshotID returns the current published snapshot ID, or "" before the
// first publication.
func (r *TagDetails) GetSnapshotID() string {
	if r == nil || r.SnapshotID == nil {
		return ""
	}
	return *r.SnapshotID
}

// GetSnapshotCommittedAt returns the current snapshot commit time, or the zero
// time before the first publication.
func (r *TagDetails) GetSnapshotCommittedAt() time.Time {
	if r == nil || r.SnapshotCommittedAt == nil {
		return time.Time{}
	}
	return r.SnapshotCommittedAt.Time
}

func (r *TagDetails) GetNumDocs() int64 {
	if r == nil {
		return 0
	}
	return r.NumDocs
}

// GetDataUpdatedAt returns the last data update in the current snapshot, or
// the zero time before the first publication.
func (r *TagDetails) GetDataUpdatedAt() time.Time {
	if r == nil || r.DataUpdatedAt == nil {
		return time.Time{}
	}
	return r.DataUpdatedAt.Time
}

func (r *TagDetails) GetIndexConfigs() map[string]IndexConfigsUnion {
	if r == nil {
		return nil
	}
	return r.IndexConfigs
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

// AliasDetails describes an alias, its current resolved target, and the index
// schema reads through the alias see.
type AliasDetails struct {
	AliasID       string          `json:"aliasId"`
	AliasName     string          `json:"aliasName"`
	TargetKind    AliasTargetKind `json:"targetKind"`
	TargetName    string          `json:"targetName"`
	TargetID      string          `json:"targetId"`
	AliasRevision int64           `json:"aliasRevision"`
	// Whether the bound target identity is missing. Document reads through a
	// dangling alias return 400; reading the alias itself still succeeds.
	Dangling bool `json:"dangling"`
	// Index schema of the target: a branch's current schema or a tag's pinned
	// schema. Nil for a dangling alias.
	IndexConfigs map[string]IndexConfigsUnion `json:"indexConfigs"`
	CreatedAt    types.UnixMilliTime          `json:"createdAt"`
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

// GetIndexConfigs returns the target's index schema, or nil for a dangling
// alias.
func (a *AliasDetails) GetIndexConfigs() map[string]IndexConfigsUnion {
	if a == nil {
		return nil
	}
	return a.IndexConfigs
}

// GetCreatedAt returns the alias creation time.
func (a *AliasDetails) GetCreatedAt() time.Time {
	if a == nil {
		return time.Time{}
	}
	return a.CreatedAt.Time
}
