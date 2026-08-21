package metricprogram

import "fmt"

func Compile(strategyPayload, verificationPayload []byte) (Program, []byte, error) {
	var strategy StrategyPlan
	if err := decodeExact(strategyPayload, &strategy); err != nil {
		return Program{}, nil, fmt.Errorf("decode strategy: %w", err)
	}
	var verification StrategyVerification
	if err := decodeExact(verificationPayload, &verification); err != nil {
		return Program{}, nil, fmt.Errorf("decode strategy verification: %w", err)
	}
	if err := validateStrategy(strategy, verification); err != nil {
		return Program{}, nil, err
	}
	operations := CanonicalOperations()
	registryDigest, err := valueDigest(operations)
	if err != nil {
		return Program{}, nil, err
	}
	source := CanonicalSource()
	semantic, err := semanticDigest(source)
	if err != nil {
		return Program{}, nil, err
	}
	bindings, referenced, err := resolveBindings(strategy.Bindings)
	if err != nil {
		return Program{}, nil, err
	}
	steps, err := buildSteps(strategy)
	if err != nil {
		return Program{}, nil, err
	}
	referenced[strategy.Selection.MetaOperation] = true
	program := Program{
		Schema: ProgramSchemaVersion, Repository: strategy.Repository, SubjectSHA: strategy.SubjectSHA,
		StrategyDigest: strategy.Digest, StrategyVerificationDigest: verification.Digest, ExecutionPolicy: ProgramExecutionPolicy,
		RootPolicy: strategy.RootPolicy, RegistryDigest: registryDigest, SourcePath: ProgramSourceFilename,
		SourceDigest: bytesDigest(source), SemanticDigest: semantic, Operations: operations, Bindings: bindings, Steps: steps,
		Selection: ProgramSelection{ProofChoice: strategy.Selection.ProofChoice, Decision: strategy.Selection.Decision, MetaOperation: strategy.Selection.MetaOperation, Reason: strategy.Selection.Reason},
		Coverage: Coverage{BindingCount: len(strategy.Bindings), ResolvedBindingCount: len(bindings), RegistryOperationCount: len(operations), ReferencedOperationCount: len(referenced), SelectionOperationResolved: true, Status: "COMPLETE"},
		RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
	program.Digest, err = valueDigest(program)
	if err != nil {
		return Program{}, nil, err
	}
	return program, source, nil
}
