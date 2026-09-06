package generation

import (
	"reflect"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func legacyDefaultRegistry() []Binding {
	registry, err := defaultRegistryWithError()
	if err != nil {
		return nil
	}
	return registry
}

func defaultRegistryWithError() ([]Binding, error) {
	contract, err := loadOperationInputContract()
	if err != nil {
		return nil, err
	}
	return []Binding{
		{Operation: sourcepolicy.OperationCollapseAssign, Activity: "CollapseAssignReturn", Output: "operation.collapse-assign-return", InputSubjectKind: contract.Bindings[sourcepolicy.OperationCollapseAssign].InputSubjectKind, InputContractSourceDigest: contract.SourceDigest, InputContractSemanticDigest: contract.SemanticDigest, IndependenceGroupID: "expression-shape", ProofChoice: ProofRegress, Executor: "scripts/refactor-metrics", Evaluator: "scripts/refactor-metrics:check", RequiredIndicatorIDs: []string{"go.ast.single-match/v1", "go.comments.preserved/v1", "go.format.fixed-point/v1"}, ReceiptRequired: true, Priority: 10},
		{Operation: sourcepolicy.OperationSplitGo, Activity: "SplitGoDeclarations", Output: "operation.split-go-declarations", InputSubjectKind: contract.Bindings[sourcepolicy.OperationSplitGo].InputSubjectKind, InputContractSourceDigest: contract.SourceDigest, InputContractSemanticDigest: contract.SemanticDigest, IndependenceGroupID: "source-topology", ProofChoice: ProofFoundation, Executor: "scripts/source-splitter", Evaluator: "scripts/source-splitter:check", RequiredIndicatorIDs: []string{"filesystem.atomic-replacement/v1", "go.filename.build-semantics/v1", "go.header.preserved/v1", "go.import.identity/v1", "go.initialization.order/v1", "go.package.conformance/v1"}, ReceiptRequired: true, Priority: 20},
		{Operation: sourcepolicy.OperationSplitGooo, Activity: "SplitGoooSections", Output: "operation.split-gooo-sections", InputSubjectKind: contract.Bindings[sourcepolicy.OperationSplitGooo].InputSubjectKind, InputContractSourceDigest: contract.SourceDigest, InputContractSemanticDigest: contract.SemanticDigest, IndependenceGroupID: "source-topology", ProofChoice: ProofFoundation, Executor: "bootstrap/source-repacker", Evaluator: "bootstrap/source-repacker:check", RequiredIndicatorIDs: []string{"filesystem.atomic-replacement/v1", "gooo.filename.semantics/v1", "gooo.parser-domain/v1"}, ReceiptRequired: true, Priority: 20},
		{Operation: sourcepolicy.OperationExtractFunction, Activity: "ExtractFunction", Output: "operation.extract-function", InputSubjectKind: contract.Bindings[sourcepolicy.OperationExtractFunction].InputSubjectKind, InputContractSourceDigest: contract.SourceDigest, InputContractSemanticDigest: contract.SemanticDigest, IndependenceGroupID: "source-extraction", ProofChoice: ProofFoundation, Executor: "bootstrap/function-extractor", Evaluator: ".github/workflows/repository-projection.yml", RequiredIndicatorIDs: []string{"filesystem.atomic-replacement/v1", "go.header.preserved/v1", "go.import.identity/v1", "go.package.conformance/v1", "go.format.fixed-point/v1"}, ReceiptRequired: true, Priority: 30},
	}, nil
}

// legacyBindingForOperation resolves one operation from the plan's registry. The
// registry is the authority shared by planning and execution consumers.
func legacyBindingForOperation(bindings []Binding, operation sourcepolicy.Operation) (Binding, bool) {
	var result Binding
	found := false
	for _, binding := range bindings {
		if binding.Operation != operation {
			continue
		}
		if found {
			return Binding{}, false
		}
		result, found = binding, true
	}
	if !found || !bindingKnown(result) || !canonicalBinding(result) {
		return Binding{}, false
	}
	return result, true
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

func legacyRegistryIndex(bindings []Binding) (map[sourcepolicy.Operation]Binding, bool) {
	if !canonicalRegistry(bindings) {
		return nil, false
	}
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
		binding.InputSubjectKind != "" && validDigest(binding.InputContractSourceDigest) && validDigest(binding.InputContractSemanticDigest) &&
		binding.ReceiptRequired && indicatorIDsKnown(binding.RequiredIndicatorIDs) &&
		(proof == ProofFoundation || proof == ProofCoherence || proof == ProofRegress)
}

func canonicalBinding(binding Binding) bool {
	expected, err := defaultRegistryWithError()
	if err != nil {
		return false
	}
	for _, candidate := range expected {
		if candidate.Operation == binding.Operation {
			return reflect.DeepEqual(normalizeRegistry([]Binding{candidate}), normalizeRegistry([]Binding{binding}))
		}
	}
	return false
}

func canonicalRegistry(bindings []Binding) bool {
	expected, err := defaultRegistryWithError()
	if err != nil {
		return false
	}
	return reflect.DeepEqual(normalizeRegistry(bindings), normalizeRegistry(expected))
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
