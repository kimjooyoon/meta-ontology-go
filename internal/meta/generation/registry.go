package generation

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func DefaultRegistry() []Binding {
	return []Binding{
		{Operation: sourcepolicy.OperationCollapseAssign, Activity: "CollapseAssignReturn", Output: "operation.collapse-assign-return", IndependenceGroupID: "expression-shape", ProofChoice: ProofRegress, Executor: "scripts/refactor-metrics", Evaluator: "scripts/refactor-metrics:check", RequiredIndicatorIDs: []string{"go.ast.single-match/v1", "go.comments.preserved/v1", "go.format.fixed-point/v1"}, ReceiptRequired: true, Priority: 10},
		{Operation: sourcepolicy.OperationSplitGo, Activity: "SplitGoDeclarations", Output: "operation.split-go-declarations", IndependenceGroupID: "source-topology", ProofChoice: ProofFoundation, Executor: "scripts/source-splitter", Evaluator: "scripts/source-splitter:check", RequiredIndicatorIDs: []string{"filesystem.atomic-replacement/v1", "go.filename.build-semantics/v1", "go.header.preserved/v1", "go.import.identity/v1", "go.initialization.order/v1", "go.package.conformance/v1"}, ReceiptRequired: true, Priority: 20},
		{Operation: sourcepolicy.OperationSplitGooo, Activity: "SplitGoooSections", Output: "operation.split-gooo-sections", IndependenceGroupID: "source-topology", ProofChoice: ProofFoundation, Executor: "bootstrap/source-repacker", Evaluator: "bootstrap/source-repacker:check", RequiredIndicatorIDs: []string{"filesystem.atomic-replacement/v1", "gooo.filename.semantics/v1", "gooo.parser-domain/v1"}, ReceiptRequired: true, Priority: 20},
		{Operation: sourcepolicy.OperationExtractFunction, Activity: "ExtractFunction", Output: "operation.extract-function", IndependenceGroupID: "source-extraction", ProofChoice: ProofFoundation, Executor: "bootstrap/function-extractor", Evaluator: ".github/workflows/repository-projection.yml", RequiredIndicatorIDs: []string{"filesystem.atomic-replacement/v1", "go.header.preserved/v1", "go.import.identity/v1", "go.package.conformance/v1", "go.format.fixed-point/v1"}, ReceiptRequired: true, Priority: 30},
	}
}

func normalizeRegistry(bindings []Binding) []Binding {
	result := append([]Binding{}, bindings...)
	for index := range result {
		result[index].RequiredIndicatorIDs = append([]string{}, result[index].RequiredIndicatorIDs...)
		sort.Strings(result[index].RequiredIndicatorIDs)
	}
	sort.Slice(result, func(i, j int) bool {
		return string(result[i].Operation) < string(result[j].Operation)
	})
	return result
}

func registryIndex(bindings []Binding) (map[sourcepolicy.Operation]Binding, bool) {
	index := make(map[sourcepolicy.Operation]Binding, len(bindings))
	for _, binding := range bindings {
		if !bindingKnown(binding) {
			return nil, false
		}
		if _, exists := index[binding.Operation]; exists {
			return nil, false
		}
		index[binding.Operation] = binding
	}
	return index, len(index) != 0
}

func bindingKnown(binding Binding) bool {
	proof := binding.ProofChoice
	return binding.Operation != "" && binding.Activity != "" && binding.Output != "" && binding.IndependenceGroupID != "" && binding.Executor != "" && binding.Evaluator != "" &&
		binding.ReceiptRequired && indicatorIDsKnown(binding.RequiredIndicatorIDs) &&
		(proof == ProofFoundation || proof == ProofCoherence || proof == ProofRegress)
}

func indicatorIDsKnown(identifiers []string) bool {
	if len(identifiers) == 0 {
		return false
	}
	for index, identifier := range identifiers {
		if identifier == "" || index > 0 && identifiers[index-1] == identifier {
			return false
		}
	}
	return true
}
