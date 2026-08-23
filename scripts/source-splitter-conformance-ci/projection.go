package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type evaluationProjection struct {
	Resolution string              `json:"resolution"`
	Receipts   []receiptProjection `json:"receipts"`
}

type receiptProjection struct {
	IndicatorID string `json:"indicator_id"`
	Verdict     string `json:"verdict"`
}

type evaluationStats struct {
	resolution   string
	passCount    int
	failCount    int
	unknownCount int
}

func projectEvaluation(raw []byte, required []string) (evaluationStats, error) {
	var projection evaluationProjection
	if err := json.Unmarshal(raw, &projection); err != nil {
		return evaluationStats{}, err
	}
	if projection.Resolution == "" || len(projection.Receipts) == 0 {
		return evaluationStats{}, describeProjection(raw)
	}
	if len(projection.Receipts) != len(required) {
		return evaluationStats{}, fmt.Errorf(
			"receipt denominator = %d, want %d", len(projection.Receipts), len(required),
		)
	}
	stats := evaluationStats{resolution: projection.Resolution}
	for index, receipt := range projection.Receipts {
		if receipt.IndicatorID != required[index] {
			return evaluationStats{}, fmt.Errorf(
				"receipt[%d].indicator_id = %q, want %q", index, receipt.IndicatorID, required[index],
			)
		}
		switch strings.ToUpper(receipt.Verdict) {
		case "PASS":
			stats.passCount++
		case "FAIL":
			stats.failCount++
		case "UNKNOWN":
			stats.unknownCount++
		default:
			return evaluationStats{}, fmt.Errorf("receipt %q has verdict %q", receipt.IndicatorID, receipt.Verdict)
		}
	}
	return stats, nil
}

func describeProjection(raw []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return err
	}
	keys := make([]string, 0, len(root))
	for key := range root {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return fmt.Errorf("evaluation lacks resolution or receipts; root keys = %v", keys)
}
