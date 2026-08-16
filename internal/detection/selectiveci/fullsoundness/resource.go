package fullsoundness

import "math"

func resourceVector(state evaluationState) (ResourceVector, Reason) {
	full, ok := resourceTotals(state.fullReceipts)
	if !ok {
		return ResourceVector{}, ReasonResourceOverflow
	}
	selected, ok := resourceTotals(state.selectedReceipts)
	if !ok {
		return ResourceVector{}, ReasonResourceOverflow
	}
	class, ok := compareResources(full, selected)
	if !ok {
		return ResourceVector{}, ReasonResourceOverflow
	}
	return ResourceVector{Full: full, Selected: selected, Class: class}, ""
}

func resourceTotals(receipts map[string]ResourceReceipt) (ResourceTotals, bool) {
	result := ResourceTotals{}
	for _, receipt := range receipts {
		if !validResourceNumbers(receipt) {
			return ResourceTotals{}, false
		}
		var ok bool
		result.CPUCoreNS, ok = addInt64(result.CPUCoreNS, receipt.CPUCoreNS)
		if !ok {
			return ResourceTotals{}, false
		}
		result.ReadBytes, ok = addInt64(result.ReadBytes, receipt.ReadBytes)
		if !ok {
			return ResourceTotals{}, false
		}
		result.WriteBytes, ok = addInt64(result.WriteBytes, receipt.WriteBytes)
		if !ok {
			return ResourceTotals{}, false
		}
		if receipt.PeakRSSBytes > result.PeakRSSBytes {
			result.PeakRSSBytes = receipt.PeakRSSBytes
		}
		denominator, multiplied := multiplyInt64(receipt.WallNS, receipt.AllocatedCPUCount)
		if !multiplied {
			return ResourceTotals{}, false
		}
		result.Utilization.Denominator, ok = addInt64(result.Utilization.Denominator, denominator)
		if !ok {
			return ResourceTotals{}, false
		}
	}
	result.Utilization.Numerator = result.CPUCoreNS
	return result, result.Utilization.Denominator > 0
}

func validResourceNumbers(receipt ResourceReceipt) bool {
	return receipt.CPUCoreNS >= 0 && receipt.AllocatedCPUCount > 0 && receipt.WallNS > 0 && receipt.PeakRSSBytes >= 0 && receipt.ReadBytes >= 0 && receipt.WriteBytes >= 0
}

func compareResources(full, selected ResourceTotals) (ResourceClass, bool) {
	comparisons := []int{compareValue(selected.CPUCoreNS, full.CPUCoreNS), compareValue(selected.PeakRSSBytes, full.PeakRSSBytes), compareValue(selected.ReadBytes, full.ReadBytes), compareValue(selected.WriteBytes, full.WriteBytes)}
	less := false
	for _, comparison := range comparisons {
		if comparison > 0 {
			return ResourceRegressed, true
		}
		less = less || comparison < 0
	}
	if less {
		return ResourceImproved, true
	}
	return ResourceEqual, true
}

func compareValue(left, right int64) int {
	difference, ok := subtractInt64(left, right)
	if !ok || difference == 0 {
		return 0
	}
	if difference < 0 {
		return -1
	}
	return 1
}

func addInt64(left, right int64) (int64, bool) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, false
	}
	if right < 0 && left < math.MinInt64-right {
		return 0, false
	}
	return left + right, true
}

func subtractInt64(left, right int64) (int64, bool) {
	if right == math.MinInt64 {
		if left >= 0 {
			return 0, false
		}
		partial, ok := addInt64(left, math.MaxInt64)
		if !ok {
			return 0, false
		}
		return addInt64(partial, 1)
	}
	return addInt64(left, -right)
}

func multiplyInt64(left, right int64) (int64, bool) {
	if left == 0 || right == 0 {
		return 0, true
	}
	if left == -1 && right == math.MinInt64 || right == -1 && left == math.MinInt64 {
		return 0, false
	}
	product := left * right
	return product, product/right == left
}
