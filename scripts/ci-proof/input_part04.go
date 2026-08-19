package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]bool, len(left))
	for _, value := range left {
		seen[value] = true
	}
	for _, value := range right {
		if !seen[value] {
			return false
		}
	}
	return true
}
func readJSON[T any](filename string) (T, error) {
	var value T
	data, err := os.ReadFile(filename)
	if err != nil {
		return value, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return value, fmt.Errorf("empty JSON input %s", filename)
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}
	return value, nil
}
func readStrictJSON[T any](filename string) (T, error) {
	var value T
	data, err := os.ReadFile(filename)
	if err != nil {
		return value, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, err
	}
	return value, nil
}
func isProofJob(name string) bool {
	for _, canonical := range proofJobs {
		if name == canonical {
			return true
		}
	}
	return false
}
func validSHA(value string) bool {
	if len(value) != 40 || value == strings.Repeat("0", 40) {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
