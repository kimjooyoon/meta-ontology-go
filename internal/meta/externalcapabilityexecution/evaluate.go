package externalcapabilityexecution

func Evaluate(observation Observation) Report {
	report := Report{
		Schema: ReportSchema, SubjectSHA: observation.SubjectSHA,
		Total: MetricDenominator, Parent: observation.Parent,
		ExternalExecutions: observation.ExternalExecutions,
		RepositoryWrites: observation.RepositoryWrites,
		ExternalRepositoryWrites: observation.ExternalRepositoryWrites,
		ObservationDigest: observation.ObservationDigest,
	}
	report.Indicators = makeIndicators(observation)
	for _, metric := range report.Indicators {
		if metric.Status == StatusSatisfied {
			report.Completed++
		}
		if metric.Status == StatusUnknown {
			report.UnknownIndicators++
		}
		countClass(&report, metric)
	}
	report.BasisPoints = report.Completed * 10000 / report.Total
	report.Proofs = makeProofs(report.Indicators)
	switch {
	case report.UnknownIndicators > 0:
		report.Decision, report.Resolution = DecisionFailClosed, ResolutionUnknown
		report.EnforcementEffect, report.Reason = EffectBlock, ReasonUnknown
	case report.Completed != report.Total:
		report.Decision, report.Resolution = DecisionFailClosed, ResolutionInvariant
		report.EnforcementEffect, report.Reason = EffectBlock, ReasonInvariant
	default:
		report.Decision, report.Resolution = DecisionExecutable, ResolutionExact
		report.EnforcementEffect, report.Reason = EffectNoEffect, ReasonExecutable
	}
	report.ReportDigest = ""
	report.ReportDigest = digestValue(report)
	return report
}

func countClass(report *Report, metric Indicator) {
	satisfied := 0
	if metric.Status == StatusSatisfied {
		satisfied = 1
	}
	switch metric.Class {
	case "DRIVER":
		report.DriverTotal++
		report.DriverCompleted += satisfied
	case "OUTCOME":
		report.OutcomeTotal++
		report.OutcomeCompleted += satisfied
	case "GUARDRAIL":
		report.GuardrailTotal++
		report.GuardrailCompleted += satisfied
	}
}
