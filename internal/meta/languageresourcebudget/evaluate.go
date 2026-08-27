package languageresourcebudget

import "sort"

func Evaluate(input Input, caseName string) Report {
	report := baseReport(input, caseName)
	if input.Schema != InputSchema {
		return closeReport(report, "EXACT", "INPUT_SCHEMA_UNKNOWN", "NO_CLAIM")
	}
	if !positiveSHA(input.ExpectedHead) || !validContract(input.Contract) {
		return closeReport(report, "EXACT", "CONTRACT_OR_SUBJECT_INVALID", "NO_CLAIM")
	}
	if !validRunner(input.Producer.Runner) {
		return closeReport(report, "LOWER_RESOLUTION", "RUNNER_IDENTITY_UNKNOWN", "SEMANTIC_EXACT_RESOURCE_UNKNOWN")
	}
	report.Semantic, _ = verifyProducer(input)
	if report.Semantic.Reason == "" {
		report.Semantic.Reason = "SEMANTIC_PRODUCER_EVIDENCE_INVALID"
	}
	semanticErr := report.Semantic.Decision != "PASS"
	complete, summaries, budgetViolations, missing := summarizeResources(input)
	report.Summary.Resources = summaries
	report.Summary.Operations = len(input.Contract.Operations)
	report.Summary.Samples = len(input.Observations)
	report.Summary.SourceFiles = input.Producer.SourceFiles
	report.Summary.GoFiles = input.Producer.GoFiles
	report.Summary.Effects = input.Producer.Effects
	report.Summary.Semantic = report.Semantic
	report.Effects = input.Producer.Effects
	report.Indicators = buildIndicators(input, summaries, complete, semanticErr, budgetViolations)
	report.Summary.Coordinates = coordinates(report.Indicators)
	report.Cases = []CaseResult{caseResult(caseName, complete, missing, budgetViolations, semanticErr)}
	report.Transitions = buildTransitions(report.Semantic, complete, budgetViolations, semanticErr)
	report.NotClaimed = append([]string(nil), input.Contract.NotClaimed...)
	if input.Producer.Effects.RepositoryWrites != 0 || input.Producer.Effects.MutationAuthority {
		return finish(report, "EXACT", "EFFECT_BOUNDARY_VIOLATED", "NO_CLAIM")
	}
	if semanticErr {
		return finish(report, "EXACT", report.Semantic.Reason, "NO_SEMANTIC_CLAIM")
	}
	if missing {
		return finish(report, "LOWER_RESOLUTION", "RESOURCE_SAMPLE_MISSING", "SEMANTIC_EXACT_RESOURCE_UNKNOWN")
	}
	if !complete {
		return finish(report, "EXACT", "RESOURCE_SAMPLE_INVALID", "SEMANTIC_EXACT_RESOURCE_CLAIM_REJECTED")
	}
	if budgetViolations > 0 {
		return finish(report, "EXACT", "RESOURCE_BUDGET_EXCEEDED", "SEMANTIC_EXACT_RESOURCE_CLAIM_REFUTED")
	}
	return finish(report, "EXACT", "RESOURCE_ENVELOPE_OBSERVED", "SEMANTIC_EXACT_RESOURCE_RUNNER_SCOPED")
}

func baseReport(input Input, caseName string) Report {
	return Report{Schema: ReportSchema, Case: caseName, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
		Interpretation: "NO_CLAIM", ResourceResolution: "LOWER_RESOLUTION", Summary: Summary{Resources: []ResourceSummary{}},
		Effects: input.Producer.Effects}
}

func finish(report Report, resolution, reason, interpretation string) Report {
	report.Resolution, report.Reason, report.Interpretation = resolution, reason, interpretation
	if reason == "RESOURCE_ENVELOPE_OBSERVED" {
		report.Decision = "PASS"
	}
	if report.Semantic.Decision == "PASS" && !hasMissing(report) {
		report.ResourceResolution = "RUNNER_SCOPED"
	}
	report.FactsDigest = digestValue(struct {
		Semantic   Semantic
		Summary    Summary
		Indicators []Indicator
	}{report.Semantic, report.Summary, report.Indicators})
	previous := ""
	for index := range report.Transitions {
		report.Transitions[index].Evidence = report.FactsDigest
		report.Transitions[index].PreviousDigest = previous
		report.Transitions[index] = sealTransition(report.Transitions[index])
		previous = report.Transitions[index].Digest
	}
	return sealReport(report)
}

func closeReport(report Report, resolution, reason, interpretation string) Report {
	return finish(report, resolution, reason, interpretation)
}

