package proofchoicealgebra

import "strings"

func validateItems(items []Item) (map[string]Item, map[string]Item, string) {
	byID := make(map[string]Item, len(items))
	claims := make(map[string]Item)
	for _, item := range items {
		if item.ID == "" || item.Statement == "" {
			return nil, nil, "ITEM_METADATA_MISSING"
		}
		if !item.Choice.Valid() {
			return nil, nil, "PROOF_CHOICE_MISSING"
		}
		if item.Kind != Claim && item.Kind != Metric {
			return nil, nil, "ITEM_KIND_UNKNOWN"
		}
		if metadataUnknown(item.Producer, item.Consumer, item.MetaOperation, item.Stage, item.Step, item.Reason) {
			return nil, nil, "UNKNOWN_CONTEXT"
		}
		if item.Kind == Metric && (item.Denominator != FixedDenominator || item.Numerator < 0 || item.Numerator > item.Denominator) {
			return nil, nil, "FIXED_DENOMINATOR_MISMATCH"
		}
		if previous, exists := byID[item.ID]; exists && !sameItem(previous, item) {
			return nil, nil, "PROOF_CHOICE_CONTRADICTION"
		}
		byID[item.ID] = item
		if item.Kind == Claim {
			claims[item.ID] = item
		}
	}
	return byID, claims, ""
}

func metadataUnknown(values ...string) bool {
	for _, value := range values {
		if value == "" || strings.EqualFold(value, "UNKNOWN") {
			return true
		}
	}
	return false
}

func sameItem(left, right Item) bool {
	left.Line, right.Line = 0, 0
	return left == right
}
