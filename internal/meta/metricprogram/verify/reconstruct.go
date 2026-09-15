package verify

func reconstruct(strategy strategyPlan, verification strategyVerification, source []byte) (program, error) {
	if err := validateInputs(strategy, verification, source); err != nil {
		return program{}, err
	}
	registryDigest, err := valueDigest(operations)
	if err != nil {
		return program{}, err
	}
	semantic, err := semanticDigest(source)
	if err != nil {
		return program{}, err
	}
	bindings, referenced, err := reconstructBindings(strategy.Bindings)
	if err != nil {
		return program{}, err
	}
	steps, err := reconstructSteps(strategy.Selection)
	if err != nil {
		return program{}, err
	}
	referenced[strategy.Selection.MetaOperation] = true
	expected := program{
		Schema: programSchema, Repository: strategy.Repository, SubjectSHA: strategy.SubjectSHA,
		StrategyDigest: strategy.Digest, StrategyVerificationDigest: verification.Digest, ExecutionPolicy: "READ_ONLY_META_PROGRAM",
		RootPolicy: strategy.RootPolicy, RegistryDigest: registryDigest, SourcePath: programSourceFilename,
		SourceDigest: bytesDigest(source), SemanticDigest: semantic, Operations: append([]operationSpec(nil), operations...), Bindings: bindings, Steps: steps,
		Selection:                 programSelection{ProofChoice: strategy.Selection.ProofChoice, Decision: strategy.Selection.Decision, MetaOperation: strategy.Selection.MetaOperation, Reason: strategy.Selection.Reason},
		Coverage:                  coverage{BindingCount: len(strategy.Bindings), ResolvedBindingCount: len(bindings), RegistryOperationCount: len(operations), ReferencedOperationCount: len(referenced), SelectionOperationResolved: true, Status: "COMPLETE"},
		RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
	expected.Digest, err = valueDigest(expected)
	return expected, err
}
