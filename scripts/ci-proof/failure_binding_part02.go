package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func failureScope(binding failureBinding) (string, error) {
	if binding.Event == "pull_request" {
		if binding.PRNumber <= 0 {
			return "", fmt.Errorf("pull-request failure requires an exact PR number")
		}
		return "pr", nil
	}
	if binding.Event == "push" && (binding.BaseRef == "dev" || binding.BaseRef == "main") {
		if binding.EventRef != "refs/heads/"+binding.BaseRef || binding.OwnerBranch != binding.BaseRef {
			return "", fmt.Errorf("protected push owner must equal the exact protected branch")
		}
		return binding.BaseRef, nil
	}
	return "", fmt.Errorf("failure scope cannot be resolved without guessing")
}
func failurePositiveInt(name string) (int64, error) {
	value, err := strconv.ParseInt(os.Getenv(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("missing or invalid failure binding field %s", name)
	}
	return value, nil
}
func failureOptionalInt(name string) (int64, error) {
	value := os.Getenv(name)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("invalid failure binding field %s", name)
	}
	return parsed, nil
}
func containsUnknown(value string) bool {
	lower := strings.ToLower(value)
	return lower == "unknown" || lower == "unavailable" || strings.Contains(lower, "<unknown>")
}
func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
