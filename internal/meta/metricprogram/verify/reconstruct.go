package verify

import (
	"fmt"
	"sort"
)

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
		Selection: programSelection{ProofChoice: strategy.Selection.ProofChoice, Decision: strategy.Selection.Decision, MetaOperation: strategy.Selection.MetaOperation, Reason: strategy.Selection.Reason},
		Coverage: coverage{BindingCount: len(strategy.Bindings), ResolvedBindingCount: len(bindings), RegistryOperationCount: len(operations), ReferencedOperationCount: len(referenced), SelectionOperationResolved: true, Status: "COMPLETE"},
		RepositoryWorkspaceWrites: false, PromotionAuthorized: false,
	}
	expected.Digest, err = valueDigest(expected)
	return expected, err
}

func reconstructBindings(bindings []strategyBinding) ([]resolvedBinding, map[string]bool, error) {
	result := make([]resolvedBinding, 0, len(bindings))
	referenced := make(map[string]bool)
	for _, binding := range bindings {
		operation, ok := findOperation(binding.MetaOperation)
		if !ok || binding.Status != "SATISFIED" || operation.ProofChoice != binding.Family {
			return nil, nil, fmt.Errorf("indicator %q is not independently resolvable", binding.IndicatorID)
		}
		digest, err := valueDigest(operation)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, resolvedBinding{IndicatorID: binding.IndicatorID, ProofChoice: binding.Family, OperationID: operation.ID, Activity: operation.Activity, Mode: operation.Mode, EvidenceDigest: binding.EvidenceDigest, OperationDigest: digest})
		referenced[operation.ID] = true
	}
	return result, referenced, nil
}

func reconstructSteps(selection strategySelection) ([]programStep, error) {
	ids := append([]string(nil), selection.SourceMetaOperations...)
	ids = append(ids, selection.MetaOperation)
	selected := make([]operationSpec, 0, len(ids))
	seen := make(map[string]bool)
	for _, id := range ids {
		operation, ok := findOperation(id)
		if !ok || seen[id] {
			return nil, fmt.Errorf("selected operation %q is invalid", id)
		}
		seen[id] = true
		selected = append(selected, operation)
	}
	sort.Slice(selected, func(left, right int) bool { return selected[left].Ordinal < selected[right].Ordinal })
	steps := make([]programStep, 0, len(selected))
	for index, operation := range selected {
		dependencies := []string{}
		if index > 0 {
			dependencies = append(dependencies, selected[index-1].ID)
		}
		digest, err := valueDigest(operation)
		if err != nil {
			return nil, err
		}
		steps = append(steps, programStep{Index: index + 1, OperationID: operation.ID, Activity: operation.Activity, Mode: operation.Mode, DependsOn: dependencies, InputEntity: operation.InputEntity, OutputEntity: operation.OutputEntity, OperationDigest: digest})
	}
	return steps, nil
}
