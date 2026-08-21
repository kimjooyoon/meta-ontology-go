package bidir

// BXTypedEnvelopeSchemaVersion identifies the parser-neutral typed boundary.
const BXTypedEnvelopeSchemaVersion = "bidir-typed-envelope/v1"

// BXTypedProjectionEnvelope binds a source document to typed semantic views.
// Digest seals the envelope against stale or relabeled observations.
type BXTypedProjectionEnvelope struct {
	SchemaVersion string
	Fixture       string
	Document      Document
	Base          Model
	Left          Model
	Right         Model
	AcceptedDelta FactDelta
	PartialDelta  FactDelta
	BaseEvidence  BXBaseEvidenceInput
	Digest        string
}

// BXObserverReceipt is an observer-owned rejected-write receipt.
type BXObserverReceipt struct {
	ObserverKind string
	Observation  BXWriteObservation
	Digest       string
}

// BXTypedObservationEnvelope carries accepted evidence and a required receipt.
type BXTypedObservationEnvelope struct {
	SchemaVersion string
	Accepted      BXWriteObservation
	Rejected      *BXObserverReceipt
}

// BXTypedEnvelopeResult is bounded contract evidence, never promotion proof.
type BXTypedEnvelopeResult struct {
	SchemaVersion      string
	Fixture            string
	FeatureGreen       bool
	ProjectionDigest   string
	BaseFingerprint    string
	LeftFingerprint    string
	RightFingerprint   string
	GetPutPassed       bool
	PutGetPassed       bool
	SemanticEquivalent bool
	ThreeWay           ThreeWayResult
	Evidence           BXEvidence
	Delta              BXDeltaEvidence
	PartialDelta       BXDeltaEvidence
	Locality           Locality
	Candidates         []string
	PortSequence       []string
	RelationSequence   []string
	ObserverKind       string
	NoWriteObserved    bool
	Deferred           []string
	CanonicalJSON      string
	Hash               string
}
