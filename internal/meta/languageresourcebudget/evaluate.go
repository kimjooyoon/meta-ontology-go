package languageresourcebudget

import (
	"bytes"
	"encoding/base64"
	"sort"
)

func Evaluate(input Input, caseName string) Report {
	report := baseReport(input, caseName)
	if input.Schema != InputSchema {
		return closeReport(report, "EXACT", "INPUT_SCHEMA_UNKNOWN", "NO_CLAIM")
	}
	if !positiveSHA(input.ExpectedHead) || !validContract(input.Contract) || input.ContractDigest != digestValue(input.Contract) {
		return closeReport(report, "EXACT", "CONTRACT_OR_SUBJECT_INVALID", "NO_CLAIM")
	}
	if !validRunner(input.Producer.Runner) {
		return closeReport(report, "LOWER_RESOLUTION", "RUNNER_IDENTITY_UNKNOWN", "SEMANTIC_EXACT_RESOURCE_UNKNOWN")
	}
	report.WriteSets = append([]WriteSetObservation(nil), input.Producer.WriteSets...)
	report.Summary.WriteSets = append([]WriteSetObservation(nil), input.Producer.WriteSets...)
	writeSetTo, writeSetReason := writeSetTransition(input)
	report.ReadOnlyResolution = "EXACT"
	if writeSetTo == "OPEN" {
		report.ReadOnlyResolution = "LOWER_RESOLUTION"
	}
	report.Semantic, _ = verifyProducer(input)
	if report.Semantic.Reason == "" {
		report.Semantic = Semantic{Decision: "FAIL_CLOSED", Resolution: "EXACT", ClaimState: "REFUTED", Reason: "SEMANTIC_PRODUCER_EVIDENCE_INVALID"}
	}
	complete, summaries, budgetViolations, missing := summarizeResources(input, report.Semantic)
	report.Summary.Resources = summaries
	report.Summary.Operations = len(input.Contract.Operations)
	report.Summary.Samples = len(input.Observations)
	report.Summary.SourceFiles = input.Producer.SourceFileCount
	report.Summary.GoFiles = input.Producer.GoFiles
	report.Summary.Runner = input.Producer.Runner
	report.Summary.Effects = input.Producer.Effects
	report.Summary.EvidenceDigest = rawEvidenceDigest(input)
	report.Summary.Semantic = report.Semantic
	report.Effects = input.Producer.Effects
	report.Indicators = buildIndicators(input, summaries, complete, report.Semantic, budgetViolations)
	report.Summary.Coordinates = coordinates(report.Indicators)
	report.Summary.Unknowns = report.Summary.Coordinates.Unknown
	report.Cases = []CaseResult{caseResult(caseName, input.EvidenceClass, complete, missing, budgetViolations, report.Semantic)}
	report.NotClaimed = append([]string(nil), input.Contract.NotClaimed...)
	report.Transitions = buildTransitions(report.Semantic, complete, missing, budgetViolations, writeSetTo, writeSetReason)
	if writeSetTo == "REFUTED" {
		return finish(report, "EXACT", "EFFECT_BOUNDARY_VIOLATED", "NO_CLAIM")
	}
	if writeSetTo == "OPEN" {
		return finish(report, "LOWER_RESOLUTION", "EFFECT_OBSERVATION_MISSING", "SEMANTIC_EXACT_RESOURCE_UNKNOWN")
	}
	if report.Semantic.ClaimState == "OPEN" {
		return finish(report, "LOWER_RESOLUTION", report.Semantic.Reason, "SEMANTIC_CLAIM_OPEN_RESOURCE_UNKNOWN")
	}
	if report.Semantic.ClaimState == "REFUTED" {
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
	return Report{Schema: ReportSchema, Case: caseName, EvidenceClass: input.EvidenceClass, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
		Interpretation: "NO_CLAIM", ResourceResolution: "LOWER_RESOLUTION", ReadOnlyResolution: "LOWER_RESOLUTION", Summary: Summary{Resources: []ResourceSummary{}},
		Effects: input.Producer.Effects}
}

func finish(report Report, resolution, reason, interpretation string) Report {
	report.Resolution, report.Reason, report.Interpretation = resolution, reason, interpretation
	if reason == "RESOURCE_ENVELOPE_OBSERVED" {
		report.Decision = "PASS"
	}
	if len(report.Transitions) >= 2 && report.Transitions[0].To == "DISCHARGED" && report.Transitions[1].To != "OPEN" && report.Transitions[2].To == "DISCHARGED" {
		report.ResourceResolution = "RUNNER_SCOPED"
	} else {
		report.ResourceResolution = "LOWER_RESOLUTION"
	}
	if len(report.Transitions) >= 3 && report.Transitions[2].To == "OPEN" {
		report.ReadOnlyResolution = "LOWER_RESOLUTION"
	} else {
		report.ReadOnlyResolution = "EXACT"
	}
	report.FactsDigest = digestValue(struct {
		Semantic   Semantic
		Summary    Summary
		Indicators []Indicator
	}{report.Semantic, report.Summary, report.Indicators})
	report.ProvenanceDigest = report.FactsDigest
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

func rawEvidenceDigest(input Input) string {
	return digestValue(struct {
		Sources      []RawSource
		Outputs      []RawOutput
		Observations []Observation
		WriteSets    []WriteSetObservation
	}{input.Producer.SourceFiles, input.Producer.RawOutputs, input.Observations, input.Producer.WriteSets})
}

func writeSetTransition(input Input) (string, string) {
	if input.Producer.Effects.RepositoryWrites != 0 || input.Producer.Effects.MutationAuthority {
		return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
	}
	if len(input.Producer.WriteSets) != len(input.Contract.Operations) {
		return "OPEN", "EFFECT_OBSERVATION_MISSING"
	}
	seen := make(map[string]bool, len(input.Producer.WriteSets))
	open := false
	for _, value := range input.Producer.WriteSets {
		if seen[value.Operation] {
			return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
		}
		seen[value.Operation] = true
		if value.RepositoryWrites != 0 || value.MutationAuthority || value.DiffExitCode != 0 || len(value.ChangedPaths) != 0 || value.UntrackedFileCount != 0 {
			return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
		}
		if value.Schema != "gooo/meta-resource-budget-write-set/v1" || value.Producer != Producer || value.Consumer != Consumer ||
			!gitDigest(value.BeforeTreeDigest) || !gitDigest(value.AfterTreeDigest) || !contentDigest(value.WriteSetDigest) ||
			value.BeforeTreeDigest != value.AfterTreeDigest || !value.AuthorityObserved || !value.BeforeStatusObserved || !value.AfterStatusObserved || value.MutationAuthority ||
			value.SampleStart != 1 || value.SampleEnd != input.Contract.SamplesPerOp ||
			value.Reason != "NET_REPOSITORY_STATE_UNCHANGED_ACROSS_OPERATION_WINDOW" {
			open = true
			continue
		}
		before, beforeErr := decodeStatus(value.BeforeStatusBase64)
		after, afterErr := decodeStatus(value.AfterStatusBase64)
		if beforeErr != nil || afterErr != nil || value.WriteSetDigest != statusDigest(before, after) {
			open = true
			continue
		}
		if !bytes.Equal(before, after) {
			return "REFUTED", "EFFECT_BOUNDARY_VIOLATED"
		}
	}
	for _, spec := range input.Contract.Operations {
		if !seen[spec.ID] {
			open = true
		}
	}
	if open {
		return "OPEN", "EFFECT_OBSERVATION_MISSING"
	}
	return "DISCHARGED", "NET_REPOSITORY_STATE_UNCHANGED"
}

func writeSetTransitionOnly(input Input) string {
	value, _ := writeSetTransition(input)
	return value
}

func decodeStatus(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(value)
}

func statusDigest(before, after []byte) string {
	value := append(append([]byte(nil), before...), 0)
	value = append(value, after...)
	return digestBytes(value)
}

func summarizeResources(input Input, semantic Semantic) (bool, []ResourceSummary, int, bool) {
	byOperation := make(map[string][]Observation, len(input.Contract.Operations))
	for _, value := range input.Observations {
		byOperation[value.Operation] = append(byOperation[value.Operation], value)
	}
	complete, missing := true, false
	summaries := make([]ResourceSummary, 0, len(input.Contract.Operations))
	for _, spec := range input.Contract.Operations {
		group := append([]Observation(nil), byOperation[spec.ID]...)
		sort.SliceStable(group, func(left, right int) bool { return group[left].Sequence < group[right].Sequence })
		summary := ResourceSummary{Operation: spec.ID, Samples: len(group)}
		if len(group) < input.Contract.SamplesPerOp {
			summary.MissingSamples = input.Contract.SamplesPerOp - len(group)
			complete, missing = false, true
		}
		if len(group) > input.Contract.SamplesPerOp {
			summary.InvalidSamples = len(group) - input.Contract.SamplesPerOp
			complete = false
		}
		walls := make([]int64, 0, len(group))
		for _, value := range group {
			if value.Sequence < 1 || value.Sequence > input.Contract.SamplesPerOp || !validObservation(value, spec, input, semantic) {
				summary.InvalidSamples++
				complete = false
				continue
			}
			walls = append(walls, value.WallTimeNS)
			if budgetViolation(value, input.Contract.Limits) > 0 {
				summary.BudgetViolations++
			}
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
		sort.Slice(walls, func(left, right int) bool { return walls[left] < walls[right] })
		if len(walls) > 0 {
			summary.WallMinNS, summary.WallMedianNS, summary.WallMaxNS = walls[0], walls[len(walls)/2], walls[len(walls)-1]
		}
		summaries = append(summaries, summary)
	}
	for operationID, values := range byOperation {
		if _, ok := operation(operationID, input.Contract); !ok && len(values) > 0 {
			complete = false
		}
	}
	violations := 0
	for _, summary := range summaries {
		violations += summary.BudgetViolations
	}
	return complete, summaries, violations, missing
}

func validObservation(value Observation, spec Operation, input Input, semantic Semantic) bool {
	return value.Schema == ObservationSchema && value.SubjectSHA == input.ExpectedHead && value.Producer == Producer && value.Consumer == Consumer &&
		value.Operation == spec.ID && value.Stage == spec.Stage && value.Step == spec.Step && value.MetaOperation == spec.MetaOperation && value.ProofChoice == spec.ProofChoice &&
		value.Reason == "RUNNER_RESOURCE_OBSERVED" && value.ExitCode == 0 && value.WallTimeNS > 0 && value.PeakRSSKiB > 0 &&
		value.ReceiptBytes >= 0 && value.GeneratedBytes >= 0 && contentDigest(value.OutputDigest) &&
		value.SourceRawDigest == semantic.SourceDigest && value.SourceSemanticDigest == semantic.SemanticDigest && value.EntryDigest == semantic.TargetDigest && value.TargetDigest == semantic.TargetDigest
}

func budgetViolation(value Observation, limits Limits) int {
	if value.WallTimeNS > limits.WallTimeMS*1000000 || value.PeakRSSKiB > limits.PeakRSSKiB || value.ReceiptBytes > limits.ReceiptBytes || value.GeneratedBytes > limits.GeneratedBytes {
		return 1
	}
	return 0
}

func coordinates(values []Indicator) Counter {
	result := Counter{}
	for _, value := range values {
		switch value.Status {
		case "NOT_APPLICABLE":
			continue
		case "SATISFIED":
			result.Satisfied++
		case "REFUTED":
			result.Refuted++
		case "UNKNOWN":
			result.Unknown++
		}
	}
	result.Total = result.Satisfied + result.Refuted + result.Unknown
	if result.Total > 0 {
		result.BasisPoints = result.Satisfied * 10000 / result.Total
	}
	return result
}

func caseResult(name, evidenceClass string, complete, missing bool, violations int, semantic Semantic) CaseResult {
	result := CaseResult{Name: name, EvidenceClass: evidenceClass, Decision: "PASS", Resolution: "EXACT", Reason: "CASE_EXPECTATION_MET"}
	switch {
	case semantic.ClaimState == "OPEN":
		result.Decision, result.Resolution, result.Reason, result.Impact = "FAIL_CLOSED", "LOWER_RESOLUTION", semantic.Reason, "SEMANTIC_CLAIM_OPEN"
	case semantic.ClaimState == "REFUTED":
		result.Decision, result.Reason, result.Impact = "FAIL_CLOSED", semantic.Reason, "SEMANTIC_CLAIM"
	case missing:
		result.Decision, result.Resolution, result.Reason, result.Impact = "FAIL_CLOSED", "LOWER_RESOLUTION", "RESOURCE_SAMPLE_MISSING", "RESOURCE_CLAIM_ONLY_AND_RESOLUTION_LOWERED"
	case !complete:
		result.Decision, result.Reason, result.Impact = "FAIL_CLOSED", "RESOURCE_SAMPLE_INVALID", "RESOURCE_CLAIM_ONLY"
	case violations > 0:
		result.Decision, result.Reason, result.Impact = "FAIL_CLOSED", "RESOURCE_BUDGET_EXCEEDED", "RESOURCE_CLAIM_ONLY"
	}
	return result
}
