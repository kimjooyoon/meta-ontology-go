package candidateleakage

import "reflect"

func baseInput(subjectSHA string) Input {
	candidateDigest := digestText("candidate:" + subjectSHA)
	return Input{
		Schema: InputSchema, SubjectSHA: subjectSHA,
		Candidate: Candidate{SubjectSHA: subjectSHA, Digest: candidateDigest,
			Decision: CandidateAllowLimited, Resolution: ResolutionExact, MetaOperation: BoundaryOperation},
		Promotion: Promotion{SubjectSHA: subjectSHA, CandidateDigest: candidateDigest,
			EvidenceDigest: digestText("promotion:" + subjectSHA), Decision: PromotionDenied,
			Resolution: ResolutionExact, MetaOperation: BoundaryOperation},
		Official: Official{SubjectSHA: subjectSHA, Status: OfficialNotImplemented,
			Decision: OfficialBlock, Resolution: OfficialResolutionNone, MetaOperation: BoundaryOperation},
	}
}

func makeOfficialPositive(input *Input, decision string) {
	input.Official.Status = OfficialOperating
	input.Official.Decision = decision
	input.Official.Resolution = ResolutionExact
}

func matches(definition Definition, report Report) bool {
	return report.Decision == definition.ExpectedDecision &&
		report.Resolution == definition.ExpectedResolution &&
		report.EnforcementEffect == definition.ExpectedEffect && report.Reason == definition.ExpectedReason &&
		report.Summary.LeakagePaths == definition.ExpectedLeakage &&
		report.Summary.UnknownPaths == definition.ExpectedUnknown && report.Summary.RepositoryWrites == 0 &&
		report.Summary.PromotionCreditBPS == 0
}

func replayEqual(left, right Report) bool {
	return reflect.DeepEqual(left, right)
}

func countCase(summary *SuiteSummary, report Report, passed bool) {
	if passed {
		summary.CasesPassed++
	}
	if report.Decision == DecisionPass && report.Resolution == ResolutionExact {
		summary.ExactPass++
	} else if report.Decision == DecisionFailClosed && report.Resolution == ResolutionExact {
		summary.ExactFailClosed++
	} else if report.Decision == DecisionFailClosed && report.Resolution == ResolutionInvariant {
		summary.InvariantFailClosed++
	}
}
