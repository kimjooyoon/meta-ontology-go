package analyzer

const (
	EvidenceStatusDeterministic  EvidenceStatus = "deterministic"
	EvidenceStatusCandidate      EvidenceStatus = "candidate"
	EvidenceStatusImplementation EvidenceStatus = "implementation"
)

// EvidenceRecord is a host-neutral, append-only projection of one analyzer result.
type EvidenceRecord struct {
	Kind          EvidenceKind
	Status        EvidenceStatus
	Subject       Identity
	Relation      Relation
	Object        Identity
	Reference     string
	Options       []Identity
	Span          Span
	Reason        string
	IdentityState IdentityState
}

// Valid reports whether the record has the fields required for its kind.
func (e EvidenceRecord) Valid() bool {
	switch e.Kind {
	case EvidenceKindFact:
		return e.Status == EvidenceStatusDeterministic && e.Subject.Valid() && e.Object.Valid() && e.Relation != "" && evidenceSpanValid(e.Span)
	case EvidenceKindCandidate:
		return e.Status == EvidenceStatusCandidate && e.Subject.Valid() && e.Relation != "" && e.Reference != "" && validIdentityOptions(e.Options) && evidenceSpanValid(e.Span)
	case EvidenceKindImplementation:
		return e.Status == EvidenceStatusImplementation && e.Reference != "" &&
			e.IdentityState.valid() && evidenceSpanValid(e.Span)
	default:
		return false
	}
}

// EvidenceReport is the evidence emitted by one host stage. A deferred report
// has no comparison digest and cannot be mistaken for a successful run.
type EvidenceReport struct {
	Contract HostingContract
	Records  []EvidenceRecord
	Reason   string
}

// Complete reports whether the host contract is implemented and every record
// is structurally valid.
func (r EvidenceReport) Complete() bool {
	if !r.Contract.PromotionReady() {
		return false
	}
	for _, record := range r.Records {
		if !record.Valid() {
			return false
		}
	}
	return true
}
