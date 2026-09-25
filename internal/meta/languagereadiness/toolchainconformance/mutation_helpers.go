package toolchainconformance

import "strings"

func mutateHead(value map[string]any) {
	value["head_sha"] = "0000000000000000000000000000000000000000"
	source, ok := value["source"].(map[string]any)
	if !ok {
		source = map[string]any{}
		value["source"] = source
	}
	source["expected_head_sha"] = "0000000000000000000000000000000000000000"
}

func mutateSummary(value map[string]any, key string, replacement int) {
	summary, ok := value["summary"].(map[string]any)
	if !ok {
		summary = map[string]any{}
		value["summary"] = summary
	}
	for candidate := range summary {
		if strings.EqualFold(candidate, key) {
			summary[candidate] = replacement
			return
		}
	}
	summary[key] = replacement
}

func mutateBooleanList(value map[string]any, listKey, field string) {
	list, ok := value[listKey].([]any)
	if !ok || len(list) == 0 {
		value[listKey] = []any{map[string]any{field: false}}
		return
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		item = map[string]any{}
		list[0] = item
	}
	item[field] = false
}
