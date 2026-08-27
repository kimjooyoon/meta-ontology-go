package ambiguitybudget

func Evaluate(input Input) Receipt {
	source, sourceErr := observeSource(input.Contract.SourcePath, input.Source)
	receipt := Receipt{
		Schema: ReceiptSchema, SubjectSHA: input.SubjectSHA, ContractID: input.Contract.ID,
		Source: source, Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
		ProofChoice: FoundationProof, NotClaimed: append([]string(nil), input.Contract.NotClaimed...), Effects: Effects{},
	}
	if reason := validateInput(input, source, sourceErr); reason != "" {
		return sealUnknown(receipt, reason)
	}

	budgetProgram, _ := findBudget(source, input.Contract.BudgetActivity)
	receipt.Budget = budgetProgram.Counts
	receipt.Cases = make([]CaseReceipt, 0, len(input.Contract.Cases))
	receipt.Claims = make([]ClaimTransition, 0, len(input.Contract.Cases))
	for _, contractCase := range input.Contract.Cases {
		program, _ := findCase(source, contractCase.Activity)
		result := caseReceipt(source.Digest, program, receipt.Budget)
		receipt.Cases = append(receipt.Cases, result)
		receipt.Claims = append(receipt.Claims, result.Claim)
		receipt.Indicators = append(receipt.Indicators, indicatorsFor(source.Digest, program, receipt.Budget)...)
	}
	receipt.Interventions = buildInterventions(input.Contract, input.Source, source, receipt.Budget)
	receipt.Summary = summarize(receipt.Cases, receipt.Interventions)
	receipt.SubjectDecision, receipt.SubjectResolution, receipt.SubjectReason = subjectVector(receipt.Cases)
	receipt.SubjectCoordinate = Coordinate{Stage: "ambiguity-budget", Step: "subject-resolution", Reason: receipt.SubjectReason}
	receipt.ConformanceDecision, receipt.ConformanceResolution, receipt.ConformanceReason = "PASS", "EXACT", "CONFORMANCE_CASES_MATCHED"
	receipt.FactsDigest = digestValue(struct {
		SubjectSHA    string
		Source        SourceObservation
		Budget        IntegerSet
		Cases         []CaseReceipt
		Interventions []InterventionReceipt
	}{input.SubjectSHA, source, receipt.Budget, receipt.Cases, receipt.Interventions})
	receipt.Proofs = buildProofs(receipt)
	return seal(receipt)
}

func caseReceipt(sourceDigest string, program ProgramObservation, budget IntegerSet) CaseReceipt {
	parsed := computesProgram{Activity: program.Activity, Text: program.Program, Kind: program.ProgramKind,
		ID: program.ID, Class: program.Class, InputState: program.InputState, Counts: program.Counts}
	decision, resolution, reason, claimTo := subjectDecision(parsed, budget)
	evidence := digestValue(struct {
		SourceDigest string
		Activity     string
		Program      string
		Counts       IntegerSet
	}{sourceDigest, program.Activity, program.Program, program.Counts})
	coordinate := Coordinate{Stage: "ambiguity-budget", Step: "case:" + program.ID, Reason: reason}
	claim := ClaimTransition{CaseID: program.ID, From: "AMBIGUITY_OBSERVED", To: claimTo,
		Stage: coordinate.Stage, Step: coordinate.Step, Reason: reason, EvidenceDigest: evidence}
	return CaseReceipt{ID: program.ID, Activity: program.Activity, Class: program.Class, InputState: program.InputState,
		Program: program.Program, ProgramDigest: program.Digest, Counts: program.Counts, Decision: decision,
		Resolution: resolution, Reason: reason, Coordinate: coordinate, Claim: claim, EvidenceDigest: evidence, Conformance: "MATCH"}
}

func indicatorsFor(sourceDigest string, program ProgramObservation, budget IntegerSet) []Indicator {
	evaluation := "WITHIN_LIMIT"
	if program.InputState == "UNKNOWN" {
		evaluation = "UNKNOWN_INPUT"
	} else if exceeds(program.Counts, budget) {
		evaluation = "EXCEEDS_LIMIT"
	}
	values := []struct {
		metric, dimension, proof string
		observed, limit          int
	}{
		{"gooo.metric.ambiguity-budget.candidate-count.v2", "interpretation_candidates", FoundationProof, program.Counts.InterpretationCandidates, budget.InterpretationCandidates},
		{"gooo.metric.ambiguity-budget.unresolved-branches.v2", "unresolved_branches", CoherenceProof, program.Counts.UnresolvedBranches, budget.UnresolvedBranches},
		{"gooo.metric.ambiguity-budget.evidence-paths.v2", "evidence_paths", RegressionProof, program.Counts.EvidencePaths, budget.EvidencePaths},
	}
	indicators := make([]Indicator, 0, len(values))
	for _, value := range values {
		evidence := digestValue(struct {
			SourceDigest string
			Activity     string
			Dimension    string
			Observed     int
			Budget       int
		}{sourceDigest, program.Activity, value.dimension, value.observed, value.limit})
		indicators = append(indicators, Indicator{MetricID: value.metric, CaseID: program.ID, Dimension: value.dimension,
			ProofChoice: value.proof, Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
			Observed: value.observed, Budget: value.limit, Relation: "<=", Evaluation: evaluation, EvidenceDigest: evidence})
	}
	return indicators
}

