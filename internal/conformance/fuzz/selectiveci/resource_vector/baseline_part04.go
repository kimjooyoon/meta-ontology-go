package resourcevector

func baselinePROV(paths []PathRecord, selected map[string]struct{}) (*PartialVector, bool) {
	vector := &PartialVector{}
	var records, finitePaths, numerator, denominator uint64
	pathCount := 0
	known := true
	for _, path := range paths {
		if _, exists := selected[path.CommandID]; !exists {
			continue
		}
		pathCount++
		if path.RecordIDs == nil || path.Finite == nil || path.ClosureNumerator == nil || path.ClosureDenominator == nil {
			known = false
			continue
		}
		var ok bool
		records, ok = add(records, uint64(len(path.RecordIDs)))
		if !ok {
			known = false
		}
		if *path.Finite {
			finitePaths, ok = add(finitePaths, 1)
			if !ok {
				known = false
			}
			numerator, ok = add(numerator, *path.ClosureNumerator)
			if !ok {
				known = false
			}
			denominator, ok = add(denominator, *path.ClosureDenominator)
			if !ok {
				known = false
			}
		}
	}
	if pathCount == 0 {
		known = false
	}
	if known {
		vector.UniquePROVRecords, vector.FinitePROVPaths = new(records), new(finitePaths)
		vector.ClosureNumerator, vector.ClosureDenominator = new(numerator), new(denominator)
	}
	return vector, known
}
func mergePartial(left, right *PartialVector) {
	left.UniquePROVRecords, left.FinitePROVPaths = right.UniquePROVRecords, right.FinitePROVPaths
	left.ClosureNumerator, left.ClosureDenominator = right.ClosureNumerator, right.ClosureDenominator
}
