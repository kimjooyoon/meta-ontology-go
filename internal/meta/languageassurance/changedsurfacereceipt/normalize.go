package changedsurfacereceipt

import (
	"path"
	"sort"
	"strings"
)

type observation struct {
	changed           []string
	receipts          []Receipt
	changedSet        map[string]struct{}
	receiptSet        map[string]Receipt
	changedDuplicates int
	receiptDuplicates int
	unknownReceipts   int
	malformedPaths    int
}

func observe(input Input) observation {
	result := observation{changedSet: map[string]struct{}{}, receiptSet: map[string]Receipt{}}
	for _, raw := range input.ChangedSurfaces {
		normalized, ok := normalizeSurface(raw)
		if !ok {
			result.malformedPaths++
			continue
		}
		if _, duplicate := result.changedSet[normalized]; duplicate {
			result.changedDuplicates++
			continue
		}
		result.changedSet[normalized] = struct{}{}
		result.changed = append(result.changed, normalized)
	}
	for _, receipt := range input.Receipts {
		normalized, ok := normalizeSurface(receipt.SurfaceID)
		if !ok {
			result.malformedPaths++
			continue
		}
		receipt.SurfaceID = normalized
		if _, duplicate := result.receiptSet[normalized]; duplicate {
			result.receiptDuplicates++
			continue
		}
		if receipt.Decision == "UNKNOWN" || receipt.Resolution == ResolutionUnknown {
			result.unknownReceipts++
		}
		result.receiptSet[normalized] = receipt
		result.receipts = append(result.receipts, receipt)
	}
	sort.Strings(result.changed)
	sort.Slice(result.receipts, func(i, j int) bool { return result.receipts[i].SurfaceID < result.receipts[j].SurfaceID })
	return result
}

func normalizeSurface(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") {
		return "", false
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || cleaned != value {
		return "", false
	}
	return cleaned, true
}
