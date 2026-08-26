package foundationseed

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/artifact/predecessorresolution"

func Evaluate(input predecessorresolution.Report, expectedHead string) Report {
	source := observeSource(input, expectedHead)
	report := Report{
		Schema: Schema, Decision: DecisionFailClosed, Reason: ReasonUnknown,
		Resolution: ResolutionLower, Source: source,
		NonClaims: nonClaims(), Authority: Authority{},
	}
	if source.ExactExhaustion {
		report.Decision = DecisionAuthorized
		report.Reason = ReasonExact
		report.Resolution = ResolutionExact
	}
	report.Indicators = indicators(source)
	report.Summary = summarize(report.Indicators)
	report.Views = views(report.Indicators)
	report.Proofs = proofs(source, report.Authority)
	report.Digest = seal(report).Digest
	return report
}

func summarize(values []Indicator) Summary {
	result := Summary{Total: len(values)}
	for _, value := range values {
		if value.Passed {
			result.Satisfied++
		}
	}
	result.BasisPoints = result.Satisfied * 10000 / result.Total
	return result
}

func nonClaims() []string {
	return append([]string(nil), fixedNonClaims...)
}
