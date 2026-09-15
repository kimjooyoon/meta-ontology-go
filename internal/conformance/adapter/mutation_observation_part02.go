package adapter

import (
	"fmt"
	"strings"
)

func validateMutationAttempt(attempt MutationAttempt, paths ObserverPaths) error {
	if strings.TrimSpace(attempt.Operation) == "" || strings.TrimSpace(attempt.Outcome) == "" {
		return fmt.Errorf("mutation attempt operation and outcome are required")
	}
	canonical, err := canonicalObserverPath(attempt.Path)
	if err != nil || canonical != attempt.Path {
		return fmt.Errorf("mutation attempt path is not canonical")
	}
	if canonical != paths.SourcePath && canonical != paths.OutputPath &&
		!observerPathContains(paths.TempRoot, canonical) {
		return fmt.Errorf("mutation attempt path is outside observer paths")
	}
	return nil
}
func mutationAttemptLess(left, right MutationAttempt) bool {
	if left.Path != right.Path {
		return left.Path < right.Path
	}
	if left.Operation != right.Operation {
		return left.Operation < right.Operation
	}
	return left.Outcome < right.Outcome
}

// CaptureUnverifiedMutation stores advisory data that can never satisfy the oracle.
func (o *NoWriteObserver) CaptureUnverifiedMutation(evidence MutationEvidence) error {
	if evidence.Status == MutationEvidenceVerified {
		return oracleError(OracleNW003, "public mutation capture cannot verify evidence")
	}
	if evidence.Status != MutationEvidenceMissing && evidence.Status != MutationEvidenceUnverified {
		return oracleError(OracleNW003, "unsupported mutation evidence status")
	}
	return o.captureMutation(evidence)
}
func (o *NoWriteObserver) captureVerifiedMutation(evidence verifiedMutationEvidence) error {
	return o.captureMutation(evidence.evidence)
}
func (o *NoWriteObserver) captureMutation(evidence MutationEvidence) error {
	if o == nil || o.stamp == nil || o.finished {
		return oracleError(OracleNW003, "observer mutation capture is closed")
	}
	if o.mutationCaptured {
		return oracleError(OracleNW003, "observer mutation capture is immutable")
	}
	o.mutation = evidence.clone()
	o.mutationCaptured = true
	return nil
}
