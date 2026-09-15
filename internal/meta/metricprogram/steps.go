package metricprogram

import (
	"fmt"
	"sort"
)

func buildSteps(plan StrategyPlan) ([]ProgramStep, error) {
	operationIDs, err := operationIDsForSelection(plan.Selection)
	if err != nil {
		return nil, err
	}
	operations := make([]OperationSpec, 0, len(operationIDs))
	for _, id := range operationIDs {
		operation, ok := findOperation(id)
		if !ok {
			return nil, fmt.Errorf("selected operation %q is invalid", id)
		}
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

func operationIDsForSelection(selection StrategySelection) ([]string, error) {
	excluded := ""
	switch selection.MetaOperation {
	case "terminate-at-fixed-point":
		excluded = "preserve-non-promoting-terminal"
	case "preserve-non-promoting-terminal":
		excluded = "terminate-at-fixed-point"
	}
	ids := make([]string, 0, len(selection.SourceMetaOperations)+1)
	seen := make(map[string]bool)
	for _, id := range selection.SourceMetaOperations {
		if id == excluded {
			continue
		}
		if seen[id] {
			return nil, fmt.Errorf("selected operation %q is duplicated", id)
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if !seen[selection.MetaOperation] {
		ids = append(ids, selection.MetaOperation)
	}
	return ids, nil
}
