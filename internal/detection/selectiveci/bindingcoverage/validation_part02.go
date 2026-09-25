package bindingcoverage

func validatePartitions(partitions []Partition, bindingPairs map[string]string) (map[string]struct{}, map[string]struct{}, Reason) {
	ids := make(map[string]struct{}, len(partitions))
	match := make(map[string]struct{}, len(bindingPairs))
	mismatch := make(map[string]struct{}, len(bindingPairs))
	polarity := make(map[string]struct{}, len(partitions))
	for _, partition := range partitions {
		if reason := validateID(partition.PartitionID); reason != "" {
			return nil, nil, reason
		}
		if _, exists := bindingPairs[partition.PartitionID]; exists {
			return nil, nil, ReasonDuplicateID
		}
		if _, exists := ids[partition.PartitionID]; exists {
			return nil, nil, ReasonDuplicateID
		}
		if reason := validateID(partition.BindingID); reason != "" {
			return nil, nil, reason
		}
		expected, exists := bindingPairs[partition.BindingID]
		if !exists {
			return nil, nil, ReasonUnknownReference
		}
		if !validPolarity(partition.Polarity) {
			return nil, nil, ReasonInvalidEnum
		}
		if reason := validateStageToken(partition.ExpectedStage); reason != "" {
			return nil, nil, reason
		}
		if reason := validateReasonToken(partition.ExpectedReason); reason != "" {
			return nil, nil, reason
		}
		if expected != expectedPair(partition.ExpectedStage, partition.ExpectedReason) {
			return nil, nil, ReasonStalePartition
		}
		key := partition.BindingID + "\x00" + string(partition.Polarity)
		if _, exists := polarity[key]; exists {
			return nil, nil, ReasonDuplicatePolarity
		}
		ids[partition.PartitionID] = struct{}{}
		polarity[key] = struct{}{}
		if partition.Polarity == PolarityMatch {
			match[partition.BindingID] = struct{}{}
		} else {
			mismatch[partition.BindingID] = struct{}{}
		}
	}
	return match, mismatch, ""
}
func missingBindings(bindings []RequiredBinding, match, mismatch map[string]struct{}) ([]string, []string) {
	missingMatch := make([]string, 0)
	missingMismatch := make([]string, 0)
	for _, binding := range bindings {
		if _, exists := match[binding.BindingID]; !exists {
			missingMatch = append(missingMatch, binding.BindingID)
		}
		if _, exists := mismatch[binding.BindingID]; !exists {
			missingMismatch = append(missingMismatch, binding.BindingID)
		}
	}
	return missingMatch, missingMismatch
}
func workUnits(required, partitions, endpointReferences uint64) (uint64, bool) {
	total, ok := addUint64(required, partitions)
	if !ok {
		return 0, false
	}
	return addUint64(total, endpointReferences)
}
