package operationconformance

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func EvaluateContract(raw []byte) (ContractReceipt, error) {
	var document contractDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return ContractReceipt{}, fmt.Errorf("FAIL_CLOSED: decode SplitGo contract: %w", err)
	}
	if !canonicalContract(document) {
		return ContractReceipt{}, fmt.Errorf("FAIL_CLOSED: SplitGo contract registry drift")
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

func canonicalContract(value contractDocument) bool {
	denominator := value.Denominator
	policy := value.DecisionPolicy
	return value.Schema == ContractSchema && value.ContractID == ContractID &&
		value.OperationID == OperationID && denominator.Version == DenominatorVersion &&
		denominator.IndicatorCount == 6 && denominator.OracleCaseCount == 18 &&
		denominator.ExpectedOutcomes["PASS"] == 6 && denominator.ExpectedOutcomes["FAIL"] == 6 &&
		denominator.ExpectedOutcomes["UNKNOWN"] == 6 && policy.Unknown == "PRESERVE_UNKNOWN_AND_BLOCK" &&
		policy.Pass.PassCount == 6 && policy.Pass.FailCount == 0 && policy.Pass.UnknownCount == 0 &&
		policy.Pass.Decision == "PASS" && policy.Otherwise == "BLOCK" &&
		reflect.DeepEqual(value.Indicators, fixedIndicators) && len(value.OracleCases) == 18
}
