package selfimprovementcandidate

import "io/fs"

func Evaluate(repository fs.FS, contractPath, head string, sourceRunID int64, raw []byte) Report {
	report := Report{Schema: ReportSchema, Metaprogram: "internal/meta/selfimprovementcandidate",
		SubjectSHA: head, SourceWorkflowRunID: sourceRunID, SourceObservationDigest: zeroDigest,
		SourceFileDigest: digestBytes(raw), PolicyVersion: PolicyVersion, Authority: Authority{}}
	source, err := decodeSource(raw)
	if err != nil {
		return fail(report, ReasonSourceUnknown, ResolutionLower)
	}
	report.SourceObservationDigest = source.Digest
	if reason, resolution := classifySource(source, head, sourceRunID); reason != "" {
		return fail(report, reason, resolution)
	}
	contract, reason := compileContract(repository, contractPath)
	if reason != "" {
		resolution := ResolutionExact
		if reason == ReasonContractUnknown {
			resolution = ResolutionLower
		}
		return fail(report, reason, resolution)
	}
	report.Contract = contract
	report.Decision, report.Resolution, report.Reason = DecisionProposed, ResolutionExact, ReasonProposed
	report.Candidates = []Candidate{buildCandidate(source.Digest, gapPolicies[0])}
	return finish(report, true)
}

func fail(report Report, reason, resolution string) Report {
	report.Decision, report.Reason, report.Resolution = DecisionFailClosed, reason, resolution
	report.Candidates = []Candidate{}
	return finish(report, false)
}
