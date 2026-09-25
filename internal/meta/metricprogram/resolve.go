package metricprogram

import "fmt"

func resolveBindings(bindings []StrategyBinding) ([]ResolvedBinding, map[string]bool, error) {
	resolved := make([]ResolvedBinding, 0, len(bindings))
	referenced := make(map[string]bool)
	for _, binding := range bindings {
		operation, ok := findOperation(binding.MetaOperation)
		if !ok {
			return nil, nil, fmt.Errorf("resolve operation %q", binding.MetaOperation)
		}
		digest, err := valueDigest(operation)
		if err != nil {
			return nil, nil, err
		}
		resolved = append(resolved, ResolvedBinding{
			IndicatorID: binding.IndicatorID, ProofChoice: binding.Family, OperationID: operation.ID,
			Activity: operation.Activity, Mode: operation.Mode, EvidenceDigest: binding.EvidenceDigest, OperationDigest: digest,
		})
		referenced[operation.ID] = true
	}
	return resolved, referenced, nil
}
