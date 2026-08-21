package bidir

import (
	"errors"
	"fmt"
)

// BXEvidence is the stable, reviewable output of one reconciliation fixture.
type BXEvidence struct {
	SchemaVersion        string
	Fixture              string
	Base                 BXBaseEvidence
	GetPutPassed         bool
	PutGetPassed         bool
	SemanticEquivalent   bool
	AcceptedRelationAdds int
	Delta                BXDeltaEvidence
	PartialDelta         BXDeltaEvidence
	AcceptedTransaction  BXTransactionEvidence
	RejectedTransaction  BXTransactionEvidence
	Locality             Locality
	PartialConflict      BXConflictEvidence
	Deferred             []string
}

func (e BXEvidence) validate() error {
	if e.SchemaVersion != BXEvidenceSchemaVersion {
		return fmt.Errorf("unsupported evidence schema %q", e.SchemaVersion)
	}
	if e.Fixture == "" {
		return errors.New("evidence fixture name is empty")
	}
	if err := validateBaseEvidence(e.Base); err != nil {
		return err
	}
	if !e.GetPutPassed || !e.PutGetPassed || !e.SemanticEquivalent {
		return errors.New("round-trip semantic evidence is not green")
	}
	if err := validateDeltaEvidence(e.Delta); err != nil {
		return fmt.Errorf("accepted delta evidence: %w", err)
	}
	if err := validateDeltaEvidence(e.PartialDelta); err != nil {
		return fmt.Errorf("partial delta evidence: %w", err)
	}
	if !e.PartialDelta.PartialObservation {
		return errors.New("partial delta is not marked as a partial observation")
	}
	if err := validateTransaction(e.AcceptedTransaction, false); err != nil {
		return fmt.Errorf("accepted transaction evidence: %w", err)
	}
	if err := validateTransaction(e.RejectedTransaction, true); err != nil {
		return fmt.Errorf("rejected transaction evidence: %w", err)
	}
	if e.Delta.CandidatePromoted || e.Delta.RemovedCreated || e.PartialDelta.CandidatePromoted || e.PartialDelta.RemovedCreated || e.PartialConflict.RemovedCreated || e.PartialConflict.CandidatePromoted {
		return errors.New("partial observation changed authoritative semantic state")
	}
	if e.RejectedTransaction.Deferred || !e.RejectedTransaction.NoWrite || !e.PartialConflict.NoWriteObserved {
		return errors.New("rejected transaction lacks observed atomic no-write evidence")
	}
	if e.PartialConflict.Kind == "" || e.PartialConflict.Count == 0 || !e.PartialConflict.Transactional {
		return errors.New("partial delta did not produce a transactional rejection")
	}
	for _, required := range deferredBXSeams() {
		if !containsString(e.Deferred, required) {
			return fmt.Errorf("deferred seam %q is missing", required)
		}
	}
	return nil
}
