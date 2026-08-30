package feedbackpredecessor

import "fmt"

func validateFoundation(input Input) error {
	foundation := input.Foundation
	if foundation == nil {
		return fmt.Errorf("foundation evidence is missing")
	}
	expected := FoundationEvidenceForConfirmedGap()
	if input.PredecessorSHA != FoundationMissingPredecessorSHA ||
		foundation.MissingPredecessorSHA != input.PredecessorSHA {
		return fmt.Errorf("foundation missing predecessor is not the confirmed gap")
	}
	if foundation.ProofChoice != expected.ProofChoice || foundation.Reason != expected.Reason ||
		foundation.LastKnownGoodSHA != expected.LastKnownGoodSHA ||
		foundation.LastKnownGoodRunID != expected.LastKnownGoodRunID ||
		foundation.LastKnownGoodArtifactID != expected.LastKnownGoodArtifactID ||
		foundation.LastKnownGoodArtifactName != expected.LastKnownGoodArtifactName ||
		foundation.LastKnownGoodArtifactDigest != expected.LastKnownGoodArtifactDigest ||
		foundation.LastKnownGoodReceiptDigest != expected.LastKnownGoodReceiptDigest ||
		!foundation.LastKnownGoodIsAncestor || foundation.NextOperation != expected.NextOperation ||
		foundation.UseCount != expected.UseCount || len(input.Candidates) != 0 {
		return fmt.Errorf("foundation evidence contradicts the confirmed gap")
	}
	if !sameStrings(foundation.GapCommitSHAs, expected.GapCommitSHAs) ||
		!sameInts(foundation.GapPRNumbers, expected.GapPRNumbers) ||
		!sameStrings(foundation.BlockedBy, expected.BlockedBy) {
		return fmt.Errorf("foundation evidence gap identity is not exact")
	}
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