func summarizeResources(input Input) (bool, []ResourceSummary, int, bool) {
	values := append([]Observation(nil), input.Observations...)
	sort.Slice(values, func(left, right int) bool {
		if values[left].Operation != values[right].Operation {
			return values[left].Operation < values[right].Operation
		}
		return values[left].Sequence < values[right].Sequence
	})
	complete := len(values) == len(input.Contract.Operations)*input.Contract.SamplesPerOp
	missing := false
	operationViolations := 0
	summaries := make([]ResourceSummary, 0, len(input.Contract.Operations))
	for _, spec := range input.Contract.Operations {
		operationViolations := 0
		group := make([]Observation, 0, input.Contract.SamplesPerOp)
		for _, value := range values {
			if value.Operation == spec.ID {
				group = append(group, value)
			}
		}
		if len(group) != input.Contract.SamplesPerOp {
			complete, missing = false, true
		}
		for index, value := range group {
			if value.Sequence != index+1 || !validObservation(value, spec, input) {
				complete = false
			}
			operationViolations += budgetViolation(value, input.Contract.Limits)
		}
		walls := make([]int64, len(group))
		for index := range group {
			walls[index] = group[index].WallTimeNS
		}
		sort.Slice(walls, func(i, j int) bool { return walls[i] < walls[j] })
		summary := ResourceSummary{Operation: spec.ID, Samples: len(group), BudgetViolations: operationViolations}
		if len(walls) > 0 {
			summary.WallMinNS, summary.WallMedianNS, summary.WallMaxNS = walls[0], walls[len(walls)/2], walls[len(walls)-1]
		}
		for _, value := range group {
			if value.PeakRSSKiB > summary.PeakRSSMaxKiB {
				summary.PeakRSSMaxKiB = value.PeakRSSKiB
			}
			if value.ReceiptBytes > summary.ReceiptMax {
				summary.ReceiptMax = value.ReceiptBytes
			}
			if value.GeneratedBytes > summary.GeneratedMax {
				summary.GeneratedMax = value.GeneratedBytes
			}
		}
		summaries = append(summaries, summary)
	}
	violations := 0
	for _, summary := range summaries {
		violations += summary.BudgetViolations
	}
	return complete, summaries, violations, missing
}

func validObservation(value Observation, spec Operation, input Input) bool {
	return value.Schema == ObservationSchema && value.SubjectSHA == input.ExpectedHead && value.Producer == Producer && value.Consumer == Consumer &&
		value.Stage == spec.Stage && value.Step == spec.Step && value.MetaOperation == spec.MetaOperation && value.ProofChoice == spec.ProofChoice &&
		value.Reason == "RUNNER_RESOURCE_OBSERVED" && value.ExitCode == 0 && value.WallTimeNS > 0 && value.PeakRSSKiB > 0 &&
		value.ReceiptBytes >= 0 && value.GeneratedBytes >= 0 && len(value.OutputDigest) > 0
}

func budgetViolation(value Observation, limits Limits) int {
	limitNS := limits.WallTimeMS * 1000000
	if value.WallTimeNS > limitNS || value.PeakRSSKiB > limits.PeakRSSKiB || value.ReceiptBytes > limits.ReceiptBytes || value.GeneratedBytes > limits.GeneratedBytes {
		return 1
	}
	return 0
}

func coordinates(values []Indicator) Counter {
	satisfied := 0
	for _, value := range values {
		if value.Satisfied {
			satisfied++
		}
	}
	basis := 0
	if len(values) > 0 {
		basis = satisfied * 10000 / len(values)
	}
	return Counter{Satisfied: satisfied, Total: len(values), BasisPoints: basis}
}

func hasMissing(report Report) bool { return report.Reason == "RESOURCE_SAMPLE_MISSING" }

func caseResult(name string, complete, missing bool, violations int, semanticErr bool) CaseResult {
	result := CaseResult{Name: name, Decision: "PASS", Resolution: "EXACT", Reason: "CASE_EXPECTATION_MET"}
	switch {
	case semanticErr:
		result.Decision, result.Reason, result.Impact = "FAIL_CLOSED", "SEMANTIC_CLAIM_REJECTED", "SEMANTIC_CLAIM"
	case missing:
		result.Decision, result.Resolution, result.Reason, result.Impact = "FAIL_CLOSED", "LOWER_RESOLUTION", "RESOURCE_SAMPLE_MISSING", "RESOURCE_CLAIM_ONLY_AND_RESOLUTION_LOWERED"
	case !complete:
		result.Decision, result.Reason, result.Impact = "FAIL_CLOSED", "RESOURCE_SAMPLE_INVALID", "RESOURCE_CLAIM_ONLY"
	case violations > 0:
		result.Decision, result.Reason, result.Impact = "FAIL_CLOSED", "RESOURCE_BUDGET_EXCEEDED", "RESOURCE_CLAIM_ONLY"
	}
	return result
}
