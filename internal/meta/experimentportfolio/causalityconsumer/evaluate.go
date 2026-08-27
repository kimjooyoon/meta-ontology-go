package causalityconsumer

import "reflect"

type receiptSemanticProjection struct {
	CandidateID       string
	SourcePath        string
	Producer          string
	Consumer          string
	MetaOperation     string
	ProofChoice       string
	SemanticValue     string
	Decision          string
	ClaimTransitions  []ClaimTransition
	CoordinateVector  []Coordinate
	Counterexamples   []Counterexample
	UnknownLocations  []UnknownLocation
	ExtensionEvidence []ExtensionEvidence
	RepositoryWrites  int
	MutationAuthority bool
}

func EvaluateCausality(input CausalityInput) CausalityReport {
	if reason := causalityInputReason(input); reason != "" {
		return closedCausality(input, reason)
	}
	samplesByCandidate := make(map[string]CausalitySampleInput, len(input.Samples))
	for _, sample := range input.Samples {
		samplesByCandidate[sample.CandidateID] = sample
	}
	results := make([]CausalitySampleResult, 0, len(input.Contract.Candidates))
	unknowns := make([]CausalityUnknown, 0)
	summary := CausalitySummary{CausalCases: CausalCaseCount{Total: len(input.Contract.Candidates)}}
	for _, candidate := range input.Contract.Candidates {
		result, findings := evaluateCausalitySample(samplesByCandidate[candidate.ID], input.Manifest)
		results = append(results, result)
		unknowns = append(unknowns, findings...)
		if result.Status == "DISCHARGED" {
			summary.CausalCases.Observed++
		}
		if result.DigestOnlyBinding {
			summary.DigestOnlyCases++
		}
		if result.HardcodedFixture {
			summary.HardcodedFixtureCases++
		}
	}
	summary.Unknowns = len(unknowns)
	decision := "CAUSALITY_DISCHARGED"
	reason := "SOURCE_SEMANTIC_CAUSALITY_DISCHARGED"
	resolution := "EXACT"
	if len(unknowns) > 0 {
		decision = "FAIL_CLOSED"
		reason = "CAUSALITY_INPUT_UNKNOWN"
		resolution = "LOWER_RESOLUTION"
	} else if summary.DigestOnlyCases > 0 {
		decision = "REFUTED"
		reason = "DIGEST_ONLY_BINDING"
	} else if summary.CausalCases.Observed != summary.CausalCases.Total {
		decision = "REFUTED"
		reason = "SEMANTIC_CAUSALITY_CONTRACT_REFUTED"
	}
	report := CausalityReport{
		Schema:            CausalityReportSchema,
		Decision:          decision,
		Resolution:        resolution,
		Reason:            reason,
		Interpretation:    "THREE_CASE_SOURCE_SEMANTIC_CAUSALITY_WITHOUT_AGGREGATION",
		SubjectSHA:        input.SubjectSHA,
		ContractID:        input.Contract.ID,
		Manifest:          input.Manifest,
		Samples:           results,
		Summary:           summary,
		TransitionSummary: summarizeCausalityTransitions(results, input.Manifest),
		UnknownFindings:   unknowns,
		NotClaimed:        append([]string(nil), input.Contract.NotClaimed...),
		FactsDigest:       digestValue(input),
	}
	return sealCausalityReport(report)
}

