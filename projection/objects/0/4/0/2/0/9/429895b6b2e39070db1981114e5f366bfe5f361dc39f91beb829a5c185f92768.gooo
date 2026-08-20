package fullsoundness

import (
	"sort"
)

func normalizeOutput(output Output) Output {
	output.FullFailureCommandIDs = sortedOutput(output.FullFailureCommandIDs)
	output.SelectedFailureCommandIDs = sortedOutput(output.SelectedFailureCommandIDs)
	output.OmittedCommandIDs = sortedOutput(output.OmittedCommandIDs)
	return output
}
func sortedOutput(values []string) []string {
	if values == nil {
		return []string{}
	}
	return sortedCopy(values)
}
func sortedCopy(values []string) []string {
	if values == nil {
		return nil
	}
	result := make([]string, len(values))
	copy(result, values)
	sort.Strings(result)
	return result
}
func outcomeKey(outcome Outcome) string {
	return outcome.CommandID + "\x00" + string(outcome.Status) + "\x00" + outcome.FailureCode + "\x00" + outcome.OutputDigest
}
func receiptKey(receipt ResourceReceipt) string {
	return receipt.CommandID + "\x00" + receipt.SnapshotDigest + "\x00" + receipt.ToolchainDigest + "\x00" + receipt.RunnerDigest
}
func stringSet(values []string) (map[string]struct{}, bool) {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := result[value]; exists {
			return nil, false
		}
		result[value] = struct{}{}
	}
	return result, true
}
func sameStringSet(left, right []string) bool {
	leftSet, leftOK := stringSet(left)
	rightSet, rightOK := stringSet(right)
	if !leftOK || !rightOK || len(leftSet) != len(rightSet) {
		return false
	}
	for value := range leftSet {
		if _, exists := rightSet[value]; !exists {
			return false
		}
	}
	return true
}
