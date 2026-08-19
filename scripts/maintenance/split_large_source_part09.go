package main

import (
	"sort"
)

func bytesCount(source []byte, target byte) int {
	n := 0
	for _, ch := range source {
		if ch == target {
			n++
		}
	}
	return n
}
func dedupe(values []string) []string {
	sort.Strings(values)
	out := make([]string, 0, len(values))
	for _, value := range values {
		if len(out) > 0 && out[len(out)-1] == value {
			continue
		}
		out = append(out, value)
	}
	return out
}
