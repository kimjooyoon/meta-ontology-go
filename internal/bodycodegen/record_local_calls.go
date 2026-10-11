package bodycodegen

import (
	"context"
	"strconv"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func evaluateRecordCasesObserved(ctx context.Context, candidate generatedRoute, mask uint16,
	cases []assemblyspec.ValueCase) ([]RecordAssemblyCase, *LocalCallObservation, error) {
	r := newLocalCallRecorder(candidate.source, candidate.report, cases)
	r.observation.Phase, r.observation.CandidateID = "record-candidate", strconv.Itoa(int(mask))
	rows, err := evaluateRecordAssemblyRecorded(ctx, candidate.source, candidate.report.Activity,
		candidate.report.RecordTypes, cases, r)
	return rows, r.finish(ctx, err), err
}

func equalRecordAttempts(prior, actual []RecordAssemblyAttempt) bool {
	if len(prior) != len(actual) || (prior == nil) != (actual == nil) {
		return false
	}
	for i, observed := range actual {
		if prior[i].LocalCalls == nil {
			observed.LocalCalls = nil
		}
		if !equalLocalCallEvidence(prior[i], observed) {
			return false
		}
	}
	return true
}

// EqualRecordCandidateHistory compares original local obligations and current
// observed calls while preserving legacy missing history as unobserved.
func EqualRecordCandidateHistory(prior, actual []RecordCandidate) bool {
	if len(prior) != len(actual) || (prior == nil) != (actual == nil) {
		return false
	}
	for i, observed := range actual {
		if prior[i].Attempt.LocalCalls == nil {
			observed.Attempt.LocalCalls = nil
		}
		if !equalLocalCallEvidence(prior[i], observed) {
			return false
		}
	}
	return true
}
