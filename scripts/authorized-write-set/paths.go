package main

import (
	"io/fs"
	"path"
	"sort"
	"strings"
)

func canonicalPath(value string) bool {
	return value != "." && fs.ValidPath(value) && path.Clean(value) == value &&
		!strings.ContainsRune(value, '\\')
}

func pathSet(values []string) (map[string]bool, bool) {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		if !canonicalPath(value) {
			return nil, false
		}
		result[value] = true
	}
	return result, true
}

func sortedPaths(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func overlapCount(left, right map[string]bool) int {
	count := 0
	for value := range left {
		if right[value] {
			count++
		}
	}
	return count
}

func equalPaths(left, right []string) bool {
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
