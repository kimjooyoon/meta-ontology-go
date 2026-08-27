package ambiguitybudget

import "strings"

func Evaluate(input Input) Receipt {
	source, sourceErr := observeSource(input.Contract.SourcePath, input.Source)
	receipt := Receipt{
		Schema: ReceiptSchema, SubjectSHA: input.SubjectSHA, ContractID: input.Contract.ID,
		Source: source, Budget: input.Contract.Budget, Producer: Producer, Consumer: Consumer,
		MetaOperation: MetaOperation, ProofChoice: FoundationProof,
		NotClaimed: append([]string(nil), input.Contract.NotClaimed...),
		Effects:    Effects{},
	}
	if reason := validateInput(input, source, sourceErr); reason != "" {
		return sealUnknown(receipt, "AMBIGUITY_INPUT_UNKNOWN", Coordinate{
			Stage: "ambiguity-budget", Step: "observe-source", Reason: reason,
		}, reason)
	}

	facts := digestValue(struct {
		SubjectSHA string
		Contract   Contract
		Source     SourceObservation
	}{input.SubjectSHA, input.Contract, source})
	receipt.FactsDigest = facts
	for _, spec := range input.Contract.Cases {
		result := evaluateCase(spec, input.Contract.Budget)
		receipt.Cases = append(receipt.Cases, result)
		receipt.Claims = append(receipt.Claims, result.Claim)
		receipt.Indicators = append(receipt.Indicators, indicatorsFor(spec, result)...)
	}
	receipt.Summary = summarize(receipt.Cases, receipt.Indicators, input.Contract.FixedDenominator)
	receipt.Proofs = buildProofs(receipt, facts)
	receipt.Decision, receipt.Resolution, receipt.Reason = "PASS", "EXACT", "AMBIGUITY_BUDGET_CONTRACT_SATISFIED"
	receipt.Coordinate = Coordinate{Stage: "ambiguity-budget", Step: "seal-receipt", Reason: receipt.Reason}
	for _, result := range receipt.Cases {
		if result.Status != "SATISFIED" {
			receipt.Decision = "FAIL_CLOSED"
			receipt.Resolution = "LOWER_RESOLUTION"
			receipt.Reason = "AMBIGUITY_BUDGET_CONTRACT_NOT_SATISFIED"
			receipt.Coordinate.Reason = receipt.Reason
			break
		}
	}
	return seal(receipt)
}

func evaluateCase(spec CaseSpec, budget IntegerSet) CaseReceipt {
	result := CaseReceipt{
		ID: spec.ID, Class: spec.Class, InputState: spec.InputState, Counts: spec.Counts,
		ExpectedDecision: spec.ExpectedDecision, ExpectedResolution: spec.ExpectedResolution,
		ExpectedReason: spec.ExpectedReason, Coordinate: spec.Coordinate, Claim: spec.Claim,
		Decision: "UNKNOWN", Resolution: "LOWER_RESOLUTION", Reason: "AMBIGUITY_COUNT_UNKNOWN",
	}
	switch {
	case spec.InputState == "UNKNOWN":
		result.Decision = "UNKNOWN"
		result.Reason = "AMBIGUITY_INPUT_UNKNOWN"
	case !validCounts(spec.Counts):
		result.Reason = "AMBIGUITY_COUNT_UNKNOWN"
	case exceeds(spec.Counts, budget):
		result.Decision = "FAIL_CLOSED"
		result.Reason = "AMBIGUITY_BUDGET_EXCEEDED"
	default:
		result.Decision, result.Resolution, result.Reason = "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"
	}
	result.Status = "SATISFIED"
	if result.Decision != spec.ExpectedDecision || result.Resolution != spec.ExpectedResolution || result.Reason != spec.ExpectedReason {
		result.Status = "NOT_SATISFIED"
	}
	return result
}

func validCounts(counts IntegerSet) bool {
	return counts.InterpretationCandidates >= 1 && counts.UnresolvedBranches >= 0 && counts.EvidencePaths >= 1
}

func exceeds(value, budget IntegerSet) bool {
	return value.InterpretationCandidates > budget.InterpretationCandidates ||
		value.UnresolvedBranches > budget.UnresolvedBranches || value.EvidencePaths > budget.EvidencePaths
}

