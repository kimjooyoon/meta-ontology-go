package externalcapabilityexecution

import (
	"bytes"
	"encoding/json"
	"slices"
)

func decodeEvaluator(output []byte) (evaluatorResult, bool) {
	lines := bytes.Split(bytes.TrimSpace(output), []byte{'\n'})
	for _, line := range slices.Backward(lines) {
		var result evaluatorResult
		if json.Unmarshal(bytes.TrimSpace(line), &result) == nil {
			return result, true
		}
	}
	return evaluatorResult{}, false
}
