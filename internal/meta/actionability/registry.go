package actionability

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func canonicalExecutors() []Executor {
	result := make([]Executor, 0, len(generation.DefaultRegistry())+2)
	for _, binding := range generation.DefaultRegistry() {
		operation := string(binding.Operation)
		result = append(result, Executor{Operation: operation, Activity: operation,
			ProofChoice: normalizeProof(string(binding.ProofChoice)), Registry: "generation",
			Executor: binding.Executor, Evaluator: binding.Evaluator})
	}
	result = append(result, Executor{Operation: "bind-indicator-meta-program",
		Activity: "BindIndicatorMetaProgram", ProofChoice: "coherence", Registry: "meta-binding",
		Executor: "bootstrap/meta-binding-witness", Evaluator: "bootstrap/meta-binding-witness:check"})
	result = append(result, Executor{Operation: "measure-integration-progress",
		Activity: "MeasureIntegrationProgress", ProofChoice: "foundation", Registry: "integration-progress",
		Executor: "cmd/integration-progress-witness", Evaluator: "cmd/integration-progress-witness:check"})
	result = append(result, Executor{Operation: "measure-language-utility",
		Activity: "MeasureLanguageUtility", ProofChoice: "foundation", Registry: "language-utility",
		Executor: "cmd/language-utility-witness", Evaluator: "cmd/language-utility-witness:check"})
	result = append(result, Executor{Operation: "partition-directory",
		Activity: "PartitionDirectory", ProofChoice: "foundation", Registry: "source-policy",
		Executor: "cmd/directory-partition-witness", Evaluator: "cmd/directory-partition-witness:check"})
	result = append(result, Executor{Operation: "separate-directory-kinds",
		Activity: "SeparateDirectoryKinds", ProofChoice: "foundation", Registry: "source-policy",
		Executor: "cmd/directory-kind-witness", Evaluator: "cmd/directory-kind-witness:check"})
	sort.Slice(result, func(left, right int) bool {
		return result[left].Operation < result[right].Operation
	})
	return result
}

func executorIndex(executors []Executor) (map[string]Executor, error) {
	index := make(map[string]Executor, len(executors))
	for _, executor := range executors {
		if executor.Operation == "" || executor.Activity == "" || executor.ProofChoice == "" ||
			executor.Registry == "" || executor.Executor == "" || executor.Evaluator == "" {
			return nil, fmt.Errorf("incomplete executor for %q", executor.Operation)
		}
		if _, exists := index[executor.Operation]; exists {
			return nil, fmt.Errorf("duplicate executor for %q", executor.Operation)
		}
		index[executor.Operation] = executor
	}
	return index, nil
}

func normalizeProof(proof string) string {
	proof = strings.ToLower(strings.TrimSpace(proof))
	if proof == "regress" {
		return "regression"
	}
	return proof
}
