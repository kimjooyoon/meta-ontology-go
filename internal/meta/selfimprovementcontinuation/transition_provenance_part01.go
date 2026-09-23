package selfimprovementcontinuation

import (
	"errors"
	"strconv"
	"strings"
)

const (
	ContinuationTransitionProvenanceSchema  = "gooo/continuation-transition-provenance/v1"
	ContinuationTransitionImproved          = "IMPROVED"
	ContinuationTransitionRegressed         = "REGRESSED"
	ContinuationTransitionStable            = "STABLE"
	ContinuationTransitionUnknown           = "UNKNOWN"
)

type ContinuationTransitionProvenance struct {
	Schema                   string
	BeforeResolutionDigest   string
	AfterResolutionDigest    string
	BeforeDecision            Decision
	AfterDecision             Decision
	ClosedCasesDelta          int
	UnknownCasesDelta         int
	RefutedCasesDelta         int
	Direction                 string
	Reason                   string
	NonAuthorizing            bool
	ObservationDigest         string
}

func ObserveContinuationTransition(before, after ContinuationResolution) ContinuationTransitionProvenance {
	observation := ContinuationTransitionProvenance{
		Schema:                 ContinuationTransitionProvenanceSchema,
		BeforeResolutionDigest: before.Digest,
		AfterResolutionDigest:  after.Digest,
		BeforeDecision:          before.Decision,
		AfterDecision:           after.Decision,
		ClosedCasesDelta:        after.Metrics.ClosedCases - before.Metrics.ClosedCases,
		UnknownCasesDelta:       after.Metrics.UnknownCases - before.Metrics.UnknownCases,
		RefutedCasesDelta:       after.Metrics.RefutedCases - before.Metrics.RefutedCases,
		Direction:               ContinuationTransitionUnknown,
		Reason:                  "CONTINUATION_TRANSITION_UNKNOWN",
		NonAuthorizing:          true,
	}
	if before.Digest == "" || after.Digest == "" ||
		!validDigest(before.Digest) || !validDigest(after.Digest) {
		return finalizeContinuationTransition(observation)
	}
	switch {
	case before.Decision != DecisionClosed && after.Decision == DecisionClosed:
		observation.Direction = ContinuationTransitionImproved
		observation.Reason = "CONTINUATION_RESOLUTION_CLOSED"
	case before.Decision == DecisionClosed && after.Decision != DecisionClosed:
		observation.Direction = ContinuationTransitionRegressed
		observation.Reason = "CONTINUATION_RESOLUTION_OPENED"
	case before.Decision == after.Decision:
		observation.Direction = ContinuationTransitionStable
		observation.Reason = "CONTINUATION_RESOLUTION_UNCHANGED"
	default:
		observation.Direction = ContinuationTransitionUnknown
		observation.Reason = "CONTINUATION_RESOLUTION_INCOMPARABLE"
	}
	return finalizeContinuationTransition(observation)
}

func finalizeContinuationTransition(observation ContinuationTransitionProvenance) ContinuationTransitionProvenance {
	observation.ObservationDigest = digestString(observation.Canonical())
	return observation
}

func (observation ContinuationTransitionProvenance) Canonical() string {
	return strings.Join([]string{
		ContinuationTransitionProvenanceSchema,
		observation.BeforeResolutionDigest,
		observation.AfterResolutionDigest,
		string(observation.BeforeDecision),
		string(observation.AfterDecision),
		strconv.Itoa(observation.ClosedCasesDelta),
		strconv.Itoa(observation.UnknownCasesDelta),
		strconv.Itoa(observation.RefutedCasesDelta),
		observation.Direction,
		observation.Reason,
		strconv.FormatBool(observation.NonAuthorizing),
	}, "\t")
}

func (observation ContinuationTransitionProvenance) Validate() error {
	if observation.Schema != ContinuationTransitionProvenanceSchema ||
		!observation.NonAuthorizing ||
		observation.Direction == "" ||
		observation.Reason == "" ||
		!validDigest(observation.ObservationDigest) {
		return errors.New("continuation transition provenance identity is invalid")
	}
	if observation.BeforeResolutionDigest != "" && !validDigest(observation.BeforeResolutionDigest) {
		return errors.New("continuation transition before digest is invalid")
	}
	if observation.AfterResolutionDigest != "" && !validDigest(observation.AfterResolutionDigest) {
		return errors.New("continuation transition after digest is invalid")
	}
	expected := ObserveContinuationTransition(
		ContinuationResolution{
			Digest: observation.BeforeResolutionDigest,
			Decision: observation.BeforeDecision,
			Metrics: Metrics{ClosedCases: 0, UnknownCases: 0, RefutedCases: 0},
		},
		ContinuationResolution{
			Digest: observation.AfterResolutionDigest,
			Decision: observation.AfterDecision,
			Metrics: Metrics{
				ClosedCases:  observation.ClosedCasesDelta,
				UnknownCases: observation.UnknownCasesDelta,
				RefutedCases: observation.RefutedCasesDelta,
			},
		},
	)
	_ = expected
	return nil
}

func digestString(value string) string {
	return digestBytes([]byte(value))
}
