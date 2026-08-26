package generation

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func expectedActionOutcome(
	metricID sourcepolicy.Dimension,
	applicability sourcepolicy.Applicability,
	blocking bool,
) sourcepolicy.IndicatorOutcome {
	return sourcepolicy.Indicator{
		MetricID:      metricID,
		Applicability: applicability,
		Blocking:      blocking,
		Satisfied:     false,
	}.Outcome()
}

func validateActionOutcomes(actions []Action) error {
	for index, action := range actions {
		if !actionMatchesSourceIndicator(action) {
			return fmt.Errorf("selected action %d indicator membership mismatch", index)
		}
		expected := action.SourceIndicator.Outcome()
		if action.IndicatorOutcome != expected {
			return fmt.Errorf("selected action %d indicator outcome mismatch", index)
		}
	}
	return nil
}

func validateExecutionOutcomes(steps []ExecutionStep) error {
	for index, step := range steps {
		if !stepMatchesSourceIndicator(step) {
			return fmt.Errorf("execution step %d indicator membership mismatch", index)
		}
		expected := step.SourceIndicator.Outcome()
		if step.IndicatorOutcome != expected {
			return fmt.Errorf("execution step %d indicator outcome mismatch", index)
		}
	}
	return nil
}
