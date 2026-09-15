package transformationeffect

import "encoding/json"

func walkSplitGoReport(node any, required map[string]struct{}, candidates map[string][]splitGoIndicatorCandidate, unexpected map[string]struct{}) {
	switch value := node.(type) {
	case map[string]any:
		id, hasID := splitGoIndicatorID(value)
		verdict, hasVerdict := splitGoIndicatorVerdict(value)
		if hasID && hasVerdict {
			raw, _ := json.Marshal(value)
			if _, expected := required[id]; expected {
				candidate := splitGoIndicatorCandidate{verdict: verdict, raw: raw}
				candidates[id] = append(candidates[id], candidate)
			} else {
				unexpected[id] = struct{}{}
			}
		}
		for _, child := range value {
			walkSplitGoReport(child, required, candidates, unexpected)
		}
	case []any:
		for _, child := range value {
			walkSplitGoReport(child, required, candidates, unexpected)
		}
	}
}

func splitGoIndicatorID(object map[string]any) (string, bool) {
	for _, key := range []string{"indicator_id", "id"} {
		if value, ok := splitGoDirectString(object, key); ok {
			return value, true
		}
	}
	return "", false
}

func splitGoIndicatorVerdict(object map[string]any) (string, bool) {
	for _, key := range []string{"verdict", "decision", "status"} {
		if value, ok := splitGoDirectString(object, key); ok {
			return value, true
		}
	}
	return "", false
}

func splitGoDirectString(object map[string]any, key string) (string, bool) {
	value, exists := object[key]
	if !exists {
		return "", false
	}
	text, ok := value.(string)
	return text, ok
}
