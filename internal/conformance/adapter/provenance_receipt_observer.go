package adapter

import (
	"encoding/json"
	"fmt"
)

// ReceiptStateDigests are computed from an independent observer capture.
type ReceiptStateDigests struct {
	Before string
	After  string
}

// StateDigests verifies the observer seal before exposing state digests.
func (o *NoWriteObservation) StateDigests(request Request) (ReceiptStateDigests, error) {
	if err := o.VerifyNoWrite(request); err != nil {
		return ReceiptStateDigests{}, err
	}
	return observerStateDigests(*o)
}

// ValidateObservedNoWrite binds a receipt to observer-owned rejected evidence.
// Receipt fields supplied by a producer are insufficient without this path.
func (r ProvenanceReceipt) ValidateObservedNoWrite(request Request, observation *NoWriteObservation) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if observation == nil {
		return oracleError(OracleNW001, "observer evidence is required for receipt validation")
	}
	if r.Binding != requestObservationBinding(request) || r.Binding != observation.Binding {
		return oracleError(OracleID001, "receipt and observer bindings do not match request")
	}
	digests, err := observation.StateDigests(request)
	if err != nil {
		return err
	}
	if err := validateObservedWorkflow(r, observation.Workflow); err != nil {
		return err
	}
	if err := validateReceiptReason(r.Outcome, observation.Reason); err != nil {
		return err
	}
	if r.BeforeStateDigest != digests.Before || r.AfterStateDigest != digests.After {
		return oracleError(OracleNW003, "receipt state digests do not match observer capture")
	}
	return nil
}

func observerStateDigests(observation NoWriteObservation) (ReceiptStateDigests, error) {
	before, err := digestFilesystemState(observation.Before)
	if err != nil {
		return ReceiptStateDigests{}, fmt.Errorf("digest before observer state: %w", err)
	}
	after, err := digestFilesystemState(observation.After)
	if err != nil {
		return ReceiptStateDigests{}, fmt.Errorf("digest after observer state: %w", err)
	}
	return ReceiptStateDigests{Before: before, After: after}, nil
}

func digestFilesystemState(state FilesystemState) (string, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return digestBytes(payload), nil
}

func validateReceiptReason(outcome ReceiptOutcome, reason RejectionKind) error {
	switch outcome {
	case ReceiptOutcomeRejected:
		if reason != "" {
			return oracleError(OracleNW003, "rejected receipt has a cancellation reason")
		}
	case ReceiptOutcomeCancelled:
		if reason != RejectionCancelled {
			return oracleError(OracleNW003, "cancelled receipt is not bound to cancellation evidence")
		}
	case ReceiptOutcomeClosed:
		if reason != RejectionClosed {
			return oracleError(OracleNW003, "closed receipt is not bound to close evidence")
		}
	}
	return nil
}
