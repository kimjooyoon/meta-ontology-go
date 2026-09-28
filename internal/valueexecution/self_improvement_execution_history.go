package valueexecution

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const SelfImprovementExecutionHistorySchema = "gooo.self-improvement.execution-history.v1"

type SelfImprovementExecutionHistory struct {
	Schema         string                                       `json:"schema"`
	Outcomes       []SelfImprovementExecutionOutcomeObservation `json:"outcomes"`
	NonAuthorizing bool                                         `json:"non_authorizing"`
	Digest         string                                       `json:"digest"`
}

type SelfImprovementExecutionHistoryDecision string

const SelfImprovementExecutionHistoryDecisionUnknown SelfImprovementExecutionHistoryDecision = "UNKNOWN"
const SelfImprovementExecutionHistoryDecisionInsufficientEvidence SelfImprovementExecutionHistoryDecision = "INSUFFICIENT_EVIDENCE"
const SelfImprovementExecutionHistoryDecisionBlockedRejected SelfImprovementExecutionHistoryDecision = "BLOCKED_REJECTED"
const SelfImprovementExecutionHistoryDecisionCompletedEvidence SelfImprovementExecutionHistoryDecision = "COMPLETED_EVIDENCE"

type SelfImprovementExecutionHistoryObservation struct {
	Schema         string                                  `json:"schema"`
	Observed       int                                     `json:"observed"`
	Completed      int                                     `json:"completed"`
	Rejected       int                                     `json:"rejected"`
	Unknown        int                                     `json:"unknown"`
	Decision       SelfImprovementExecutionHistoryDecision `json:"decision"`
	Reason         string                                  `json:"reason"`
	NonAuthorizing bool                                    `json:"non_authorizing"`
	Digest         string                                  `json:"digest"`
}

func NewSelfImprovementExecutionHistory() SelfImprovementExecutionHistory {
	history := SelfImprovementExecutionHistory{
		Schema:         SelfImprovementExecutionHistorySchema,
		Outcomes:       []SelfImprovementExecutionOutcomeObservation{},
		NonAuthorizing: true,
	}
	history.Digest = selfImprovementExecutionHistoryDigest(history)
	return history
}

// Append returns a new history and preserves every prior execution outcome.
func (history SelfImprovementExecutionHistory) Append(outcome SelfImprovementExecutionOutcomeObservation) SelfImprovementExecutionHistory {
	if history.Schema == "" {
		history.Schema = SelfImprovementExecutionHistorySchema
	}
	outcomes := append([]SelfImprovementExecutionOutcomeObservation(nil), history.Outcomes...)
	history.Outcomes = append(outcomes, outcome)
	history.NonAuthorizing = true
	history.Digest = selfImprovementExecutionHistoryDigest(history)
	return history
}

// Observe summarizes history without erasing unknown or rejected outcomes.
func (history SelfImprovementExecutionHistory) Observe() SelfImprovementExecutionHistoryObservation {
	observation := SelfImprovementExecutionHistoryObservation{
		Schema:         SelfImprovementExecutionHistorySchema,
		Observed:       len(history.Outcomes),
		Decision:       SelfImprovementExecutionHistoryDecisionUnknown,
		NonAuthorizing: true,
	}
	if history.Schema != SelfImprovementExecutionHistorySchema || !history.NonAuthorizing || history.Digest == "" {
		observation.Reason = "EXECUTION_HISTORY_IDENTITY_UNKNOWN"
		observation.Digest = selfImprovementExecutionHistoryObservationDigest(observation)
		return observation
	}
	for _, outcome := range history.Outcomes {
		switch {
		case !outcome.NonAuthorizing || outcome.Digest == "" || outcome.Status == SelfImprovementExecutionOutcomeStatusUnknown:
			observation.Unknown++
		case outcome.Status == SelfImprovementExecutionOutcomeStatusCompleted:
			observation.Completed++
		case outcome.Status == SelfImprovementExecutionOutcomeStatusRejected:
			observation.Rejected++
		default:
			observation.Unknown++
		}
	}
	switch {
	case observation.Unknown > 0:
		observation.Reason = "EXECUTION_HISTORY_CONTAINS_UNKNOWN"
	case observation.Rejected > 0:
		observation.Decision = SelfImprovementExecutionHistoryDecisionBlockedRejected
		observation.Reason = "EXECUTION_HISTORY_CONTAINS_REJECTED"
	case observation.Completed == 0:
		observation.Decision = SelfImprovementExecutionHistoryDecisionInsufficientEvidence
		observation.Reason = "EXECUTION_HISTORY_REQUIRES_COMPLETED_OUTCOME"
	default:
		observation.Decision = SelfImprovementExecutionHistoryDecisionCompletedEvidence
		observation.Reason = "EXECUTION_HISTORY_COMPLETED_EVIDENCE"
	}
	observation.Digest = selfImprovementExecutionHistoryObservationDigest(observation)
	return observation
}

func selfImprovementExecutionHistoryDigest(history SelfImprovementExecutionHistory) string {
	history.Digest = ""
	encoded, _ := json.Marshal(history)
	return cache.HashBytes(encoded).String()
}

func selfImprovementExecutionHistoryObservationDigest(observation SelfImprovementExecutionHistoryObservation) string {
	observation.Digest = ""
	encoded, _ := json.Marshal(observation)
	return cache.HashBytes(encoded).String()
}
