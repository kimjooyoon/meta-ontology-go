package verify

import (
	"fmt"
	"sort"
)

func reconstructBindings(bindings []strategyBinding) ([]resolvedBinding, map[string]bool, error) {
	result := make([]resolvedBinding, 0, len(bindings))
	referenced := make(map[string]bool)
	for _, binding := range bindings {
		operation, ok := findOperation(binding.MetaOperation)
		if !ok || binding.Status != "SATISFIED" || operation.ProofChoice != binding.Family {
			return nil, nil, fmt.Errorf("indicator %q is not independently resolvable", binding.IndicatorID)
		}
		digest, err := valueDigest(operation)
		if err != nil {
			return nil, nil, err
		}
		result = append(result, resolvedBinding{IndicatorID: binding.IndicatorID, ProofChoice: binding.Family, OperationID: operation.ID, Activity: operation.Activity, Mode: operation.Mode, EvidenceDigest: binding.EvidenceDigest, OperationDigest: digest})
		referenced[operation.ID] = true
	}
	return result, referenced, nil
}

func reconstructSteps(selection strategySelection) ([]programStep, error) {
	excluded := ""
	switch selection.MetaOperation {
	case "terminate-at-fixed-point":
		excluded = "preserve-non-promoting-terminal"
	case "preserve-non-promoting-terminal":
		excluded = "terminate-at-fixed-point"
	}
	ids := make([]string, 0, len(selection.SourceMetaOperations)+1)
	seenIDs := make(map[string]bool)
	for _, id := range selection.SourceMetaOperations {
		if id == excluded {
			continue
		}
		if seenIDs[id] {
			return nil, fmt.Errorf("selected operation %q is duplicated", id)
		}
		seenIDs[id] = true
		ids = append(ids, id)
	}
	if !seenIDs[selection.MetaOperation] {
		ids = append(ids, selection.MetaOperation)
	}
	selected := make([]operationSpec, 0, len(ids))
	for _, id := range ids {
		operation, ok := findOperation(id)
		if !ok {
			return nil, fmt.Errorf("selected operation %q is invalid", id)
		}
		selected = append(selected, operation)
	}
	sort.Slice(selected, func(left, right int) bool { return selected[left].Ordinal < selected[right].Ordinal })
	steps := make([]programStep, 0, len(selected))
	for index, operation := range selected {
		dependencies := []string{}
		if index > 0 {
			dependencies = append(dependencies, selected[index-1].ID)
		}
		digest, err := valueDigest(operation)
		if err != nil {
			return nil, err
		}
		steps = append(steps, programStep{Index: index + 1, OperationID: operation.ID, Activity: operation.Activity, Mode: operation.Mode, DependsOn: dependencies, InputEntity: operation.InputEntity, OutputEntity: operation.OutputEntity, OperationDigest: digest})
	}
	return steps, nil
}
