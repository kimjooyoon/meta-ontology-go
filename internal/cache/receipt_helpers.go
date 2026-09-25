package cache

func hasDuplicateDigests(values []Digest) bool {
	seen := make(map[Digest]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func hasDuplicateEvidenceRefs(values []EvidenceRef) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value.Name]; exists {
			return true
		}
		seen[value.Name] = struct{}{}
	}
	return false
}

func digestSliceEqual(left, right []Digest) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func evidenceRefsEqual(left, right []EvidenceRef) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
