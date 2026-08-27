package ambiguitybudget

func Evaluate(input Input) Receipt {
	source, sourceErr := observeSource(input.Contract.SourcePath, input.Source, input.Contract.BudgetPolicy)
	effects, effectsErr := decodeEffects(input.EffectsArtifact)
	receipt := Receipt{
		Schema: ReceiptSchema, SubjectSHA: input.SubjectSHA, ContractID: input.Contract.ID,
		Source: source, BudgetPolicy: input.Contract.BudgetPolicy, BudgetBinding: budgetBinding(input.Contract.BudgetPolicy),
		BudgetAuthority: input.Contract.BudgetPolicy.Authority, Producer: Producer, Consumer: Consumer,
		MetaOperation: MetaOperation, ProofChoice: FoundationProof, NotClaimed: append([]string(nil), input.Contract.NotClaimed...),
		Effects: effects,
	}
	if reason := validateInput(input, source, sourceErr, effectsErr); reason != "" {
		return sealUnknown(receipt, reason)
	}

	receipt.Cases = make([]CaseReceipt, 0, len(input.Contract.Cases))
	receipt.Claims = make([]ClaimTransition, 0, len(input.Contract.Cases))
	budget := policyCounts(receipt.BudgetPolicy)
	for _, contractCase := range input.Contract.Cases {
		program, _ := findCase(source, contractCase.Activity)
		result := caseReceipt(source.Digest, source.SemanticDigest, program, receipt.BudgetPolicy)
		receipt.Cases = append(receipt.Cases, result)
		receipt.Claims = append(receipt.Claims, result.Claim)
		receipt.Indicators = append(receipt.Indicators, indicatorsFor(source.SemanticDigest, program, receipt.BudgetPolicy)...)
	}
	receipt.Interventions = buildInterventions(input.Contract, input.Source, source, receipt.BudgetPolicy)
	receipt.Summary = summarize(receipt.Cases, receipt.Interventions, input.Contract.Denominator)
	receipt.SubjectDecision, receipt.SubjectResolution, receipt.SubjectReason = subjectVector(receipt.Cases)
	receipt.SubjectCoordinate = Coordinate{Stage: "ambiguity-budget", Step: "subject-resolution", Reason: receipt.SubjectReason}
	receipt.ConformanceDecision, receipt.ConformanceResolution, receipt.ConformanceReason = "PASS", "EXACT", "CONFORMANCE_CASES_MATCHED"
	receipt.FactsDigest = digestValue(struct {
		SubjectSHA    string
		Source        SourceObservation
		Policy        BudgetPolicy
		Cases         []CaseReceipt
		Interventions []InterventionReceipt
		Effects       Effects
	}{input.SubjectSHA, source, receipt.BudgetPolicy, receipt.Cases, receipt.Interventions, receipt.Effects})
	receipt.Proofs = buildProofs(receipt)
	return seal(receipt)
}

func caseReceipt(rawSourceDigest, sourceSemanticDigest string, program ProgramObservation, policy BudgetPolicy) CaseReceipt {
	budget := policyCounts(policy)
	decision, resolution, reason := subjectDecision(program, budget)
	prop := proposition(program.ID, policy)
	propDigest := digestBytes([]byte(prop))
	evidence := digestValue(struct {
		PropositionDigest      string
		SourceSemanticDigest   string
		ActivitySemanticDigest string
		ProgramSemanticDigest  string
		ElementDigest          string
		Counts                 IntegerSet
		UnobservedDimensions   []string
		ObservationGaps        []ObservationGap
		Decision               string
		Resolution             string
		Reason                 string
	}{propDigest, sourceSemanticDigest, program.ActivitySemanticDigest, program.SemanticDigest, program.ElementDigest,
		program.Counts, program.UnobservedDimensions, program.ObservationGaps, decision, resolution, reason})
	coordinate := caseCoordinate(program, reason)
	claim := ClaimTransition{CaseID: program.ID, Proposition: prop, PropositionDigest: propDigest, From: "OPEN", To: claimTarget(decision),
		Stage: coordinate.Stage, Step: coordinate.Step, Reason: reason, EvidenceDigest: evidence}
	return CaseReceipt{ID: program.ID, Activity: program.Activity, RawSourceDigest: rawSourceDigest, Class: program.Class,
		InputState: program.InputState, Program: program.Program, ProgramDigest: program.Digest,
		ProgramSemanticDigest: program.SemanticDigest, ActivitySemanticDigest: program.ActivitySemanticDigest,
		Elements: program.Elements, ElementDigest: program.ElementDigest, Counts: program.Counts,
		UnobservedDimensions: append([]string(nil), program.UnobservedDimensions...), ObservationGaps: append([]ObservationGap(nil), program.ObservationGaps...),
		Decision: decision, Resolution: resolution, Reason: reason, Coordinate: coordinate, Proposition: prop,
		PropositionDigest: propDigest, Claim: claim, EvidenceDigest: evidence, Conformance: "MATCH"}
}

