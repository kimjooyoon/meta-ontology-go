package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// SemanticChangeRow is intentionally separate from InferenceRow. A semantic
// change claim is not an inference transition and cannot be queried as one.
type SemanticChangeRow struct {
	RecordID  ID `json:"record_id"`
	SubjectID ID `json:"subject_id"`
	ObjectID  ID `json:"object_id"`

	Kind            semantic.SemanticChangeKind `json:"semantic_change_kind"`
	Phase           semantic.InferencePhase     `json:"phase"`
	PhaseOrdinal    uint64                      `json:"phase_ordinal"`
	AuthorityLayer  semantic.AuthorityLayer     `json:"authority_layer"`
	AuthorityEffect semantic.AuthorityEffect    `json:"authority_effect"`

	Rule           semantic.RuleBinding         `json:"rule"`
	Before         semantic.SnapshotDigests     `json:"before"`
	After          semantic.SnapshotDigests     `json:"after"`
	Evidence       []semantic.EvidenceReference `json:"evidence"`
	CanonicalDelta string                       `json:"canonical_delta,omitempty"`
	DeltaDigest    string                       `json:"delta_digest,omitempty"`
}

// InferenceEvidenceRow exposes evidence records only when requested or
// selected by evidence ID. It never promotes an evidence record to authority.
type InferenceEvidenceRow struct {
	ID           ID                         `json:"id"`
	Digest       string                     `json:"digest"`
	Before       semantic.SnapshotDigests   `json:"before"`
	After        semantic.SnapshotDigests   `json:"after"`
	SourceBacked bool                       `json:"source_backed"`
	Independent  bool                       `json:"independent"`
	Controls     semantic.InferenceControls `json:"controls"`
}

// InferenceChainResult is one finite, typed explanation chain. Chain was
// accepted only after semantic.NewInferencePathChain returned successfully.
type InferenceChainResult struct {
	Chain    semantic.InferencePathChain `json:"chain"`
	Edges    []InferenceRow              `json:"edges"`
	Depth    int                         `json:"depth"`
	Complete bool                        `json:"complete"`
}

// InferenceWork makes the total budget accounting observable and replayable.
// Used is the sum of the category counters and never exceeds Limit on a
// successful query.
type InferenceWork struct {
	Limit             int `json:"limit"`
	Used              int `json:"used"`
	EdgesInspected    int `json:"edges_inspected"`
	ClaimsInspected   int `json:"claims_inspected"`
	EvidenceInspected int `json:"evidence_inspected"`
	ChainInspected    int `json:"chain_edges_inspected"`
}
