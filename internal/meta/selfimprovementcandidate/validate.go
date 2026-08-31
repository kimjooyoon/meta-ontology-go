package selfimprovementcandidate

import (
	"fmt"
	"slices"
)

func Validate(report Report, expectedHead string, sourceRunID int64) error {
	if report.Schema != ReportSchema || report.Metaprogram != "internal/meta/selfimprovementcandidate" ||
		report.SubjectSHA != expectedHead || report.SourceWorkflowRunID != sourceRunID ||
		!validSHA(expectedHead) || !validDigest(report.SourceObservationDigest) ||
		!validDigest(report.SourceFileDigest) || report.PolicyVersion != PolicyVersion ||
		report.Digest != reportDigest(report) || !authorityClosed(report.Authority) {
		return fmt.Errorf("candidate report identity mismatch")
	}
	if len(report.Indicators) != indicatorTotal || len(report.Views) != 3 || len(report.Proofs) != 3 ||
		!slices.Equal(report.NotClaimed, reportNonClaims) {
		return fmt.Errorf("candidate report shape mismatch")
	}
	switch report.Decision {
	case DecisionProposed:
		if !validProposed(report) {
			return fmt.Errorf("proposed candidate mismatch")
		}
	case DecisionFailClosed:
		if !validFailure(report) {
			return fmt.Errorf("failed candidate mismatch")
		}
	default:
		return fmt.Errorf("unknown candidate decision")
	}
	return nil
}

func authorityClosed(authority Authority) bool {
	return authority.RepositoryWrites == 0 && !authority.MutationAuthorized &&
		!authority.ExecutionAuthorized && !authority.PromotionAuthorized &&
		!authority.AutomaticAdoptionAuthorized
}

func validProposed(report Report) bool {
	return report.Resolution == ResolutionExact && report.Reason == ReasonProposed &&
		len(report.Candidates) == 1 && validCandidate(report.Candidates[0], report.SourceObservationDigest) &&
		coordinateEquals(report.Summary.Coordinates, 16, 16) &&
		coordinateEquals(report.Summary.SourceCoordinates, 16, 16) &&
		report.Summary.EligibleGaps == 1 && report.Summary.CandidateCount == 1 &&
		report.Summary.AchievedDelta == 0 && report.Summary.TargetDelta == 1 &&
		report.Summary.Unknowns == 0 && evidenceState(report, true)
}

func validFailure(report Report) bool {
	return (report.Resolution == ResolutionExact || report.Resolution == ResolutionLower) &&
		report.Reason != ReasonProposed && len(report.Candidates) == 0 &&
		coordinateEquals(report.Summary.Coordinates, 0, 16) &&
		coordinateEquals(report.Summary.SourceCoordinates, 0, 16) &&
		report.Summary.EligibleGaps == 0 && report.Summary.CandidateCount == 0 &&
		report.Summary.AchievedDelta == 0 && report.Summary.TargetDelta == 0 &&
		report.Summary.Unknowns == 0 && evidenceState(report, false)
}
