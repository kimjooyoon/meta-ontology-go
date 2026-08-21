package metricstrategy

func Generate(metricsPath, ledgerPath, receiptPath, repository, subjectSHA string) (Plan, error) {
	inputs, err := loadInputs(metricsPath, ledgerPath, receiptPath, repository, subjectSHA)
	if err != nil {
		return Plan{}, err
	}
	bindings, err := buildBindings(inputs.ledger.Indicators)
	if err != nil {
		return Plan{}, err
	}
	candidates, err := buildCandidates(bindings)
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{
		Schema: PlanSchema, Repository: repository, SubjectSHA: subjectSHA, ExecutionPolicy: ExecutionPolicy,
		Input: InputEvidence{SourceIndicatorSchema: inputs.baseline.SourceIndicatorSchema, SourcePolicySchema: inputs.baseline.SourcePolicySchema, SourceMetricsDigest: inputs.baseline.SourceMetricsDigest, InterventionSchema: inputs.ledger.Schema, InterventionDigest: inputs.ledger.Digest, VerificationSchema: inputs.receipt.Schema, VerificationDigest: inputs.receipt.Digest, IndicatorCount: len(bindings), ProjectionCount: len(inputs.ledger.Projections)},
		RootPolicy: rootPolicy(inputs.ledger.Baseline.RootPolicy),
		Policy: StrategyPolicy{Schema: PolicySchema, Choices: proofChoices(), FailureRule: "FIRST_UNSATISFIED_CANONICAL_FAMILY", FixedPointRule: "REGRESSION_TERMINATES_AT_VERIFIED_ZERO_RESIDUAL"},
		Bindings: bindings, Candidates: candidates,
		Selection: choose(candidates, inputs.ledger.Projections, true),
		RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
	return sealPlan(plan)
}

func rootPolicy(value interface{ GetCountsApplicability() string }) RootPolicy {
	return RootPolicy{}
}

