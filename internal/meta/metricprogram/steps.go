package metricprogram

import (
	"fmt"
	"sort"
)

func buildSteps(plan StrategyPlan) ([]ProgramStep, error) {
	operationIDs := append([]string(nil), plan.Selection.SourceMetaOperations...)
	operationIDs = append(operationIDs, plan.Selection.MetaOperation)
	operations := make([]OperationSpec, 0, len(operationIDs))
	seen := make(map[string]bool)
	for _, id := range operationIDs {
		operation, ok := findOperation(id)
		if !ok || seen[id] {
			return nil, fmt.Errorf("selected operation %q is invalid", id)
		}
		seen[id] = true
		operations = append(operations, operation)
	}
	sort.Slice(operations, func(left, right int) bool { return operations[left].Ordinal < operations[right].Ordinal })
	steps := make([]ProgramStep, 0, len(operations))
	for index, operation := range operations {
		dependsOn := []string{}
		if index > 0 {
			dependsOn = append(dependsOn, operations[index-1].ID)
		}
		digest, err := valueDigest(operation)
		if err != nil {
			return nil, err
		}
		steps = append(steps, ProgramStep{
			Index: index + 1, OperationID: operation.ID, Activity: operation.Activity, Mode: operation.Mode,
			DependsOn: dependsOn, InputEntity: operation.InputEntity, OutputEntity: operation.OutputEntity, OperationDigest: digest,
		})
	}
	return steps, nil
}