func indicatorsFor(spec CaseSpec, result CaseReceipt) []Indicator {
	values := []struct {
		id, dimension, class, proof string
		observed, expected          int
	}{
		{"gooo.metric.ambiguity-budget.candidate-count.v1", "interpretation_candidates", "DRIVER", FoundationProof, result.Counts.InterpretationCandidates, spec.Counts.InterpretationCandidates},
		{"gooo.metric.ambiguity-budget.unresolved-branches.v1", "unresolved_branches", "DRIVER", CoherenceProof, result.Counts.UnresolvedBranches, spec.Counts.UnresolvedBranches},
		{"gooo.metric.ambiguity-budget.evidence-paths.v1", "evidence_paths", "DRIVER", RegressionProof, result.Counts.EvidencePaths, spec.Counts.EvidencePaths},
	}
	indicators := make([]Indicator, len(values))
	for index, value := range values {
		indicators[index] = Indicator{MetricID: value.id, CaseID: spec.ID, Dimension: value.dimension,
			Class: value.class, ProofChoice: value.proof, Producer: Producer, Consumer: Consumer,
			MetaOperation: MetaOperation, Observed: value.observed, Expected: value.expected,
			Satisfied: value.observed == value.expected}
	}
	return indicators
}

func summarize(cases []CaseReceipt, indicators []Indicator, denominator int) Summary {
	summary := Summary{CasesTotal: len(cases), CoordinatesTotal: len(indicators), FixedDenominator: denominator}
	for _, result := range cases {
		if result.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		switch result.Class {
		case "ZERO":
			summary.ZeroAmbiguityCases++
		case "BOUNDARY":
			summary.BoundaryCases++
		case "OVER":
			summary.OverBudgetCases++
		case "UNKNOWN":
			summary.UnknownCases++
		}
		if result.Resolution == "LOWER_RESOLUTION" {
			summary.LowerResolutionCases++
		}
	}
	for _, indicator := range indicators {
		if indicator.Satisfied {
			summary.CoordinatesSatisfied++
		}
	}
	return summary
}

func buildProofs(receipt Receipt, evidence string) []Proof {
	allIndicators := receipt.Summary.CoordinatesSatisfied == receipt.Summary.CoordinatesTotal
	transitionsPreserved := len(receipt.Claims) == len(receipt.Cases)
	for index, result := range receipt.Cases {
		transitionsPreserved = transitionsPreserved && result.Claim == receipt.Claims[index]
	}
	return []Proof{
		{Choice: FoundationProof, Claim: "source and integer ambiguity coordinates are observed", Producer: Producer, Consumer: Consumer, MetaOperation: "observe-ambiguity-coordinates", EvidenceDigest: evidence, Passed: allIndicators},
		{Choice: CoherenceProof, Claim: "budget decisions retain their claim transitions", Producer: Producer, Consumer: Consumer, MetaOperation: "preserve-claim-transitions", EvidenceDigest: evidence, Passed: transitionsPreserved},
		{Choice: RegressionProof, Claim: "over-budget and unknown inputs lower resolution without writes", Producer: Producer, Consumer: Consumer, MetaOperation: "lower-resolution-on-budget-overflow", EvidenceDigest: evidence, Passed: receipt.Summary.LowerResolutionCases == 2 && receipt.Effects.RepositoryWrites == 0 && !receipt.Effects.MutationAuthority},
	}
}

func validateInput(input Input, source SourceObservation, sourceErr error) string {
	if !validSHA(input.SubjectSHA) {
		return "SUBJECT_SHA_INVALID"
	}
	if reason := validateContract(input.Contract); reason != "" {
		return reason
	}
	if sourceErr != nil || source.Package != input.Contract.SourcePackage || source.Namespace != input.Contract.SourceNamespace ||
		source.Entities != input.Contract.SourceEntities || source.Activities != input.Contract.SourceActivities {
		return "SOURCE_BINDING_UNKNOWN"
	}
	return ""
}

func validSHA(value string) bool {
	return len(value) == 40 && strings.Trim(value, "0123456789abcdefABCDEF") == ""
}

func seal(receipt Receipt) Receipt {
	receipt.Digest = receiptDigest(receipt)
	return receipt
}

func sealUnknown(receipt Receipt, reason string, coordinate Coordinate, detail string) Receipt {
	receipt.Decision, receipt.Resolution, receipt.Reason = "UNKNOWN", "LOWER_RESOLUTION", reason
	receipt.Coordinate = coordinate
	receipt.FactsDigest = digestValue(struct {
		SubjectSHA string
		Contract   Contract
		Source     SourceObservation
		Detail     string
	}{receipt.SubjectSHA, Contract{ID: receipt.ContractID}, receipt.Source, detail})
	receipt.Summary = Summary{CasesTotal: ExpectedCaseTotal, CoordinatesTotal: FixedDenominator, FixedDenominator: FixedDenominator}
	receipt.Proofs = []Proof{{Choice: FoundationProof, Claim: "input identity remains explicit while resolution is lowered", Producer: Producer, Consumer: Consumer, MetaOperation: "record-unknown-input", EvidenceDigest: receipt.FactsDigest, Passed: true}}
	return seal(receipt)
}
