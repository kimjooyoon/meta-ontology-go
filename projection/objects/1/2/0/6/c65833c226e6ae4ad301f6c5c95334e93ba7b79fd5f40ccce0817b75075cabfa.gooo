package bindingcoverage

func validatePrecedence(entries []Precedence) (map[string]bool, string) {
	seenRanks := make(map[int64]bool, len(entries))
	seenPairs := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.Rank < 0 || !validToken(entry.Stage, "stage:") || !validToken(entry.Reason, "reason:") {
			return nil, "UNKNOWN_PRECEDENCE"
		}
		pair := evidenceKey(entry.Stage, entry.Reason)
		if seenRanks[entry.Rank] || seenPairs[pair] {
			return nil, "UNKNOWN_PRECEDENCE"
		}
		seenRanks[entry.Rank] = true
		seenPairs[pair] = true
	}
	return seenPairs, ""
}
func validateBindings(bindings []Binding, precedence map[string]bool) (map[string]Binding, string) {
	seen := make(map[string]Binding, len(bindings))
	for _, binding := range bindings {
		if !validStableID(binding.ID) || !validStableID(binding.FromFieldID) || !validStableID(binding.ToFieldID) || binding.FromFieldID == binding.ToFieldID {
			return nil, "UNKNOWN_BINDING"
		}
		if !validKind(binding.Kind) || !validToken(binding.ExpectedStage, "stage:") || !validToken(binding.ExpectedReason, "reason:") {
			return nil, "UNKNOWN_BINDING"
		}
		if !precedence[evidenceKey(binding.ExpectedStage, binding.ExpectedReason)] || seen[binding.ID].ID != "" {
			return nil, "UNKNOWN_BINDING"
		}
		seen[binding.ID] = binding
	}
	return seen, ""
}
func validatePartitions(partitions []Partition, bindings map[string]Binding, precedence map[string]bool) string {
	seen := make(map[string]bool, len(partitions))
	for _, partition := range partitions {
		binding, ok := bindings[partition.BindingID]
		if !ok {
			return "DANGLING_PARTITION"
		}
		if partition.Polarity != PolarityMatch && partition.Polarity != PolarityMismatch {
			return "UNKNOWN_PARTITION"
		}
		if !validToken(partition.Stage, "stage:") || !validToken(partition.Reason, "reason:") || !precedence[evidenceKey(partition.Stage, partition.Reason)] {
			return "UNKNOWN_PARTITION"
		}
		if partition.Stage != binding.ExpectedStage || partition.Reason != binding.ExpectedReason {
			return "STALE_PARTITION"
		}
		key := partition.BindingID + "\x00" + partition.Polarity
		if seen[key] {
			return "DUPLICATE_PARTITION_POLARITY"
		}
		seen[key] = true
	}
	return ""
}
