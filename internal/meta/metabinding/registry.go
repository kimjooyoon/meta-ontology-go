package metabinding

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
)

func canonicalRegistry() []Binding {
	bindings := make([]Binding, 0)
	for _, operation := range metricprogram.CanonicalOperations() {
		bindings = append(bindings, Binding{
			Operation: operation.ID, Activity: operation.Activity,
			ProofChoice: normalizeProof(operation.ProofChoice), Registry: "metric-program",
		})
	}
	source := sourceBindings()
	sourceOperations := make(map[string]struct{}, len(source))
	for _, binding := range source {
		sourceOperations[binding.Operation] = struct{}{}
	}
	for _, binding := range generation.DefaultRegistry() {
		operation := string(binding.Operation)
		if _, exists := sourceOperations[operation]; exists {
			continue
		}
		bindings = append(bindings, Binding{
			Operation: operation, Activity: binding.Activity,
			ProofChoice: normalizeProof(fmt.Sprint(binding.ProofChoice)), Registry: "generation",
			Executor: binding.Executor, Evaluator: binding.Evaluator,
		})
	}
	bindings = append(bindings, source...)
	sort.Slice(bindings, func(left, right int) bool { return bindings[left].Operation < bindings[right].Operation })
	return bindings
}

func sourceBindings() []Binding {
	return []Binding{
		{Operation: "bind-indicator-meta-program", Activity: "BindIndicatorMetaProgram", ProofChoice: "coherence", Registry: "meta-binding"},
		{Operation: "exempt-project-root-readme", Activity: "BindRootREADMEExemption", ProofChoice: "foundation", Registry: "source-policy"},
		{Operation: "exempt-workflow-discovery-root", Activity: "ExemptWorkflowDiscoveryRoot", ProofChoice: "foundation", Registry: "source-policy"},
		{Operation: "inspect-wrapper", Activity: "InspectWrapper", ProofChoice: "coherence", Registry: "source-policy"},
		{Operation: "observe", Activity: "ObserveMetric", ProofChoice: "coherence", Registry: "source-policy"},
		{Operation: "partition-directory", Activity: "PartitionDirectory", ProofChoice: "foundation", Registry: "source-policy", Executor: "cmd/directory-partition-witness", Evaluator: "cmd/directory-partition-witness:check"},
		{Operation: "separate-directory-kinds", Activity: "SeparateDirectoryKinds", ProofChoice: "foundation", Registry: "source-policy", Executor: "cmd/directory-kind-witness", Evaluator: "cmd/directory-kind-witness:check"},
	}
}
func normalizeProof(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "regress" {
		return "regression"
	}
	return value
}

func registryIndex(bindings []Binding) (map[string]Binding, error) {
	index := make(map[string]Binding, len(bindings))
	for _, binding := range bindings {
		if binding.Operation == "" || binding.Activity == "" || binding.ProofChoice == "" {
			return nil, fmt.Errorf("incomplete binding for %q", binding.Operation)
		}
		if _, exists := index[binding.Operation]; exists {
			return nil, fmt.Errorf("duplicate binding for %q", binding.Operation)
		}
		index[binding.Operation] = binding
	}
	return index, nil
}
