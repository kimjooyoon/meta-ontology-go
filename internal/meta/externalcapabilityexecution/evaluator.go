package externalcapabilityexecution

import (
	"bytes"
	"encoding/json"
)

func decodeEvaluator(output []byte) (evaluatorResult, bool) {
	lines := bytes.Split(bytes.TrimSpace(output), []byte{'\n'})
	for index := len(lines) - 1; index >= 0; index-- {
		var result evaluatorResult
		if json.Unmarshal(bytes.TrimSpace(lines[index]), &result) == nil {
			return result, true
		}
	}
	return evaluatorResult{}, false
}
