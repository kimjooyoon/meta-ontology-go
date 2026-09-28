package main

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"

// Diagnostics explain the selected failure; they do not select or authorize an action.
func observationFailureFromError(action generation.Action, runErr *operationError, process generation.ProcessObservation) generation.ObservationFailure {
	failure := observationFailure(action, runErr.stage, runErr.step, runErr.reason, runErr.class, runErr.next, runErr.blockedBy, process)
	failure.FailureEvidence = append([]generation.ObservationFailureEvidence{}, runErr.evidence...)
	failure.Counterexample = runErr.counterexample
	failure.DerivedRelations = append([]generation.CounterexampleRelation{}, runErr.derivedRelations...)
	failure.Diagnostics = append([]string(nil), runErr.diagnostics...)
	return failure
}
