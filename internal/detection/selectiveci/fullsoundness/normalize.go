package fullsoundness

import "sort"

func normalizeInput(input Input) Input {
	input.Obligations = copyObligations(input.Obligations)
	input.Commands = copyCommands(input.Commands)
	input.ImpactedObligationIDs = sortedCopy(input.ImpactedObligationIDs)
	input.SelectedCommandIDs = sortedCopy(input.SelectedCommandIDs)
	input.FullOutcomes = copyOutcomes(input.FullOutcomes)
	input.SelectedOutcomes = copyOutcomes(input.SelectedOutcomes)
	input.FullResourceReceipts = copyReceipts(input.FullResourceReceipts)
	input.SelectedResourceReceipts = copyReceipts(input.SelectedResourceReceipts)
	if input.SelectionReceipt != nil {
		receipt := *input.SelectionReceipt
		receipt.CommandIDs = sortedCopy(receipt.CommandIDs)
		input.SelectionReceipt = &receipt
	}
	sort.Slice(input.Obligations, func(i, j int) bool { return input.Obligations[i].ID < input.Obligations[j].ID })
	sort.Slice(input.Commands, func(i, j int) bool { return input.Commands[i].ID < input.Commands[j].ID })
	sort.Slice(input.FullOutcomes, func(i, j int) bool { return outcomeKey(input.FullOutcomes[i]) < outcomeKey(input.FullOutcomes[j]) })
	sort.Slice(input.SelectedOutcomes, func(i, j int) bool {
		return outcomeKey(input.SelectedOutcomes[i]) < outcomeKey(input.SelectedOutcomes[j])
	})
	sort.Slice(input.FullResourceReceipts, func(i, j int) bool {
		return receiptKey(input.FullResourceReceipts[i]) < receiptKey(input.FullResourceReceipts[j])
	})
	sort.Slice(input.SelectedResourceReceipts, func(i, j int) bool {
		return receiptKey(input.SelectedResourceReceipts[i]) < receiptKey(input.SelectedResourceReceipts[j])
	})
	return input
}

func copyObligations(values []Obligation) []Obligation {
	if values == nil {
		return nil
	}
	result := make([]Obligation, len(values))
	copy(result, values)
	return result
}

func copyOutcomes(values []Outcome) []Outcome {
	if values == nil {
		return nil
	}
	result := make([]Outcome, len(values))
	copy(result, values)
	return result
}

func copyReceipts(values []ResourceReceipt) []ResourceReceipt {
	if values == nil {
		return nil
	}
	result := make([]ResourceReceipt, len(values))
	copy(result, values)
	return result
}

func copyCommands(values []Command) []Command {
	if values == nil {
		return nil
	}
	result := make([]Command, len(values))
	copy(result, values)
	for index := range result {
		result[index].ObligationIDs = sortedCopy(result[index].ObligationIDs)
	}
	return result
}

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
