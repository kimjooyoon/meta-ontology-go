package bidir

// BXStateEvidence records all transaction dimensions as digests.
type BXStateEvidence struct {
	Semantic string
	Source   string
	Region   string
	Slot     string
	Bytes    string
	LStat    string
}

// BXTransactionEvidence records an observed before/after transaction.
type BXTransactionEvidence struct {
	Before         BXStateEvidence
	After          BXStateEvidence
	ObserverKind   string
	Observed       bool
	Atomic         bool
	NoWrite        bool
	Deferred       bool
	DeferredReason string
}

// BXEvidenceSpanSet records evidence IDs and source-span cardinality.
type BXEvidenceSpanSet struct {
	IDs                 []string
	FactKeys            []string
	Spans               []SourceSpan
	Records             []BXEvidenceRecord
	IDCount             int
	SpanCount           int
	Hash                string
	EvidenceIDAuthority string
}

// BXDeltaEvidence records ordered delta and locality evidence.
type BXDeltaEvidence struct {
	SequenceHash          string
	OrderHash             string
	CanonicalJSON         string
	Added                 []string
	Removed               []string
	Locality              Locality
	LocalityClosureHash   string
	LocalityCanonicalJSON string
	ClosureMembers        []ID
	ClosureValid          bool
	Candidates            []string
	PortSequence          []string
	RelationSequence      []string
	PortOrderHash         string
	RelationOrderHash     string
	EvidenceSpans         BXEvidenceSpanSet
	EvidenceHash          string
	PartialObservation    bool
	RemovedCreated        bool
	CandidatePromoted     bool
}

// BXConflictEvidence records an expected partial-information rejection.
type BXConflictEvidence struct {
	Kind              ConflictKind
	Count             int
	Transactional     bool
	NoWriteObserved   bool
	RemovedCreated    bool
	CandidatePromoted bool
}
