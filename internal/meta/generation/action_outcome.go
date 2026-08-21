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
		expected := expectedActionOutcome(
			action.MetricID, action.Applicability, action.Blocking,
		)
		if action.IndicatorOutcome != expected {
			return fmt.Errorf("selected action %d indicator outcome mismatch", index)
		}
	}
	return nil
}

func validateExecutionOutcomes(steps []ExecutionStep) error {
	for index, step := range steps {
		expected := expectedActionOutcome(
			step.MetricID, step.Applicability, step.Blocking,
		)
		if step.IndicatorOutcome != expected {
			return fmt.Errorf("execution step %d indicator outcome mismatch", index)
		}
	}
	return nil
}
