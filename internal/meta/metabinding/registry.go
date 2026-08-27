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
	for _, binding := range generation.DefaultRegistry() {
		operation := string(binding.Operation)
		activity := map[string]string{
			"collapse-assign-return": "CollapseAssignReturn",
			"split-go-declarations":  "SplitGoDeclarations",
			"split-gooo-sections":    "SplitGoooSections",
		}[operation]
		if activity == "" {
			activity = operation
		}
		bindings = append(bindings, Binding{
			Operation: operation, Activity: activity,
			ProofChoice: normalizeProof(fmt.Sprint(binding.ProofChoice)), Registry: "generation",
			Executor: binding.Executor, Evaluator: binding.Evaluator,
		})
	}
	bindings = append(bindings, sourceBindings()...)
	sort.Slice(bindings, func(left, right int) bool { return bindings[left].Operation < bindings[right].Operation })
	return bindings
}

func sourceBindings() []Binding {
	return []Binding{
		{Operation: "bind-indicator-meta-program", Activity: "BindIndicatorMetaProgram", ProofChoice: "coherence", Registry: "meta-binding"},
		{Operation: "exempt-project-root-readme", Activity: "BindRootREADMEExemption", ProofChoice: "foundation", Registry: "source-policy"},
		{Operation: "extract-function", Activity: "ExtractFunction", ProofChoice: "foundation", Registry: "repository-projection", Executor: "bootstrap/function-extractor", Evaluator: ".github/workflows/repository-projection.yml"},
		{Operation: "inspect-wrapper", Activity: "InspectWrapper", ProofChoice: "coherence", Registry: "source-policy"},
		{Operation: "measure-integration-progress", Activity: "MeasureIntegrationProgress", ProofChoice: "foundation", Registry: "integration-progress", Executor: "cmd/integration-progress-witness", Evaluator: ".github/workflows/integration-progress-evidence.yml"},
		{Operation: "measure-language-utility", Activity: "MeasureLanguageUtility", ProofChoice: "foundation", Registry: "language-utility", Executor: "cmd/language-utility-witness", Evaluator: ".github/workflows/language-utility-evidence.yml"},
		{Operation: "observe", Activity: "ObserveMetric", ProofChoice: "coherence", Registry: "source-policy"},
		{Operation: "partition-directory", Activity: "PartitionDirectory", ProofChoice: "foundation", Registry: "source-policy", Executor: "cmd/directory-partition-witness", Evaluator: "cmd/directory-partition-witness:check"},
		{Operation: "preserve-workflow-discovery", Activity: "PreserveWorkflowDiscovery", ProofChoice: "foundation", Registry: "source-policy", Executor: "scripts/line-metrics", Evaluator: ".github/workflows/repository-projection.yml"},
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
