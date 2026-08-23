package operationconformance

func boolValue(value map[string]any, key string) (bool, bool) {
	item, ok := value[key].(bool)
	return item, ok
}

func stringValue(value map[string]any, key string) (string, bool) {
	item, ok := value[key].(string)
	return item, ok && item != ""
}

func intValue(value map[string]any, key string) (int, bool) {
	item, ok := value[key].(float64)
	if !ok || item != float64(int(item)) {
		return 0, false
	}
	return int(item), true
}

func stringsValue(value map[string]any, key string) ([]string, bool) {
	raw, ok := value[key].([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, len(raw))
	for index, item := range raw {
		text, valid := item.(string)
		if !valid {
			return nil, false
		}
		result[index] = text
	}
	return result, true
}

func equalFields(value map[string]any, left, right string) bool {
	first, firstOK := stringValue(value, left)
	second, secondOK := stringValue(value, right)
	return firstOK && secondOK && first == second
}
