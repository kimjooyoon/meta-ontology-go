package proofchoicealgebra

import (
	"bytes"
	"encoding/json"
)

func regressionEvidence(value Value, lowered lowered, baseline []byte) Evidence {
	result := Evidence{ClaimID: value.ID, Subject: value.Subject, Route: Regression}
	if len(baseline) == 0 {
		result.State, result.Reason = UnknownState, "REGRESSION_BASELINE_UNKNOWN"
		return finishEvidence(result)
	}
	prior, err := decodeArtifact(baseline)
	if err != nil {
		result.State, result.Reason = UnknownState, "REGRESSION_BASELINE_UNKNOWN"
		return finishEvidence(result)
	}
	current := makeArtifact(lowered)
	currentData, _ := json.Marshal(current)
	result.FirstArtifactDigest = digestBytes(baseline)
	result.SecondArtifactDigest = digestBytes(currentData)
	result.ByteEqual = bytes.Equal(baseline, currentData)
	result.SemanticEqual = prior.SemanticDigest == current.SemanticDigest
	result.ObservationSlots = regressionSlots(value, result, current.SemanticDigest)
	result.Provenance = provenanceOf(result.ObservationSlots)
	if result.ByteEqual && result.SemanticEqual {
		result.State = "OBSERVED"
	} else {
		result.State, result.Reason = InsufficientState, "REGRESSION_REPLAY_MISMATCH"
	}
	return finishEvidence(result)
}

func regressionSlots(value Value, result Evidence, digest string) []ObservationSlot {
	return []ObservationSlot{
		{ID: value.ID + ":baseline-artifact", Observed: true, Provenance: []string{"artifact://baseline"}},
		{ID: value.ID + ":replay-artifact", Observed: true, Provenance: []string{"artifact://replay/" + digest}},
		{ID: value.ID + ":replay-equality", Observed: result.ByteEqual && result.SemanticEqual, Provenance: []string{"replay://comparison"}},
	}
}