func summarize(cases []CaseReceipt, interventions []InterventionReceipt) Summary {
	summary := Summary{CasesTotal: len(cases), IntegerDimensions: IntegerDimensions,
		InterventionsTotal: len(interventions), FixedDenominator: FixedDenominator}
	for _, result := range cases {
		if result.InputState == "KNOWN" {
			summary.KnownCases++
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
		if result.Claim.To == "OPEN" {
			summary.OpenClaims++
		}
	}
	return summary
}

func subjectVector(cases []CaseReceipt) (string, string, string) {
	for _, result := range cases {
		if result.Resolution == "LOWER_RESOLUTION" {
			return "MIXED", "LOWER_RESOLUTION", "AMBIGUITY_CASE_VECTOR_CONTAINS_LOWER_RESOLUTION"
		}
	}
	return "EXACT", "EXACT", "AMBIGUITY_CASE_VECTOR_EXACT"
}

func buildProofs(receipt Receipt) []Proof {
	claimsPreserved := len(receipt.Claims) == len(receipt.Cases)
	for index, result := range receipt.Cases {
		claimsPreserved = claimsPreserved && receipt.Claims[index] == result.Claim && result.Claim.EvidenceDigest != ""
	}
	interventionsPassed := len(receipt.Interventions) == ExpectedInterventions
	for _, intervention := range receipt.Interventions {
		interventionsPassed = interventionsPassed && intervention.Satisfied
	}
	evidence := receipt.FactsDigest
	return []Proof{
		{Choice: FoundationProof, Claim: "budget and cases are observed from lowered computes declarations", Producer: Producer, Consumer: Consumer, MetaOperation: "observe-gooo-computes", EvidenceDigest: evidence, Passed: receipt.Source.Lowering == canonicalLowering && receipt.Budget == expectedBudget()},
		{Choice: CoherenceProof, Claim: "every subject transition has stage, step, reason, and evidence digest", Producer: Producer, Consumer: Consumer, MetaOperation: "preserve-provenanced-claim-transitions", EvidenceDigest: evidence, Passed: claimsPreserved},
		{Choice: RegressionProof, Claim: "semantic and nonsemantic interventions are distinguishable without effects", Producer: Producer, Consumer: Consumer, MetaOperation: "replay-intervention-boundary", EvidenceDigest: evidence, Passed: interventionsPassed && receipt.Effects == (Effects{})},
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
		source.Lowering != canonicalLowering || source.Activities != ExpectedCaseTotal+1 {
		return "SOURCE_BINDING_UNKNOWN"
	}
	budget, ok := findBudget(source, input.Contract.BudgetActivity)
	if !ok || budget.Counts != expectedBudget() {
		return "BUDGET_COMPUTES_UNKNOWN"
	}
	for _, contractCase := range input.Contract.Cases {
		program, ok := findCase(source, contractCase.Activity)
		if !ok || program.ID != contractCase.ID || program.ProgramKind != "CASE" {
			return "CASE_COMPUTES_UNKNOWN"
		}
	}
	return ""
}

func seal(receipt Receipt) Receipt {
	receipt.Digest = receiptDigest(receipt)
	return receipt
}

func sealUnknown(receipt Receipt, reason string) Receipt {
	receipt.ConformanceDecision = "FAIL_CLOSED"
	receipt.ConformanceResolution = "LOWER_RESOLUTION"
	receipt.ConformanceReason = reason
	receipt.SubjectDecision = "UNKNOWN"
	receipt.SubjectResolution = "LOWER_RESOLUTION"
	receipt.SubjectReason = reason
	receipt.SubjectCoordinate = Coordinate{Stage: "ambiguity-budget", Step: "observe-source", Reason: reason}
	receipt.FactsDigest = digestValue(struct {
		SubjectSHA string
		Source     SourceObservation
		Reason     string
	}{receipt.SubjectSHA, receipt.Source, reason})
	return seal(receipt)
}
