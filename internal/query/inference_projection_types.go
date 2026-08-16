package query

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

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

	Before   semantic.SnapshotDigests   `json:"before,omitempty"`
	After    semantic.SnapshotDigests   `json:"after,omitempty"`
	Controls semantic.InferenceControls `json:"controls,omitempty"`

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

// InferenceQueryResult is a canonical, non-authoritative query receipt.
// Complete is false for every rejection or budget overrun; rejected results
// contain no partial rows.
type InferenceQueryResult struct {
	Schema      string                 `json:"schema"`
	Status      string                 `json:"status"`
	Request     InferenceQuery         `json:"request"`
	RequestHash string                 `json:"request_hash,omitempty"`
	Edges       []InferenceRow         `json:"edges,omitempty"`
	Claims      []SemanticChangeRow    `json:"claims,omitempty"`
	Evidence    []InferenceEvidenceRow `json:"evidence,omitempty"`
	Chain       *InferenceChainResult  `json:"chain,omitempty"`
	Work        InferenceWork          `json:"work"`
	Complete    bool                   `json:"complete"`
	Error       *EnvelopeError         `json:"error,omitempty"`
	Hash        string                 `json:"canonical_hash,omitempty"`
}

// InferenceResponse and InferenceResult are compatibility spellings for
// callers that use the existing envelope/result terminology.
type InferenceResponse = InferenceQueryResult
type InferenceResult = InferenceQueryResult

// InferenceProjection is a detached, read-only view of one validated and
// normalized InferencePathV1 snapshot.
type InferenceProjection struct {
	path semantic.InferencePathV1
}
