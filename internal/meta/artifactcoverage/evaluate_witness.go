package artifactcoverage

import (
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/actionability"
)

func evaluateOperations(action actionability.Report, observations ObservationDocument,
	bindings []ArtifactBinding,
) (Summary, []OperationWitness, string) {
	index := make(map[string]ArtifactBinding, len(bindings))
	for _, binding := range bindings {
		index[binding.Operation] = binding
	}
	operations := append([]actionability.OperationWitness(nil), action.Operations...)
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].Operation < operations[j].Operation
	})
	summary := Summary{RepositoryWrites: observations.RepositoryWrites}
	witnesses := make([]OperationWitness, 0, len(operations))
	selected, selectedCount := "", -1
	for _, operation := range operations {
		if !operation.Executable {
			continue
		}
		witness := observeOperation(operation, index[operation.Operation], observations)
		summary.RequiredOperations++
		if witness.ExactHead {
			summary.ExactHeadOperations++
		}
		if witness.DigestBound {
			summary.DigestBoundOperations++
		}
		if witness.ReplayBound {
			summary.ReplayBoundOperations++
		}
		if witness.Canonical {
			summary.CanonicalOperations++
		} else {
			summary.UncoveredOperations++
			if operation.IndicatorCount > selectedCount ||
				operation.IndicatorCount == selectedCount && operation.Operation < selected {
				selected, selectedCount = operation.Operation, operation.IndicatorCount
			}
		}
		if witness.ObservedArtifacts > 1 {
			summary.AmbiguousOperations++
		}
		witnesses = append(witnesses, witness)
	}
	summary.CanonicalCoverageBasisPoints = coverage(summary.CanonicalOperations, summary.RequiredOperations)
	summary.ExactHeadCoverageBasisPoints = coverage(summary.ExactHeadOperations, summary.RequiredOperations)
	summary.DigestBoundCoverageBasisPoints = coverage(summary.DigestBoundOperations, summary.RequiredOperations)
	summary.ReplayBoundCoverageBasisPoints = coverage(summary.ReplayBoundOperations, summary.RequiredOperations)
	return summary, witnesses, selected
}

func coverage(covered, total int) int {
	if total == 0 {
		return 10000
	}
	return covered * 10000 / total
}
