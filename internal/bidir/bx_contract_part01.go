package bidir

// BXEvidenceSchemaVersion identifies the hard BX evidence contract.
const BXEvidenceSchemaVersion = "bidir-bx-evidence/v1"

// BXBaseEvidenceInput supplies the six required base artifacts.
type BXBaseEvidenceInput struct {
	DSL        Document
	IR         Model
	Go         FactSet
	SourceMap  []SourceSpan
	Evidence   FactSet
	Provenance []SourceSpan
}

// BXArtifactEvidence records one observed base artifact digest.
type BXArtifactEvidence struct {
	Hash  string
	Count int
}

// BXBaseEvidence is the measured form of the six base artifacts.
type BXBaseEvidence struct {
	DSL        BXArtifactEvidence
	IR         BXArtifactEvidence
	Go         BXArtifactEvidence
	SourceMap  BXArtifactEvidence
	Evidence   BXArtifactEvidence
	Provenance BXArtifactEvidence
}

// BXLStat is the observed metadata required for a source write boundary.
type BXLStat struct {
	Path   string
	Size   int64
	Mode   uint32
	Exists bool
}

// BXFileSnapshot is one observer-confirmed source state.
type BXFileSnapshot struct {
	Bytes []byte
	LStat BXLStat
}

// BXWriteObservation records before/after file observations.
type BXWriteObservation struct {
	Observed bool
	Before   BXFileSnapshot
	After    BXFileSnapshot
}

// BXRejectedWriteObserver owns before/after snapshots around a rejected
// operation. The producer supplies only the operation, never its snapshots.
type BXRejectedWriteObserver interface {
	Kind() string
	ObserveRejected(operation func() error) (BXWriteObservation, error)
}

// BXEvidenceFixture extends a reconciliation fixture with hard evidence.
type BXEvidenceFixture interface {
	ReconciliationFixture
	BaseEvidence() BXBaseEvidenceInput
	ObserveAcceptedWrite(before, after Document) BXWriteObservation
	RejectedWriteObserver(document Document) (BXRejectedWriteObserver, error)
}
