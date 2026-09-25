package performance

import (
	"fmt"
)

// StageSpec connects a stage implementation to its budget.
type StageSpec struct {
	Stage  Stage
	Run    Runner
	Budget Budget
}

func (s StageSpec) validate() error {
	if !s.Stage.Valid() {
		if s.Stage == "" {
			return fmt.Errorf("performance stage is empty")
		}
		return fmt.Errorf("performance stage %q is not a standard compiler stage", s.Stage)
	}
	if s.Run == nil {
		return fmt.Errorf("performance stage %q has no runner", s.Stage)
	}
	return nil
}
func validateSpecs(specs []StageSpec) error {
	seen := make(map[Stage]struct{}, len(specs))
	for _, spec := range specs {
		if err := spec.validate(); err != nil {
			return err
		}
		if _, exists := seen[spec.Stage]; exists {
			return fmt.Errorf("performance stage %q is registered more than once", spec.Stage)
		}
		seen[spec.Stage] = struct{}{}
	}
	return nil
}
func orderedSpecs(specs []StageSpec) []StageSpec {
	ordered := append([]StageSpec(nil), specs...)
	for i := 1; i < len(ordered); i++ {
		current := ordered[i]
		j := i - 1
		for j >= 0 && stageOrder(ordered[j].Stage) > stageOrder(current.Stage) {
			ordered[j+1] = ordered[j]
			j--
		}
		ordered[j+1] = current
	}
	return ordered
}
