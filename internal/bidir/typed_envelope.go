package bidir

import (
	"errors"
	"fmt"
	"strings"
)

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

// SealBXTypedProjectionEnvelope creates a deterministic envelope digest.
func SealBXTypedProjectionEnvelope(envelope BXTypedProjectionEnvelope) (BXTypedProjectionEnvelope, error) {
	if envelope.SchemaVersion == "" {
		envelope.SchemaVersion = BXTypedEnvelopeSchemaVersion
	}
	if strings.TrimSpace(envelope.Fixture) == "" {
		return BXTypedProjectionEnvelope{}, errors.New("typed projection fixture is empty")
	}
	envelope.Digest = typedProjectionDigest(envelope)
	return envelope, nil
}

// NewBXTypedObservationEnvelope constructs the typed observation boundary.
func NewBXTypedObservationEnvelope(accepted BXWriteObservation, rejected *BXObserverReceipt) BXTypedObservationEnvelope {
	return BXTypedObservationEnvelope{SchemaVersion: BXTypedEnvelopeSchemaVersion, Accepted: accepted, Rejected: rejected}
}

// CaptureBXObserverReceipt captures a receipt without accepting producer snapshots as proof.
func CaptureBXObserverReceipt(observer BXRejectedWriteObserver, operation func() error) (*BXObserverReceipt, error) {
	if observer == nil {
		return nil, errors.New("typed observation observer is nil")
	}
	observation, err := observer.ObserveRejected(operation)
	if err != nil {
		return nil, fmt.Errorf("capture observer receipt: %w", err)
	}
	kind := strings.TrimSpace(observer.Kind())
	if kind == "" {
		return nil, errors.New("typed observation observer kind is empty")
	}
	receipt := &BXObserverReceipt{ObserverKind: kind, Observation: observation}
	receipt.Digest = observerReceiptDigest(*receipt)
	return receipt, nil
}

// AdaptBXTypedEnvelope joins typed laws and the existing BX evidence bundle.
func AdaptBXTypedEnvelope(projection BXTypedProjectionEnvelope, observation BXTypedObservationEnvelope) (BXTypedEnvelopeResult, error) {
	result := BXTypedEnvelopeResult{SchemaVersion: BXTypedEnvelopeSchemaVersion, FeatureGreen: false, Deferred: deferredBXSeams()}
	if err := projection.validate(); err != nil {
		return result, fmt.Errorf("typed projection envelope: %w", err)
	}
	if err := observation.validate(projection.Document); err != nil {
		return result, fmt.Errorf("typed observation envelope: %w", err)
	}
	base, err := Get(projection.Document)
	if err != nil {
		return result, fmt.Errorf("typed projection document: %w", err)
	}
	if !SemanticEquivalent(base, projection.Base) {
		return result, errors.New("typed projection base model is stale")
	}
	accepted, err := Reconcile(base, projection.AcceptedDelta)
	if err != nil {
		return result, fmt.Errorf("typed projection accepted delta: %w", err)
	}
	if err := CheckPutGet(projection.Document, accepted.Model); err != nil {
		return result, fmt.Errorf("typed projection Put-Get law: %w", err)
	}
	threeWay, err := ReconcileThreeWay(projection.Base, projection.Left, projection.Right)
	if err != nil {
		return result, fmt.Errorf("typed projection three-way law: %w", err)
	}
	if !SemanticEquivalent(accepted.Model, threeWay.Model) {
		return result, errors.New("typed projection three-way result diverges from accepted model")
	}
	fixture := bxTypedEnvelopeFixture{projection: projection, observation: observation}
	evidence, err := MeasureBXFixture(fixture)
	if err != nil {
		return result, fmt.Errorf("typed projection evidence: %w", err)
	}
	result = typedEnvelopeResult(projection, observation, threeWay, evidence)
	result.CanonicalJSON, err = typedEnvelopeJSON(result)
	if err != nil {
		return BXTypedEnvelopeResult{}, fmt.Errorf("typed projection canonical evidence: %w", err)
	}
	result.Hash = digest(result.CanonicalJSON)
	if err := result.Validate(); err != nil {
		return BXTypedEnvelopeResult{}, err
	}
	return result, nil
}

// Validate checks result self-consistency and keeps the feature gate closed.
func (result BXTypedEnvelopeResult) Validate() error {
	if result.SchemaVersion != BXTypedEnvelopeSchemaVersion || result.FeatureGreen {
		return errors.New("typed envelope result is not bounded contract evidence")
	}
	if result.CanonicalJSON == "" || result.Hash != digest(result.CanonicalJSON) {
		return errors.New("typed envelope result hash is missing or stale")
	}
	canonical, err := typedEnvelopeJSON(result)
	if err != nil || canonical != result.CanonicalJSON {
		return errors.New("typed envelope canonical evidence is stale")
	}
	if !result.GetPutPassed || !result.PutGetPassed || !result.SemanticEquivalent || !result.ThreeWay.Succeeded() {
		return errors.New("typed envelope semantic laws are not satisfied")
	}
	if result.Evidence.RejectedTransaction.Deferred {
		return errors.New("typed envelope rejected transaction was deferred")
	}
	if err := result.Evidence.validate(); err != nil {
		return fmt.Errorf("typed envelope evidence: %w", err)
	}
	if result.Evidence.Delta.CandidatePromoted || result.Evidence.PartialDelta.RemovedCreated || result.Evidence.PartialDelta.CandidatePromoted {
		return errors.New("typed envelope partial observation changed authority")
	}
	return nil
}

func (envelope BXTypedProjectionEnvelope) validate() error {
	if envelope.SchemaVersion != BXTypedEnvelopeSchemaVersion || envelope.Digest == "" {
		return errors.New("typed projection envelope is unsealed")
	}
	if envelope.Digest != typedProjectionDigest(envelope) {
		return errors.New("typed projection envelope digest mismatch")
	}
	return nil
}

func (envelope BXTypedObservationEnvelope) validate(document Document) error {
	if envelope.SchemaVersion != BXTypedEnvelopeSchemaVersion {
		return errors.New("unsupported typed observation schema")
	}
	if envelope.Rejected == nil {
		return errors.New("observer receipt is absent")
	}
	if err := envelope.Rejected.validate(document); err != nil {
		return err
	}
	return nil
}

func (receipt BXObserverReceipt) validate(document Document) error {
	if receipt.ObserverKind == "" || receipt.Digest == "" {
		return errors.New("observer receipt is incomplete")
	}
	if receipt.Digest != observerReceiptDigest(receipt) {
		return errors.New("observer receipt digest mismatch")
	}
	if !receipt.Observation.Observed || !sameSnapshot(receipt.Observation.Before, receipt.Observation.After) {
		return errors.New("observer receipt does not prove no-write")
	}
	if err := snapshotMatches(receipt.Observation.Before, document); err != nil {
		return fmt.Errorf("observer receipt before snapshot: %w", err)
	}
	return snapshotMatches(receipt.Observation.After, document)
}
