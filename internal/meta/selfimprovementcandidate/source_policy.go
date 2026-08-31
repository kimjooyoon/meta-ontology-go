package selfimprovementcandidate

import "slices"

var expectedSourceNonClaims = []string{
	"business correctness",
	"value-level computation",
	"production readiness",
	"performance beyond this runner and fixed sample set",
	"general-purpose code generation",
	"improvement candidate quality",
	"automatic execution or adoption",
}

func classifySource(source sourceObservation, head string, runID int64) (string, string) {
	switch {
	case source.Decision != "OBSERVED" && source.Decision != "FAIL_CLOSED":
		return ReasonSourceUnknown, ResolutionLower
	case source.Resolution == ResolutionLower:
		return ReasonSourceLowered, ResolutionLower
	case source.Resolution != ResolutionExact:
		return ReasonSourceUnknown, ResolutionLower
	case source.Decision == "FAIL_CLOSED":
		return ReasonSourceRejected, ResolutionExact
	case !validSHA(head) || source.SubjectSHA != head || source.SourceWorkflowRunID != runID:
		return ReasonSourceIdentity, ResolutionExact
	case source.Digest != sourceDigest(source):
		return ReasonSourceIntegrity, ResolutionExact
	case source.Summary.CandidateCount != 0:
		return ReasonSourceCandidate, ResolutionExact
	case !sourceAuthorityClosed(source.Authority):
		return ReasonSourceAuthority, ResolutionExact
	case !slices.Equal(source.NotClaimed, expectedSourceNonClaims):
		return ReasonGapAbsent, ResolutionExact
	case !sourceShapeKnown(source):
		return ReasonSourceShape, ResolutionExact
	default:
		return "", ""
	}
}

func sourceAuthorityClosed(authority sourceAuthority) bool {
	return authority.RepositoryWrites == 0 && !authority.MutationAuthorized &&
		!authority.ExecutionAuthorized && !authority.PromotionAuthorized &&
		!authority.AutomaticAdoptionAuthorized
}
