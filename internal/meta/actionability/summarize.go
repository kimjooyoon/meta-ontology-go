package actionability

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metabinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func summarize(in input, executors map[string]Executor) (Summary, []OperationWitness, string) {
	summary := Summary{UnboundIndicators: in.binding.Summary.UnboundIndicators}
	counts := make(map[string]int)
	for _, indicator := range in.metrics.Meta.Indicators {
		if indicator.Applicability == sourcepolicy.ApplicabilityNotApplicable {
			summary.NonApplicableIndicators++
			if indicator.Subject == "." {
				summary.ProjectRootExemptions++
			}
			continue
		}
		if indicator.Blocking {
			summary.ApplicableBlockingIndicators++
			counts[string(indicator.Operation)]++
		}
	}
	bindings := make(map[string]metabinding.Witness, len(in.binding.Witnesses))
	for _, witness := range in.binding.Witnesses {
		bindings[witness.Operation] = witness
	}
	keys := make([]string, 0, len(counts))
	for operation := range counts {
		keys = append(keys, operation)
	}
	sort.Strings(keys)
	operations := make([]OperationWitness, 0, len(keys))
	selected, selectedCount := "", -1
	for _, operation := range keys {
		binding, bound := bindings[operation]
		executor, registered := executors[operation]
		metaBound := bound && binding.Bound
		executable := metaBound && registered &&
			normalizeProof(binding.ProofChoice) == executor.ProofChoice
		status := "EXECUTABLE"
		switch {
		case !metaBound:
			status = "META_BINDING_MISSING"
		case !registered:
			status = "EXECUTOR_MISSING"
		case normalizeProof(binding.ProofChoice) != executor.ProofChoice:
			status = "PROOF_CHOICE_MISMATCH"
		}
		count := counts[operation]
		if executable {
			summary.ExecutableOperations++
			summary.ActionableIndicators += count
		} else if count > selectedCount || count == selectedCount && operation < selected {
			selected, selectedCount = operation, count
		}
		operations = append(operations, OperationWitness{Operation: operation,
			IndicatorCount: count, ProofChoice: binding.ProofChoice,
			BindingRegistry: binding.Registry, MetaBound: metaBound, Executable: executable,
			ExecutorRegistry: executor.Registry, Executor: executor.Executor,
			Evaluator: executor.Evaluator, Status: status})
	}
	summary.RequiredOperations = len(keys)
	summary.MissingOperations = summary.RequiredOperations - summary.ExecutableOperations
	summary.UnactionableIndicators = summary.ApplicableBlockingIndicators - summary.ActionableIndicators
	summary.IndicatorCoverageBasisPoints = coverage(summary.ActionableIndicators, summary.ApplicableBlockingIndicators)
	summary.OperationCoverageBasisPoints = coverage(summary.ExecutableOperations, summary.RequiredOperations)
	return summary, operations, selected
}
