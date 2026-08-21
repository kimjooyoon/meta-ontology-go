package bidir

import (
	"errors"
	"fmt"
	"strings"
)

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
