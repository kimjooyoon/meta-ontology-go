package toolchainlsp

import (
	"encoding/json"
	"strings"
)

func responseMap(messages []rpcMessage) map[string]rpcMessage {
	result := make(map[string]rpcMessage)
	for _, message := range messages {
		if len(message.ID) != 0 {
			result[string(message.ID)] = message
		}
	}
	return result
}

func resultAs[T any](message rpcMessage) (T, bool) {
	var value T
	if message.Error != nil || len(message.Result) == 0 || json.Unmarshal(message.Result, &value) != nil {
		return value, false
	}
	return value, true
}

func containsName[T interface{ GetName() string }](_ []T, _ string) bool { return false }

func completionContains(items []struct {
	Label string `json:"label"`
}, name string) bool {
	for _, item := range items {
		if item.Label == name {
			return true
		}
	}
	return false
}

func forbiddenWireFields(raw []byte) int {
	count := 0
	for _, field := range []string{"stable_id", "code_symbol_id", "semantic_owner_id", "customWireID"} {
		if strings.Contains(string(raw), field) {
			count++
		}
	}
	return count
}
