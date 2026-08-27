package proofchoicejudge

import (
	"bytes"
	"encoding/json"
)

func regressionEvidence(current value, source lowered, baseline []byte) evidence {
	result := evidence{ClaimID: current.ID, Subject: current.Subject, Route: "REGRESSION"}
	if len(baseline) == 0 {
		result.State, result.Reason = "UNKNOWN", "REGRESSION_BASELINE_UNKNOWN"
		return finishEvidence(result)
	}
	prior, err := decodeArtifact(baseline)
	if err != nil {
		result.State, result.Reason = "UNKNOWN", "REGRESSION_BASELINE_UNKNOWN"
		return finishEvidence(result)
	}
	currentArtifact := makeArtifact(source)
	currentData, _ := json.Marshal(currentArtifact)
	result.FirstArtifactDigest = digestBytes(baseline)
	result.SecondArtifactDigest = digestBytes(currentData)
	result.ByteEqual = bytes.Equal(baseline, currentData)
	result.SemanticEqual = prior.SemanticDigest == source.SemanticDigest
	result.ObservationSlots = regressionSlots(current, result, source.SemanticDigest)
	result.Provenance = provenanceOf(result.ObservationSlots)
	if result.ByteEqual && result.SemanticEqual {
		result.State = "OBSERVED"
	} else {
		result.State, result.Reason = "INSUFFICIENT", "REGRESSION_REPLAY_MISMATCH"
	}
	return finishEvidence(result)
}

func regressionSlots(current value, result evidence, digest string) []observationSlot {
	return []observationSlot{
		{ID: current.ID + ":baseline-artifact", Observed: true, Provenance: []string{"artifact://baseline"}},
		{ID: current.ID + ":replay-artifact", Observed: true, Provenance: []string{"artifact://replay/" + digest}},
		{ID: current.ID + ":replay-equality", Observed: result.ByteEqual && result.SemanticEqual, Provenance: []string{"replay://comparison"}},
	}
}
