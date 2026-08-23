package transformationeffect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type splitGoIndicatorCandidate struct {
	verdict string
	raw     []byte
}

func projectSplitGoReport(reportRaw []byte, requiredIDs []string) ([]generationReceipt, string, []string, error) {
	var reportTree any
	if err := json.Unmarshal(reportRaw, &reportTree); err != nil {
		return nil, "", nil, fmt.Errorf("decode SplitGo evaluator report: %w", err)
	}
	root, ok := reportTree.(map[string]any)
	if !ok {
		return nil, "", nil, errors.New("SplitGo evaluator report must be a JSON object")
	}
	required := make(map[string]struct{}, len(requiredIDs))
	for _, id := range requiredIDs {
		required[id] = struct{}{}
	}
	candidates := make(map[string][]splitGoIndicatorCandidate, len(required))
	unexpected := make(map[string]struct{})
	walkSplitGoReport(reportTree, required, candidates, unexpected)
	forceUnknown, reasons := splitGoReportPolicy(root, requiredIDs, candidates, unexpected)

	receipts := make([]generationReceipt, 0, len(requiredIDs))
	resolution := "EXACT"
	for _, id := range requiredIDs {
		verdict := "UNKNOWN"
		digestMaterial := append(bytes.Clone(reportRaw), []byte("\x00"+id)...)
		if !forceUnknown {
			candidate := candidates[id][0]
			verdict = normalizeSplitGoVerdict(candidate.verdict)
			digestMaterial = candidate.raw
		}
		if verdict == "UNKNOWN" {
			resolution = "LOWER_RESOLUTION"
		}
		receipt, err := newSplitGoReceipt(id, verdict, splitGoDigestHex(digestMaterial))
		if err != nil {
			return nil, "", nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, resolution, reasons, nil
}
