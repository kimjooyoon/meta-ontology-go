package generation

import (
	"sort"
	"strings"
)

func operationReceiptIndex(receipts []OperationReceipt) (map[string]OperationReceipt, bool) {
	result := make(map[string]OperationReceipt, len(receipts))
	for _, receipt := range receipts {
		if receipt.ActionIndicatorID == "" {
			return nil, false
		}
		if _, duplicate := result[receipt.ActionIndicatorID]; duplicate {
			return nil, false
		}
		result[receipt.ActionIndicatorID] = receipt
	}
	return result, true
}

func indicatorReceiptIndex(receipts []IndicatorReceipt) (map[string]IndicatorReceipt, bool) {
	result := make(map[string]IndicatorReceipt, len(receipts))
	for _, receipt := range receipts {
		if receipt.ID == "" {
			return nil, false
		}
		if _, duplicate := result[receipt.ID]; duplicate {
			return nil, false
		}
		result[receipt.ID] = receipt
	}
	return result, true
}

func actionObligationID(actionID, indicatorID string) string {
	return actionID + "::" + indicatorID
}

func validActionIndicatorID(value string) bool {
	return strings.HasPrefix(value, "sha256:") &&
		validDigest(strings.TrimPrefix(value, "sha256:"))
}

func sortedSelectedActions(actions []Action) []Action {
	result := append([]Action{}, actions...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].IndicatorID < result[j].IndicatorID
	})
	return result
}