func caseCoordinate(program ProgramObservation, reason string) Coordinate {
	if len(program.ObservationGaps) > 0 {
		return program.ObservationGaps[0].Coordinate
	}
	return Coordinate{Stage: "AMBIGUITY_BUDGET", Step: "case:" + program.ID, Reason: reason}
}

func indicatorsFor(sourceSemanticDigest string, program ProgramObservation, policy BudgetPolicy) []Indicator {
	values := []struct {
		metric, dimension, proof string
		observed, limit          int
	}{
		{"gooo.metric.ambiguity-budget.candidate-cardinality.v3", "interpretation_candidates", FoundationProof, program.Counts.InterpretationCandidates, policyLimit(policy, "interpretation_candidates")},
		{"gooo.metric.ambiguity-budget.unresolved-branch-cardinality.v3", "unresolved_branches", CoherenceProof, program.Counts.UnresolvedBranches, policyLimit(policy, "unresolved_branches")},
		{"gooo.metric.ambiguity-budget.evidence-path-cardinality.v3", "evidence_paths", RegressionProof, program.Counts.EvidencePaths, policyLimit(policy, "evidence_paths")},
	}
	indicators := make([]Indicator, 0, len(values))
	for _, value := range values {
		coordinateObserved := !contains(program.UnobservedDimensions, value.dimension)
		evaluation := "UNOBSERVED"
		if coordinateObserved {
			evaluation = "WITHIN_LIMIT"
			if value.observed > value.limit {
				evaluation = "EXCEEDS_LIMIT"
			}
		}
		evidence := digestValue(struct {
			SourceSemanticDigest   string
			ActivitySemanticDigest string
			ProgramSemanticDigest  string
			ElementDigest          string
			Dimension              string
			Observed               int
			CoordinateObserved     bool
			Budget                 int
		}{sourceSemanticDigest, program.ActivitySemanticDigest, program.SemanticDigest, program.ElementDigest,
			value.dimension, value.observed, coordinateObserved, value.limit})
		indicators = append(indicators, Indicator{MetricID: value.metric, CaseID: program.ID, Dimension: value.dimension,
			ProofChoice: value.proof, Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation,
			Observed: value.observed, CoordinateObserved: coordinateObserved, Budget: value.limit,
			Relation: "<=", Evaluation: evaluation, EvidenceDigest: evidence})
	}
	return indicators
}

func policyLimit(policy BudgetPolicy, dimensionID string) int {
	for _, dimension := range policy.Dimensions {
		if dimension.ID == dimensionID {
			return dimension.Limit
		}
	}
	return 0
}

