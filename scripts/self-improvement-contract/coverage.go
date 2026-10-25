package main

func coverExecutors(
	model contractModel, registry []RegistryBinding,
) []ExecutorCoverage {
	execute := model.Activities[executeActivityID]
	inputs := make(map[string]bool, len(execute.Inputs))
	for _, input := range execute.Inputs {
		inputs[input] = true
	}
	result := make([]ExecutorCoverage, 0, len(registry))
	for _, binding := range registry {
		name := model.EntityByID[binding.ExecutorEntityID]
		coverage := ExecutorCoverage{
			Operation: binding.Operation, Executor: binding.Executor,
			EntityID: binding.ExecutorEntityID, EntityName: name,
			Covered:    name != "" && inputs[binding.ExecutorEntityID],
			ActivityID: executeActivityID, ActivityName: execute.Name,
		}
		if origin, found := execute.InputOrigins[binding.ExecutorEntityID]; coverage.Covered && found {
			coverage.InputOrigin = &origin
		}
		result = append(result, coverage)
	}
	return result
}

func completeCoverage(coverage []ExecutorCoverage) bool {
	if len(coverage) == 0 {
		return false
	}
	for _, item := range coverage {
		if !item.Covered {
			return false
		}
	}
	return true
}

func coveredCount(coverage []ExecutorCoverage) int {
	count := 0
	for _, item := range coverage {
		if item.Covered {
			count++
		}
	}
	return count
}
