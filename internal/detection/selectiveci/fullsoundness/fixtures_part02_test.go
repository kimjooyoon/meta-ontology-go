package fullsoundness

func findReceipt(values []ResourceReceipt, commandID string) *ResourceReceipt {
	for index := range values {
		if values[index].CommandID == commandID {
			return &values[index]
		}
	}
	return nil
}
func rebindSnapshot(input *Input, snapshot string) {
	input.SnapshotDigest = snapshot
	input.SelectionReceipt.SnapshotDigest = snapshot
	for index := range input.FullResourceReceipts {
		input.FullResourceReceipts[index].SnapshotDigest = snapshot
	}
	for index := range input.SelectedResourceReceipts {
		input.SelectedResourceReceipts[index].SnapshotDigest = snapshot
	}
}
func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseObligations(values []Obligation) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseCommands(values []Command) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseOutcomes(values []Outcome) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reverseReceipts(values []ResourceReceipt) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