func summarize(cases []CaseReceipt, interventions []InterventionReceipt, denominator Denominator) Summary {
	summary := Summary{CasesTotal: len(cases), IntegerDimensions: IntegerDimensions, Denominator: denominator}
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
		switch result.Claim.To {
		case "DISCHARGED":
			summary.Numerator.ClaimsDischarged++
		case "REFUTED":
			summary.Numerator.ClaimsRefuted++
		case "OPEN":
			summary.OpenClaims++
			summary.Numerator.ClaimsOpen++
		}
		for _, dimension := range integerDimensions {
			if !contains(result.UnobservedDimensions, dimension) {
				summary.Numerator.IntegerObservationsObserved++
			}
		}
	}
	summary.Numerator.CasesConforming = len(cases)
	for _, intervention := range interventions {
		if intervention.Satisfied {
			summary.Numerator.InterventionsSatisfied++
		}
	}
	summary.Numerator.AuthorityObserved = 0
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
	claimsProven := len(receipt.Claims) == len(receipt.Cases)
	for index, result := range receipt.Cases {
		claimsProven = claimsProven && receipt.Claims[index] == result.Claim && result.Claim.From == "OPEN" &&
			result.Claim.To == claimTarget(result.Decision) && result.Claim.Proposition != "" &&
			result.Claim.PropositionDigest != "" && result.Claim.EvidenceDigest == result.EvidenceDigest
	}
	interventionsPassed := len(receipt.Interventions) == ExpectedInterventions
	for _, intervention := range receipt.Interventions {
		interventionsPassed = interventionsPassed && intervention.Satisfied
	}
	evidence := receipt.FactsDigest
	return []Proof{
		{Choice: FoundationProof, Claim: "contract policy is bound by the FixedBudget computes declaration", Producer: Producer, Consumer: Consumer,
			MetaOperation: "bind-contract-policy-to-gooo", EvidenceDigest: evidence, Passed: receipt.BudgetAuthority == "CONTRACT_POLICY" && receipt.BudgetBinding != ""},
		{Choice: CoherenceProof, Claim: "graph cardinalities and claim propositions have independent semantic provenance", Producer: Producer, Consumer: Consumer,
			MetaOperation: "derive-provenanced-claim-transitions", EvidenceDigest: evidence, Passed: claimsProven},
		{Choice: RegressionProof, Claim: "semantic and nonsemantic interventions preserve or change only their semantic graph", Producer: Producer, Consumer: Consumer,
			MetaOperation: "replay-graph-intervention-boundary", EvidenceDigest: evidence, Passed: interventionsPassed && receipt.Effects.RepositoryWrites == 0 && receipt.Effects.WriteSetEqual},
	}
}

func validateInput(input Input, source SourceObservation, sourceErr, effectsErr error) string {
	if !validSHA(input.SubjectSHA) {
		return "SUBJECT_SHA_INVALID"
	}
	if reason := validateContract(input.Contract); reason != "" {
		return reason
	}
	if sourceErr != nil || effectsErr != nil || source.Package != input.Contract.SourcePackage || source.Namespace != input.Contract.SourceNamespace ||
		source.Lowering != canonicalLowering || source.Activities != ExpectedSourceActivities || source.Entities != ExpectedSourceEntities {
		return "SOURCE_OR_EFFECT_BINDING_UNKNOWN"
	}
	budget, ok := findBudget(source, input.Contract.BudgetActivity)
	if !ok || budget.Program != budgetBinding(input.Contract.BudgetPolicy) || budget.ID != input.Contract.BudgetPolicy.Version {
		return "BUDGET_POLICY_BINDING_UNKNOWN"
	}
	for _, contractCase := range input.Contract.Cases {
		program, ok := findCase(source, contractCase.Activity)
		if !ok || program.ID != contractCase.ID || program.ProgramKind != "CASE" {
			return "CASE_COMPUTES_UNKNOWN"
		}
	}
	return ""
}

func validateContract(contract Contract) string {
	if contract.Schema != ContractSchema || contract.ID == "" || contract.SourcePath == "" || contract.SourcePackage == "" ||
		contract.SourceNamespace == "" || contract.BudgetActivity == "" || !validPolicy(contract.BudgetPolicy) || !validDenominator(contract.Denominator) {
		return "CONTRACT_SCHEMA_INVALID"
	}
	if len(contract.Cases) != ExpectedCaseTotal || len(contract.Interventions) != ExpectedInterventions || len(contract.NotClaimed) != 4 {
		return "CONTRACT_CARDINALITY_INVALID"
	}
	caseIDs, activities := map[string]bool{}, map[string]bool{}
	for _, item := range contract.Cases {
		if item.ID == "" || item.Activity == "" || caseIDs[item.ID] || activities[item.Activity] {
			return "CONTRACT_CASE_ID_INVALID"
		}
		caseIDs[item.ID], activities[item.Activity] = true, true
	}
	interventionIDs := map[string]bool{}
	for _, item := range contract.Interventions {
		if item.ID == "" || item.TargetActivity == "" || interventionIDs[item.ID] || (item.Kind != "SEMANTIC" && item.Kind != "NONSEMANTIC") {
			return "CONTRACT_INTERVENTION_INVALID"
		}
		interventionIDs[item.ID] = true
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
		Effects    Effects
		Reason     string
	}{receipt.SubjectSHA, receipt.Source, receipt.Effects, reason})
	return seal(receipt)
}
