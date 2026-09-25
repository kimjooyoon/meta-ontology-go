package rollbackintegrityshadow

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/rollbackfixedpoint"

func runSuite() ([]CaseResult, Summary) {
	specs := caseSpecs()
	results := make([]CaseResult, 0, len(specs))
	summary := Summary{CasesTotal: len(specs)}
	for _, spec := range specs {
		source := baseSource()
		if spec.mutate != nil {
			spec.mutate(&source)
		}
		report := rollbackfixedpoint.Build(source)
		err := rollbackfixedpoint.Validate(report)
		passed := err == nil && report.Decision == spec.decision &&
			report.Resolution == spec.resolution && report.Mode == spec.mode
		result := CaseResult{Name: spec.name, ExpectedDecision: spec.decision,
			ActualDecision: report.Decision, ExpectedResolution: spec.resolution,
			ActualResolution: report.Resolution, ExpectedMode: spec.mode, ActualMode: report.Mode,
			Satisfied: report.Summary.Satisfied, CoordinatesTotal: report.Summary.Total,
			Unresolved: report.Summary.Unresolved, NotSatisfied: report.Summary.NotSatisfied,
			RepositoryWrites: report.RepositoryWrites, ReportDigest: report.ReportDigest, Passed: passed}
		if err != nil {
			result.ValidationError = err.Error()
		} else {
			summary.MetaReportsValid++
		}
		if passed {
			summary.CasesPassed++
		}
		summary.CoordinatesTotal += report.Summary.Total
		classifyCase(spec, &summary)
		results = append(results, result)
	}
	summary.CaseCoverageBPS = summary.CasesPassed * 10000 / summary.CasesTotal
	return results, summary
}

func classifyCase(spec caseSpec, summary *Summary) {
	if spec.decision == "PASS" {
		summary.TerminalCases++
	} else if spec.resolution == ResolutionLower {
		summary.UnknownDecisionCases++
	} else {
		summary.KnownRejectCases++
	}
}
