package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// InferenceQuery is an explicit bounded read request over a normalized
// semantic.InferencePathV1 snapshot. Empty typed selectors are unconstrained;
// non-empty selectors are combined with AND semantics.
type InferenceQuery struct {
	Schema    string `json:"schema"`
	Predicate string `json:"predicate,omitempty"`

	RecordID   ID `json:"record_id,omitempty"`
	SubjectID  ID `json:"subject_id,omitempty"`
	ObjectID   ID `json:"object_id,omitempty"`
	EvidenceID ID `json:"evidence_id,omitempty"`

	Kind      semantic.InferenceKind      `json:"kind,omitempty"`
	Phase     semantic.InferencePhase     `json:"phase,omitempty"`
	Layer     semantic.AuthorityLayer     `json:"authority_layer,omitempty"`
	Effect    semantic.AuthorityEffect    `json:"authority_effect,omitempty"`
	ClaimKind semantic.SemanticChangeKind `json:"semantic_change_kind,omitempty"`

	IncludeClaims   bool `json:"include_claims"`
	IncludeEvidence bool `json:"include_evidence"`
	Explain         bool `json:"explain"`
	ChainStartID    ID   `json:"chain_start_id,omitempty"`
	ChainEndID      ID   `json:"chain_end_id,omitempty"`

	Before   semantic.SnapshotDigests   `json:"before"`
	After    semantic.SnapshotDigests   `json:"after"`
	Controls semantic.InferenceControls `json:"controls"`

	Limit    int `json:"limit"`
	MaxDepth int `json:"max_depth"`
	MaxWork  int `json:"max_work"`
}

// InferenceRequest is the envelope-oriented spelling of InferenceQuery.
type InferenceRequest = InferenceQuery

// InferenceRow is a detached edge projection. It preserves stable IDs and
// all typed authority/evidence fields without exposing display metadata.
type InferenceRow struct {
	RecordID  ID `json:"record_id"`
	SubjectID ID `json:"subject_id"`
	ObjectID  ID `json:"object_id"`

	Kind            semantic.InferenceKind   `json:"kind"`
	Phase           semantic.InferencePhase  `json:"phase"`
	PhaseOrdinal    uint64                   `json:"phase_ordinal"`
	AuthorityLayer  semantic.AuthorityLayer  `json:"authority_layer"`
	AuthorityEffect semantic.AuthorityEffect `json:"authority_effect"`

	Rule              semantic.RuleBinding         `json:"rule"`
	Before            semantic.SnapshotDigests     `json:"before"`
	After             semantic.SnapshotDigests     `json:"after"`
	Evidence          []semantic.EvidenceReference `json:"evidence"`
	SourceRoots       []ID                         `json:"source_roots,omitempty"`
	AcceptanceReceipt ID                           `json:"acceptance_receipt,omitempty"`
}
