package generation

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func DefaultRegistry() []Binding {
	return []Binding{
		{Operation: sourcepolicy.OperationCollapseAssign, IndependenceGroupID: "expression-shape", ProofChoice: ProofRegress, Executor: "scripts/refactor-metrics", Priority: 10},
		{Operation: sourcepolicy.OperationSplitGo, IndependenceGroupID: "source-topology", ProofChoice: ProofFoundation, Executor: "scripts/source-splitter", Priority: 20},
		{Operation: sourcepolicy.OperationSplitGooo, IndependenceGroupID: "source-topology", ProofChoice: ProofFoundation, Executor: "bootstrap/source-repacker", Priority: 20},
	}
}

func normalizeRegistry(bindings []Binding) []Binding {
	result := append([]Binding{}, bindings...)
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
	return binding.Operation != "" && binding.IndependenceGroupID != "" && binding.Executor != "" &&
		(proof == ProofFoundation || proof == ProofCoherence || proof == ProofRegress)
}
