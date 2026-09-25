package operationconformance

import "fmt"

func indicatorRegistryDrift(observed []IndicatorDefinition) string {
	if len(observed) != len(fixedIndicators) {
		return fmt.Sprintf("indicators.count got=%d want=%d", len(observed), len(fixedIndicators))
	}
	for index, expected := range fixedIndicators {
		actual := observed[index]
		switch {
		case actual.ID != expected.ID:
			return fmt.Sprintf("indicators[%d].id got=%q want=%q", index, actual.ID, expected.ID)
		case actual.Role != expected.Role:
			return fmt.Sprintf("indicators[%d].role got=%q want=%q", index, actual.Role, expected.Role)
		case actual.Route != expected.Route:
			return fmt.Sprintf("indicators[%d].route got=%q want=%q", index, actual.Route, expected.Route)
		case actual.RuleID != expected.RuleID:
			return fmt.Sprintf("indicators[%d].rule_id got=%q want=%q", index, actual.RuleID, expected.RuleID)
		}
	}
	return ""
}