func evaluateCausalitySample(sample CausalitySampleInput, manifest CausalityManifest) (CausalitySampleResult, []CausalityUnknown) {
	caseContract, _ := causalityCaseContract(manifest, sample.CandidateID)
	result := CausalitySampleResult{
		CandidateID:          sample.CandidateID,
		Baseline:             makeCausalCaseResult(sample.Baseline),
		Semantic:             makeCausalCaseResult(sample.Semantic),
		NonSemantic:          makeCausalCaseResult(sample.NonSemantic),
		RequiredChangeFields: append([]string(nil), caseContract.RequiredChangeFields...),
		Status:               "DISCHARGED",
		Stage:                "CAUSALITY",
		Step:                 "preserve-intervention-boundary",
		Reason:               "SOURCE_SEMANTIC_CAUSALITY_DISCHARGED",
	}
	result.Baseline.Status = "DISCHARGED"
	result.Baseline.Stage = "BASELINE"
	result.Baseline.Step = "bind-receipt"
	result.Baseline.Reason = "BASELINE_BOUND"
	setCausalTransition(&result.Baseline, "DISCHARGED", "BASELINE", "bind-receipt", "SOURCE_OBSERVATION_BOUND")
	result.SourceSemanticValueChanged = sample.Baseline.Observation.SemanticValue != sample.Semantic.Observation.SemanticValue
	result.SourceDigestChanged = sample.Baseline.Observation.SourceDigest != sample.Semantic.Observation.SourceDigest
	result.SemanticProjectionChanged = !reflect.DeepEqual(receiptSemanticProjectionOf(sample.Baseline.Receipt), receiptSemanticProjectionOf(sample.Semantic.Receipt))
	result.DecisionChanged = sample.Baseline.Receipt.Decision != sample.Semantic.Receipt.Decision
	result.ClaimTransitionsChanged = !reflect.DeepEqual(sample.Baseline.Receipt.ClaimTransitions, sample.Semantic.Receipt.ClaimTransitions)
	result.NonSemanticSourceDigestChanged = sample.Baseline.Observation.SourceDigest != sample.NonSemantic.Observation.SourceDigest
	result.NonSemanticSemanticValuePreserved = sample.Baseline.Observation.SemanticValue == sample.NonSemantic.Observation.SemanticValue
	result.NonSemanticProjectionChanged = !reflect.DeepEqual(receiptSemanticProjectionOf(sample.Baseline.Receipt), receiptSemanticProjectionOf(sample.NonSemantic.Receipt))
	result.NonSemanticDecisionChanged = sample.Baseline.Receipt.Decision != sample.NonSemantic.Receipt.Decision
	result.ChangedFields = changedReceiptFields(sample.Baseline.Receipt, sample.Semantic.Receipt)

	unknowns := make([]CausalityUnknown, 0)
	if !result.SourceDigestChanged {
		unknowns = append(unknowns, causalityUnknown(sample.CandidateID, sample.Semantic.CaseID, "SEMANTIC_INTERVENTION", "observe-source", "SEMANTIC_DIGEST_UNCHANGED"))
	}
	if !result.SourceSemanticValueChanged {
		unknowns = append(unknowns, causalityUnknown(sample.CandidateID, sample.Semantic.CaseID, "SEMANTIC_INTERVENTION", "observe-source", "SEMANTIC_INTERVENTION_NOT_OBSERVED"))
	}
	if !result.NonSemanticSourceDigestChanged {
		unknowns = append(unknowns, causalityUnknown(sample.CandidateID, sample.NonSemantic.CaseID, "NON_SEMANTIC_INTERVENTION", "observe-source", "NON_SEMANTIC_DIGEST_UNCHANGED"))
	}
	if !result.NonSemanticSemanticValuePreserved {
		unknowns = append(unknowns, causalityUnknown(sample.CandidateID, sample.NonSemantic.CaseID, "NON_SEMANTIC_INTERVENTION", "observe-source", "NON_SEMANTIC_SEMANTIC_VALUE_CHANGED"))
	}
	if len(unknowns) > 0 {
		result.Status = "UNKNOWN"
		result.Stage = "CAUSALITY"
		result.Step = "observe-source"
		result.Reason = unknowns[0].Reason
		result.Semantic.Status = "UNKNOWN"
		result.Semantic.Stage = unknowns[0].Stage
		result.Semantic.Step = unknowns[0].Step
		result.Semantic.Reason = unknowns[0].Reason
		setCausalTransition(&result.Semantic, "OPEN", unknowns[0].Stage, unknowns[0].Step, unknowns[0].Reason)
		setCausalTransition(&result.NonSemantic, "OPEN", "NON_SEMANTIC_INTERVENTION", "observe-source", "NON_SEMANTIC_INTERVENTION_NOT_OBSERVED")
		return result, unknowns
	}
	if result.NonSemanticProjectionChanged || result.NonSemanticDecisionChanged {
		result.Status = "REFUTED"
		result.Stage = "NON_SEMANTIC_INTERVENTION"
		result.Step = "compare-receipt-projection"
		result.Reason = "NON_SEMANTIC_SEMANTIC_DRIFT"
		result.NonSemantic.Status = "REFUTED"
		result.NonSemantic.Stage = result.Stage
		result.NonSemantic.Step = result.Step
		result.NonSemantic.Reason = result.Reason
		setCausalTransition(&result.NonSemantic, "REFUTED", result.Stage, result.Step, result.Reason)
		return result, nil
	}
	result.NonSemantic.Status = "DISCHARGED"
	result.NonSemantic.Stage = "NON_SEMANTIC_INTERVENTION"
	result.NonSemantic.Step = "compare-receipt-projection"
	result.NonSemantic.Reason = "SEMANTIC_PROJECTION_PRESERVED"
	setCausalTransition(&result.NonSemantic, "DISCHARGED", result.NonSemantic.Stage, result.NonSemantic.Step, result.NonSemantic.Reason)
	if !containsAll(result.ChangedFields, result.RequiredChangeFields) {
		result.Status = "REFUTED"
		result.Stage = "SEMANTIC_INTERVENTION"
		result.Step = "compare-receipt-projection"
		result.Semantic.Status = "REFUTED"
		result.Semantic.Stage = result.Stage
		if !result.SemanticProjectionChanged {
			result.Reason = "DIGEST_ONLY_BINDING"
			result.DigestOnlyBinding = true
			result.HardcodedFixture = true
		} else {
			result.Reason = "SEMANTIC_INTERVENTION_NOT_BOUND_TO_CONTRACTED_RECEIPT_FIELD"
		}
		result.Semantic.Reason = result.Reason
		setCausalTransition(&result.Semantic, "REFUTED", result.Stage, result.Step, result.Reason)
		return result, nil
	}
	result.Semantic.Status = "DISCHARGED"
	result.Semantic.Stage = "SEMANTIC_INTERVENTION"
	result.Semantic.Step = "compare-receipt-projection"
	result.Semantic.Reason = "CONTRACTED_RECEIPT_FIELD_CHANGED"
	setCausalTransition(&result.Semantic, "DISCHARGED", result.Semantic.Stage, result.Semantic.Step, result.Semantic.Reason)
	return result, nil
}

