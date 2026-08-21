package generation

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
