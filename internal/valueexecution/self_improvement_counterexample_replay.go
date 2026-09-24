package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementCounterexampleReplayObservationSchema = "gooo.self-improvement.counterexample-replay.v1"

type SelfImprovementCounterexampleReplayStatus string

const SelfImprovementCounterexampleReplayStatusMatched SelfImprovementCounterexampleReplayStatus = "MATCHED"
const SelfImprovementCounterexampleReplayStatusMismatch SelfImprovementCounterexampleReplayStatus = "MISMATCH"
const SelfImprovementCounterexampleReplayStatusUnknown SelfImprovementCounterexampleReplayStatus = "UNKNOWN"

type SelfImprovementCounterexampleReplayObservation struct {
	Schema         string                                    `json:"schema"`
	ExpectedDigest string                                    `json:"expected_digest"`
	ReplayDigest   string                                    `json:"replay_digest"`
	Status         SelfImprovementCounterexampleReplayStatus `json:"status"`
	Reason         string                                    `json:"reason"`
	NonAuthorizing bool                                      `json:"non_authorizing"`
	Digest         string                                    `json:"digest"`
}

// ObserveSelfImprovementCounterexampleReplay verifies that a retained
// counterexample can be reused without silently changing its origin.
func ObserveSelfImprovementCounterexampleReplay(
	expected SelfImprovementCounterexampleObservation,
	replay SelfImprovementCounterexampleObservation,
) SelfImprovementCounterexampleReplayObservation {
	observation := SelfImprovementCounterexampleReplayObservation{
		Schema:         SelfImprovementCounterexampleReplayObservationSchema,
		ExpectedDigest: expected.Digest,
		ReplayDigest:   replay.Digest,
		Status:         SelfImprovementCounterexampleReplayStatusUnknown,
		Reason:         "COUNTEREXAMPLE_REPLAY_UNKNOWN",
		NonAuthorizing: true,
	}
	switch {
	case expected.Status != SelfImprovementCounterexampleStatusRetained || replay.Status != SelfImprovementCounterexampleStatusRetained:
		observation.Reason = "COUNTEREXAMPLE_NOT_RETAINED"
	case expected.SourceURI != replay.SourceURI || expected.InputDigest != replay.InputDigest:
		observation.Status = SelfImprovementCounterexampleReplayStatusMismatch
		observation.Reason = "COUNTEREXAMPLE_ORIGIN_MISMATCH"
	case expected.Digest == "" || replay.Digest == "":
		observation.Reason = "COUNTEREXAMPLE_DIGEST_MISSING"
	case expected.Digest != replay.Digest:
		observation.Status = SelfImprovementCounterexampleReplayStatusMismatch
		observation.Reason = "COUNTEREXAMPLE_EVIDENCE_MISMATCH"
	default:
		observation.Status = SelfImprovementCounterexampleReplayStatusMatched
		observation.Reason = "COUNTEREXAMPLE_REPLAY_MATCHED"
	}
	observation.Digest = selfImprovementCounterexampleReplayObservationDigest(observation)
	return observation
}

func selfImprovementCounterexampleReplayObservationDigest(observation SelfImprovementCounterexampleReplayObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
