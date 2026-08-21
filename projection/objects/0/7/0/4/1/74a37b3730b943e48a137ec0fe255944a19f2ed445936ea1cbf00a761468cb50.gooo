package bidir

import (
	"errors"
	"fmt"
)

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
