package bidir

import (
	"errors"
	"fmt"
)

func acceptedTransaction(contract BXEvidenceFixture, before, after Document, base Model, result ReconcileResult) (BXTransactionEvidence, error) {
	observation := contract.ObserveAcceptedWrite(before, after)
	if err := observationMatches(observation, before, after); err != nil {
		return BXTransactionEvidence{}, fmt.Errorf("accepted write observation: %w", err)
	}
	return BXTransactionEvidence{
		Before:       stateEvidence(base, before, result.Locality, observation.Before),
		After:        stateEvidence(result.Model, after, result.Locality, observation.After),
		ObserverKind: "accepted-fixture",
		Observed:     observation.Observed,
		Atomic:       true,
	}, nil
}
func partialEvidence(document Document, base Model, delta FactDelta, observer BXRejectedWriteObserver) (BXConflictEvidence, BXTransactionEvidence, BXDeltaEvidence, error) {
	if observer == nil {
		return BXConflictEvidence{}, BXTransactionEvidence{}, BXDeltaEvidence{}, errors.New("rejected write observer is nil")
	}
	result := ReconcileResult{Model: base}
	var reconcileErr error
	called := false
	observation, observerErr := observer.ObserveRejected(func() error {
		called = true
		result, reconcileErr = Reconcile(base, delta)
		return reconcileErr
	})
	if observerErr != nil {
		return BXConflictEvidence{}, BXTransactionEvidence{}, BXDeltaEvidence{}, fmt.Errorf("observe rejected write: %w", observerErr)
	}
	if !called {
		return BXConflictEvidence{}, BXTransactionEvidence{}, BXDeltaEvidence{}, errors.New("rejected write observer did not run operation")
	}
	partial := makeDeltaEvidenceUnchecked(delta, LocalityBetween(base, result.Model), true, base, result.Model)
	before := stateEvidence(base, document, LocalityBetween(base, result.Model), observation.Before)
	after := stateEvidence(result.Model, document, LocalityBetween(base, result.Model), observation.After)
	transaction := BXTransactionEvidence{
		Before: before, After: after, ObserverKind: observer.Kind(), Observed: observation.Observed,
		Atomic: before == after, NoWrite: before == after,
	}
	evidence := BXConflictEvidence{Transactional: SemanticEquivalent(base, result.Model)}
	evidence.RemovedCreated = removedCreated(base, result.Model, delta)
	evidence.CandidatePromoted = candidatePromoted(base, delta, result.Model)
	var conflictErr *ReconcileError
	if !errors.As(reconcileErr, &conflictErr) {
		evidence.NoWriteObserved = transaction.NoWrite
		return evidence, transaction, partial, nil
	}
	evidence.Count = len(conflictErr.Conflicts)
	if evidence.Count > 0 {
		evidence.Kind = conflictErr.Conflicts[0].Kind
	}
	partial.RemovedCreated = evidence.RemovedCreated
	partial.CandidatePromoted = evidence.CandidatePromoted
	evidence.NoWriteObserved = transaction.NoWrite
	return evidence, transaction, partial, nil
}
