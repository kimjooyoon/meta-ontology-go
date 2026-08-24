package externalecosystemconformance

import "encoding/json"

func Evaluate(subject string, capsule Capsule, evidence Evidence) Report {
	report := baseReport(subject, capsule.ReferenceID, evidence)
	resolution, reason, exactDocuments := validateEvidence(evidence)
	report.Summary.DocumentsExact = exactDocuments
	if resolution != "" {
		return fail(report, resolution, reason)
	}
	resolution, reason = validateCapsule(capsule)
	if resolution != "" {
		return fail(report, resolution, reason)
	}
	report.Summary.CommitExact = 1
	report.Summary.CapabilityMappings = len(capabilityRules)
	if !validModule(evidence.GoMod) {
		return fail(report, ResolutionInvariant, ReasonModuleMismatch)
	}
	report.Summary.ModuleExact = 1
	if evidence.ExternalExecutions > 0 {
		return fail(report, ResolutionInvariant, ReasonExecution)
	}
	if evidence.RepositoryWrites > 0 {
		return fail(report, ResolutionInvariant, ReasonWrite)
	}
	report.Decision = DecisionReferenceBound
	report.Resolution = ResolutionExact
	report.EnforcementEffect = EffectNoEffect
	report.Reason = ReasonReferenceBound
	report.Summary.BoundCapabilities = len(capabilityRules)
	report.Summary.MappingCoverageBPS = 10000
	return finish(report)
}

func baseReport(subject, referenceID string, evidence Evidence) Report {
	return Report{
		Schema:             ReportSchema,
		SubjectSHA:         subject,
		ReferenceID:        referenceID,
		RepositoryWrites:   evidence.RepositoryWrites,
		ExternalExecutions: evidence.ExternalExecutions,
		Summary: Summary{
			ReferenceDenominator:    len(capabilityRules),
			DocumentsTotal:          2,
			LanguageDenominator:     12,
			LanguageBeforeOperating: 10,
			LanguageOfficialAfter:   10,
			ObservedWrites:          evidence.RepositoryWrites,
			ObservedExecutions:      evidence.ExternalExecutions,
		},
	}
}

func fail(report Report, resolution, reason string) Report {
	report.Decision = DecisionFailClosed
	report.Resolution = resolution
	report.EnforcementEffect = EffectBlock
	report.Reason = reason
	report.Summary.BlockedPaths = 1
	if resolution == ResolutionUnknown {
		report.Summary.UnknownPaths = 1
	}
	return finish(report)
}

func finish(report Report) Report {
	report.Indicators = makeIndicators(report)
	report.ReportDigest = ""
	raw, _ := json.Marshal(report)
	report.ReportDigest = digest(raw)
	return report
}
