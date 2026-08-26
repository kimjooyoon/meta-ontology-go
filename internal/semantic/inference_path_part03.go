package semantic

// EvidenceReference has no producer or actor field. Its digest binds the
// reference to an append-only, producer-independent evidence record.
type EvidenceReference struct {
	ID     ID
	Digest string
}
type InferenceEvidence struct {
	ID           ID
	Digest       string
	Before       SnapshotDigests
	After        SnapshotDigests
	SourceBacked bool
	Independent  bool
	Controls     InferenceControls
}

// InferenceRecord contains the identity and proof tuple common to every edge
// and semantic-change claim. No display name, path, timestamp, or actor is in
// this tuple.
type InferenceRecord struct {
	RecordID  ID
	SubjectID ID
	ObjectID  ID
	Rule      RuleBinding
	Phase     PhasePlacement
	Before    SnapshotDigests
	After     SnapshotDigests
	Authority AuthorityBinding
	Evidence  []EvidenceReference
	Controls  InferenceControls
}
type InferenceEdge struct {
	InferenceRecord
	Kind              InferenceKind
	SourceRoots       []ID
	AcceptanceReceipt ID
}
type SemanticChangeClaim struct {
	InferenceRecord
	Kind           SemanticChangeKind
	CanonicalDelta string
	DeltaDigest    string
}

// InferencePathV1 is the finite, append-only semantic record set. It is a
// typed evidence carrier, not a graph database, scheduler, or PROV replacement.
type InferencePathV1 struct {
	Version  string
	Edges    []InferenceEdge
	Claims   []SemanticChangeClaim
	Evidence []InferenceEvidence
}
type InferencePathIssue struct {
	Code   string
	Record ID
	Detail string
}
type InferencePathErrors struct{ Issues []InferencePathIssue }
