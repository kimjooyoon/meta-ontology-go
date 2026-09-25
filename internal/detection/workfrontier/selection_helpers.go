package workfrontier

func selectionKey(path RepairPath) string {
	return paddedUint64(uint64(path.PolicyPriority)) + path.stableID()
}

func paddedUint64(value uint64) string {
	const width = 20
	digits := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits)
}

func conflicts(left, right RepairPath) bool {
	return intersects(left.WriteSet, right.WriteSet) ||
		intersects(left.WriteSet, right.ReadSet) || intersects(right.WriteSet, left.ReadSet)
}

func intersects(left, right []string) bool {
	if len(left) > len(right) {
		left, right = right, left
	}
	set := make(map[string]struct{}, len(left))
	for _, value := range left {
		set[value] = struct{}{}
	}
	for _, value := range right {
		if _, ok := set[value]; ok {
			return true
		}
	}
	return false
}

func unknownAll(input Input) Result {
	result := Result{Status: DecisionUnknown, Quality: "UNKNOWN", FullSuiteRequired: true}
	for _, path := range input.Paths {
		if id := path.stableID(); id != "" {
			result.Unknown = append(result.Unknown, id)
		}
	}
	return normalizeResult(result)
}

func duplicatePathIDs(paths []RepairPath) bool {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		id := path.stableID()
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			return true
		}
		seen[id] = struct{}{}
	}
	return false
}