func makeCausalCaseResult(input CausalityCaseInput) CausalCaseResult {
	return CausalCaseResult{
		CaseID:                  input.CaseID,
		Kind:                    input.Kind,
		SourcePath:              input.Observation.SourcePath,
		SourceDigest:            input.Observation.SourceDigest,
		SemanticValue:           input.Observation.SemanticValue,
		ReceiptDigest:           input.Receipt.Digest,
		Decision:                input.Receipt.Decision,
		Status:                  "OPEN",
		Stage:                   "CAUSALITY",
		Step:                    "bind-receipt",
		Reason:                  "receipt bound to observed source",
		ClaimTransitions:        []ClaimTransition{{ID: input.CaseID + "-claim", From: "OPEN", To: "OPEN", Stage: "CAUSALITY", Step: "await-evidence", Reason: "evidence not yet adjudicated"}},
		ReceiptClaimTransitions: append([]ClaimTransition{}, input.Receipt.ClaimTransitions...),
		CoordinateVector:        append([]Coordinate{}, input.Receipt.CoordinateVector...),
	}
}

func setCausalTransition(result *CausalCaseResult, to, stage, step, reason string) {
	if len(result.ClaimTransitions) == 0 {
		result.ClaimTransitions = []ClaimTransition{{ID: result.CaseID + "-claim", From: "OPEN"}}
	}
	result.ClaimTransitions[0].To = to
	result.ClaimTransitions[0].Stage = stage
	result.ClaimTransitions[0].Step = step
	result.ClaimTransitions[0].Reason = reason
}

