package languagedebugexperiment

func Evaluate(input Input) (Report, error) {
	if err := input.Contract.Validate(); err != nil {
		return Report{}, err
	}
	value, reason := collectFacts(input)
	indicators := buildIndicators(input.Contract, value)
	if value.Unknowns > 0 {
		for index := range indicators {
			indicators[index].Satisfied = false
		}
	}
	report := Report{
		Schema: "gooo/language-debug-experiment-report/v1", SubjectSHA: input.SubjectSHA,
		Decision: "PASS", Reason: "DEBUG_EXPERIMENT_SATISFIED", Resolution: "EXACT",
		Indicators: indicators, Views: buildViews(indicators),
		RepositoryWrites: value.RepositoryWrites, MutationAuthority: value.MutationAuthority,

		Summary: summarize(value, input.ExecutableDigest, indicators)}
	if reason != "" {
		report.Decision = "FAIL_CLOSED"
		report.Reason = reason
		if value.Unknowns > 0 {
			report.Resolution = "LOWER_RESOLUTION"
		}
	} else if report.Summary.Coordinates.Satisfied != report.Summary.Coordinates.Total {
		report.Decision = "FAIL_CLOSED"
		report.Reason = "DEBUG_EXPERIMENT_NOT_SATISFIED"
	}
	return sealReport(report), nil
}

func summarize(value facts, executable string, indicators []Indicator) Summary {
	return Summary{
		Coordinates: coordinates(indicators), DebugReceipts: value.DebugReceipts,
		PausedSessions: value.PausedSessions, BreakpointsReached: value.BreakpointsReached,
		TraceEvents: value.TraceEvents, ExecutionDigestVariants: value.ExecutionDigestVariants,
		CurrentEvents: value.CurrentEvents, RemainingEvents: value.RemainingEvents,
		UnknownBreakpointRejections: value.UnknownBreakpointRejections, Unknowns: value.Unknowns,
		Compiler: Compiler{ExecutableDigest: executable, Go127Runtimes: value.Go127Runtimes},
		Effects:  Effects{RepositoryWrites: value.RepositoryWrites, MutationAuthority: value.MutationAuthority},
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
