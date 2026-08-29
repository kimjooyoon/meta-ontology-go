package generation

import (
	"fmt"
	"sort"
)

func missingReceiptUnknown(action Action, required string) ReceiptUnknown {
	return ReceiptUnknown{
		ActionIndicatorID: action.IndicatorID, RequiredIndicatorID: required,
		Operation: action.Operation, Activity: action.Activity, Output: action.Output,
		Executor: action.Executor, Evaluator: action.Evaluator,
		Stage: ReceiptUnknownStage, Step: ReceiptUnknownStep,
		Reason: ReceiptReasonMissingIndicator, UnknownClass: ReceiptUnknownClassDirectMissing,
		NextOperation: action.Executor, BlockedBy: []string{},
	}
}

func malformedReceiptUnknown(action Action, required, class string) ReceiptUnknown {
	return ReceiptUnknown{
		ActionIndicatorID: action.IndicatorID, RequiredIndicatorID: required,
		Operation: action.Operation, Activity: action.Activity, Output: action.Output,
		Executor: action.Executor, Evaluator: action.Evaluator,
		Stage: ReceiptUnknownStage, Step: ReceiptUnknownStep,
		Reason: ReceiptReasonUnknownIndicator, UnknownClass: class,
		NextOperation: action.Evaluator, BlockedBy: []string{},
	}
}

func normalizeReceiptUnknowns(unknowns []ReceiptUnknown) []ReceiptUnknown {
	result := append([]ReceiptUnknown{}, unknowns...)
	for index := range result {
		if result[index].BlockedBy == nil {
			result[index].BlockedBy = []string{}
		}
		sort.Strings(result[index].BlockedBy)
	}
	sort.Slice(result, func(left, right int) bool {
		return receiptUnknownKey(result[left]) < receiptUnknownKey(result[right])
	})
	return result
}

func receiptUnknownKey(unknown ReceiptUnknown) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s",
		unknown.ActionIndicatorID, unknown.RequiredIndicatorID,
		unknown.Operation, unknown.Stage, unknown.Reason)
}

func validReceiptUnknowns(unknowns []ReceiptUnknown) bool {
	canonical := normalizeReceiptUnknowns(unknowns)
	if len(canonical) != len(unknowns) {
		return false
	}
	for index, unknown := range unknowns {
		if unknown.ActionIndicatorID == "" || unknown.RequiredIndicatorID == "" ||
			unknown.Operation == "" || unknown.Activity == "" || unknown.Output == "" ||
			unknown.Executor == "" || unknown.Evaluator == "" ||
			unknown.Stage != ReceiptUnknownStage || unknown.Step != ReceiptUnknownStep ||
			unknown.NextOperation == "" || !validActionIndicatorID(unknown.ActionIndicatorID) ||
			unknown.BlockedBy == nil ||
			(index > 0 && receiptUnknownKey(unknowns[index-1]) >= receiptUnknownKey(unknown)) {
			return false
		}
		switch unknown.Reason {
		case ReceiptReasonMissingIndicator, ReceiptReasonUnknownIndicator:
		default:
			return false
		}
		switch unknown.UnknownClass {
		case ReceiptUnknownClassDirectMissing, ReceiptUnknownClassMalformedEvidence,
			ReceiptUnknownClassUnexpectedEvidence:
		default:
			return false
		}
		for position := 1; position < len(unknown.BlockedBy); position++ {
			if unknown.BlockedBy[position-1] >= unknown.BlockedBy[position] {
				return false
			}
		}
	}
	return true
}