func receiptSemanticProjectionOf(receipt Receipt) receiptSemanticProjection {
	return receiptSemanticProjection{
		CandidateID:       receipt.CandidateID,
		SourcePath:        receipt.SourcePath,
		Producer:          receipt.Producer,
		Consumer:          receipt.Consumer,
		MetaOperation:     receipt.MetaOperation,
		ProofChoice:       receipt.ProofChoice,
		SemanticValue:     receipt.SemanticValue,
		Decision:          receipt.Decision,
		ClaimTransitions:  append([]ClaimTransition{}, receipt.ClaimTransitions...),
		CoordinateVector:  append([]Coordinate{}, receipt.CoordinateVector...),
		Counterexamples:   append([]Counterexample{}, receipt.Counterexamples...),
		UnknownLocations:  append([]UnknownLocation{}, receipt.UnknownLocations...),
		ExtensionEvidence: append([]ExtensionEvidence{}, receipt.ExtensionEvidence...),
		RepositoryWrites:  receipt.RepositoryWrites,
		MutationAuthority: receipt.MutationAuthority,
	}
}

func changedReceiptFields(baseline, semantic Receipt) []string {
	changed := make([]string, 0, len(requiredCausalityReceiptFields))
	if baseline.SemanticValue != semantic.SemanticValue {
		changed = append(changed, "semantic_value")
	}
	if baseline.Decision != semantic.Decision {
		changed = append(changed, "decision")
	}
	if !reflect.DeepEqual(baseline.ClaimTransitions, semantic.ClaimTransitions) {
		changed = append(changed, "claim_transitions")
	}
	return changed
}

func containsAll(values, required []string) bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	for _, value := range required {
		if !set[value] {
			return false
		}
	}
	return true
}

func causalityUnknown(candidateID, caseID, stage, step, reason string) CausalityUnknown {
	return CausalityUnknown{CandidateID: candidateID, CaseID: caseID, Stage: stage, Step: step, Reason: reason}
}

func summarizeCausalityTransitions(results []CausalitySampleResult, manifest CausalityManifest) CausalityTransitionSummary {
	refuted, discharged, open := 0, 0, 0
	for _, sample := range results {
		for _, result := range []CausalCaseResult{sample.Baseline, sample.Semantic, sample.NonSemantic} {
			if len(result.ClaimTransitions) == 0 {
				open++
				continue
			}
			switch result.ClaimTransitions[0].To {
			case "REFUTED":
				refuted++
			case "DISCHARGED":
				discharged++
			default:
				open++
			}
		}
	}
	denominator := manifest.TransitionDenominator
	if denominator == 0 {
		denominator = ExpectedCausalTransitions
	}
	return CausalityTransitionSummary{
		FixedDenominator: denominator,
		Refuted:          TransitionBucket{Numerator: refuted, Denominator: denominator},
		Discharged:       TransitionBucket{Numerator: discharged, Denominator: denominator},
		Open:             TransitionBucket{Numerator: open, Denominator: denominator},
		Reason:           manifest.TransitionDenominatorReason,
	}
}

func closedCausality(input CausalityInput, reason string) CausalityReport {
	notClaimed := ExpectedContract().NotClaimed
	if len(input.Contract.NotClaimed) > 0 {
		notClaimed = input.Contract.NotClaimed
	}
	report := CausalityReport{
		Schema:         CausalityReportSchema,
		Decision:       "FAIL_CLOSED",
		Resolution:     "LOWER_RESOLUTION",
		Reason:         reason,
		Interpretation: "NO_CAUSALITY_COMPARISON_CLAIM",
		SubjectSHA:     input.SubjectSHA,
		ContractID:     input.Contract.ID,
		Manifest:       input.Manifest,
		Samples:        []CausalitySampleResult{},
		Summary: CausalitySummary{
			CausalCases: CausalCaseCount{Observed: 0, Total: ExpectedCandidates},
			Unknowns:    1,
		},
		TransitionSummary: CausalityTransitionSummary{
			FixedDenominator: ExpectedCausalTransitions,
			Refuted:          TransitionBucket{Denominator: ExpectedCausalTransitions},
			Discharged:       TransitionBucket{Denominator: ExpectedCausalTransitions},
			Open:             TransitionBucket{Numerator: ExpectedCausalTransitions, Denominator: ExpectedCausalTransitions},
			Reason:           causalityTransitionDenominatorReason,
		},
		UnknownFindings: []CausalityUnknown{{Stage: "CAUSALITY_VALIDATION", Step: "validate-input", Reason: reason}},
		NotClaimed:      append([]string(nil), notClaimed...),
		FactsDigest:     digestValue(input),
	}
	return sealCausalityReport(report)
}
