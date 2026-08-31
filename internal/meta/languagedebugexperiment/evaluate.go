package languagedebugexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"

func Evaluate(input Input) (Report, error) {
	if err := input.Contract.Validate(); err != nil {
		return Report{}, err
	}
	value, reason := collectFacts(input)
	indicators := buildIndicators(input.Contract, value)
	if value.Unknowns > 0 || len(value.RefutedCases) > 0 {
		for index := range indicators {
			indicators[index].Satisfied = false
		}
	}
	report := Report{
		Schema: "gooo/language-debug-experiment-report/v1", SubjectSHA: input.SubjectSHA,
		Decision: "PASS", Reason: "DEBUG_EXPERIMENT_SATISFIED", Resolution: "EXACT",
		Indicators: indicators, Views: buildViews(indicators),
		RepositoryWrites: value.RepositoryWrites, MutationAuthority: value.MutationAuthority,
		Replay: replayEvidence(input), RuntimeObservations: input.RuntimeObservations,
		Build: input.Build, EvaluatorBuild: input.EvaluatorBuild, Test: input.Test, Graph: input.Graph,
		UnknownCases: value.UnknownCases, RefutedCases: value.RefutedCases,
		Summary: summarize(value, input.ExecutableDigest, indicators)}
	if reason != "" {
		report.Reason = reason
		if len(value.RefutedCases) > 0 {
			report.Decision = "REFUTED"
			report.Resolution = "EXACT"
		} else if value.Unknowns > 0 {
			report.Decision = "UNKNOWN"
			report.Resolution = "LOWER_RESOLUTION"
			for _, unknown := range value.UnknownCases {
				if unknown.UnknownClass == "UNKNOWN_DECISION" || unknown.UnknownClass == "MALFORMED_EVIDENCE" {
					report.Decision = "FAIL_CLOSED"
					break
				}
			}
		} else {
			report.Decision = "FAIL_CLOSED"
		}
	} else if report.Summary.Coordinates.Satisfied != report.Summary.Coordinates.Total {
		report.Decision = "FAIL_CLOSED"
		report.Reason = "DEBUG_EXPERIMENT_NOT_SATISFIED"
	}
	return sealReport(report), nil
}

func replayEvidence(input Input) ReplayEvidence {
	first, second := input.First.DeterministicDigest(), input.Second.DeterministicDigest()
	return ReplayEvidence{Schema: languagedebug.DeterministicPayloadSchema, RuntimeReceiptSchema: RuntimeReceiptSchema, FirstDigest: first,
		SecondDigest: second, Equal: first == second,
		ExcludedFields: languagedebug.DeterministicExcludedFields()}
}

func summarize(value facts, executable string, indicators []Indicator) Summary {
	return Summary{
		Coordinates: coordinates(indicators), DebugReceipts: value.DebugReceipts,
		PausedSessions: value.PausedSessions, BreakpointsReached: value.BreakpointsReached,
		TraceEvents: value.TraceEvents, ExecutionDigestVariants: value.ExecutionDigestVariants,
		ReplayMatches: value.ReplayMatches, ResourceObservations: value.ResourceObservations,
		CurrentEvents: value.CurrentEvents, RemainingEvents: value.RemainingEvents,
		UnknownBreakpointRejections: value.UnknownBreakpointRejections, Unknowns: value.Unknowns,
		RefutedCases: len(value.RefutedCases),
		Compiler:     Compiler{ExecutableDigest: executable, Go127Runtimes: value.Go127Runtimes},
		Effects:      Effects{RepositoryWrites: value.RepositoryWrites, MutationAuthority: value.MutationAuthority},
	}
}

func coordinates(indicators []Indicator) Counter {
	counter := Counter{Total: len(indicators)}
	for _, indicator := range indicators {
		if indicator.Satisfied {
			counter.Satisfied++
		}
	}
	counter.BasisPoints = basisPoints(counter.Satisfied, counter.Total)
	return counter
}

func basisPoints(value, total int) int {
	if total == 0 {
		return 0
	}
	return value * 10000 / total
}
