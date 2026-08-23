package operationconformance

import (
	"encoding/json"
	"fmt"
)

func EvaluateContract(raw []byte) (ContractReceipt, error) {
	var document contractDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return ContractReceipt{}, fmt.Errorf("FAIL_CLOSED: decode SplitGo contract: %w", err)
	}
	if field := contractDrift(document); field != "" {
		return ContractReceipt{}, fmt.Errorf("FAIL_CLOSED: SplitGo contract registry drift field=%s", field)
	}
	receipt := ContractReceipt{Decision: DecisionPass, ContractID: ContractID,
		Version: DenominatorVersion, ContractDigest: digestBytes(raw), Total: len(fixedOracleCases)}
	for index, item := range document.OracleCases {
		expected := fixedOracleCases[index]
		observed := resolveOracle(item.IndicatorID, item.Observation)
		if item.ID != expected.ID || item.IndicatorID != expected.Indicator ||
			Decision(item.Expected) != expected.Expected || observed != expected.Expected {
			return ContractReceipt{}, fmt.Errorf("FAIL_CLOSED: oracle case %d drift", index)
		}
		receipt.Matched++
		switch observed {
		case DecisionPass:
			receipt.PassCases++
		case DecisionFail:
			receipt.FailCases++
		case DecisionUnknown:
			receipt.UnknownCases++
		}
	}
	return receipt, nil
}

func contractDrift(value contractDocument) string {
	denominator := value.Denominator
	policy := value.DecisionPolicy
	checks := []struct {
		field   string
		matches bool
	}{
		{"schema", value.Schema == ContractSchema},
		{"contract_id", value.ContractID == ContractID},
		{"operation_id", value.OperationID == OperationID},
		{"denominator.version", denominator.Version == DenominatorVersion},
		{"denominator.indicator_count", denominator.IndicatorCount == 6},
		{"denominator.oracle_case_count", denominator.OracleCaseCount == 18},
		{"denominator.expected_outcomes", denominator.ExpectedOutcomes["PASS"] == 6 && denominator.ExpectedOutcomes["FAIL"] == 6 && denominator.ExpectedOutcomes["UNKNOWN"] == 6},
		{"decision_policy.unknown", policy.Unknown == "PRESERVE_UNKNOWN_AND_BLOCK"},
		{"decision_policy.pass", policy.Pass.PassCount == 6 && policy.Pass.FailCount == 0 && policy.Pass.UnknownCount == 0 && policy.Pass.Decision == "PASS"},
		{"decision_policy.otherwise", policy.Otherwise == "BLOCK"},
	}
	for _, check := range checks {
		if !check.matches {
			return check.field
		}
	}
	if field := indicatorRegistryDrift(value.Indicators); field != "" {
		return field
	}
	if len(value.OracleCases) != 18 {
		return "oracle_cases"
	}
	return ""
}

func indicatorRegistryDrift(observed []IndicatorDefinition) string {
	if len(observed) != len(fixedIndicators) {
		return "indicators.count"
	}
	for index, expected := range fixedIndicators {
		actual := observed[index]
		switch {
		case actual.ID != expected.ID:
			return fmt.Sprintf("indicators[%d].id", index)
		case actual.Role != expected.Role:
			return fmt.Sprintf("indicators[%d].role", index)
		case actual.Route != expected.Route:
			return fmt.Sprintf("indicators[%d].route", index)
		case actual.RuleID != expected.RuleID:
			return fmt.Sprintf("indicators[%d].rule_id", index)
		}
	}
	return ""
}
