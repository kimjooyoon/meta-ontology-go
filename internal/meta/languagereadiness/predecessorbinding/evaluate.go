package predecessorbinding

func Evaluate(headSHA string, observations []Observation, repositoryWrites int) Report {
	report := Report{
		Schema:           Schema,
		RegistrySchema:   RegistrySchema,
		RegistryDigest:   registryDigest(),
		HeadSHA:          headSHA,
		UseCase:          UseCase,
		Decision:         DecisionPass,
		Reason:           ReasonExactlyCounted,
		RepositoryWrites: repositoryWrites,
	}
	byID := make(map[string][]Observation)
	for _, observation := range observations {
		byID[observation.ID] = append(byID[observation.ID], observation)
	}
	for _, coordinate := range coordinates {
		evidence := classifyEvidence(coordinate, byID[coordinate.ID])
		report.Evidence = append(report.Evidence, evidence)
		switch evidence.State {
		case StateStaticLiteral:
			report.Summary.StaticLiteral++
		case StateDynamicInput:
			report.Summary.DynamicInput++
		default:
			report.Summary.Unknown++
		}
	}
	report.Summary.Total = Total
	report.Summary.DynamicBPS = report.Summary.DynamicInput * 10_000 / Total
	switch {
	case repositoryWrites != 0:
		report.Decision, report.Reason = DecisionFailClosed, ReasonWriteEffect
	case !validSHA(headSHA):
		report.Decision, report.Reason = DecisionFailClosed, ReasonHeadUnknown
	case report.Summary.Unknown != 0:
		report.Decision, report.Reason = DecisionFailClosed, ReasonEvidenceUnknown
	}
	report.Indicators = indicators(report)
	report.Proofs = proofs(report)
	report.ReportDigest = digestJSON(report)
	return report
}

func classifyEvidence(coordinate Coordinate, observed []Observation) Evidence {
	base := Observation{ID: coordinate.ID, GoField: coordinate.GoField,
		SourcePath: SourcePath, Provider: Provider, State: StateUnknown}
	if len(observed) != 1 {
		return Evidence{Observation: base, Reason: "OBSERVATION_CARDINALITY_INVALID"}
	}
	value := observed[0]
	if value.GoField != coordinate.GoField || value.SourcePath != SourcePath ||
		value.Provider != Provider {
		return Evidence{Observation: base, Reason: "OBSERVATION_BINDING_MISMATCH"}
	}
	switch value.State {
	case StateStaticLiteral:
		return Evidence{Observation: value, Reason: "COMPILE_TIME_COORDINATE"}
	case StateDynamicInput:
		return Evidence{Observation: value, Reason: "PROVIDER_PARAMETER_COORDINATE"}
	default:
		return Evidence{Observation: base, Reason: "OBSERVATION_STATE_UNKNOWN"}
	}
}

func indicators(report Report) []Indicator {
	summary := report.Summary
	return []Indicator{
		newIndicator("gooo.metric.language.predecessor-dynamic-binding-bps.v1", "outcome",
			"COHERENCE", "classify-predecessor-bindings", summary.DynamicBPS, 10_000, "BASIS_POINT"),
		newIndicator("gooo.metric.language.predecessor-dynamic-coordinates.v1", "driver",
			"COHERENCE", "count-dynamic-predecessor-coordinates", summary.DynamicInput, Total, "COORDINATE"),
		newIndicator("gooo.metric.language.predecessor-static-coordinates.guardrail.v1", "guardrail",
			"REGRESSION", "count-static-predecessor-coordinates", summary.StaticLiteral, 0, "COORDINATE"),
		newIndicator("gooo.metric.language.predecessor-unknown-coordinates.guardrail.v1", "guardrail",
			"FOUNDATION", "lower-resolution-on-unknown-coordinate", summary.Unknown, 0, "COORDINATE"),
		newIndicator("gooo.metric.language.predecessor-observer-writes.guardrail.v1", "guardrail",
			"FOUNDATION", "preserve-read-only-observation", report.RepositoryWrites, 0, "REPOSITORY_WRITE"),
	}
}

func newIndicator(metricID, class, proof, operation string, value, target int, unit string) Indicator {
	satisfied := value >= target
	if class == "guardrail" {
		satisfied = value <= target
	}
	return Indicator{MetricID: metricID, Class: class, ProofChoice: proof,
		Producer: "predecessorbinding.Evaluate", Consumer: "self-improvement-cycle",
		MetaOperation: operation, Value: value, Target: target, Unit: unit,
		Satisfied: satisfied}
}

func proofs(report Report) []Proof {
	return []Proof{
		{ID: "fixed-eight-coordinate-registry", Choice: "FOUNDATION", Passed: len(coordinates) == Total},
		{ID: "ast-binding-classification", Choice: "COHERENCE", Passed: report.Summary.Unknown == 0},
		{ID: "unknown-fails-closed", Choice: "REGRESSION", Passed: report.Summary.Unknown == 0},
		{ID: "read-only-observer", Choice: "FOUNDATION", Passed: report.RepositoryWrites == 0},
	}
}
