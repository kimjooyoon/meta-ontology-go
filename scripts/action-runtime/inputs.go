package main

import (
	"regexp"
	"sort"
	"strings"
)

var inputPattern = regexp.MustCompile(`^([A-Za-z0-9_-]+):(?:\s|$)`)

func readInputs(lines []string, start, stepIndent int) []string {
	inputs := make([]string, 0)
	withIndent := -1
	for _, line := range lines[start:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := indentation(line)
		if indent <= stepIndent {
			break
		}
		if trimmed == "with:" {
			withIndent = indent
			continue
		}
		if withIndent < 0 {
			continue
		}
		if indent <= withIndent {
			withIndent = -1
			continue
		}
		if match := inputPattern.FindStringSubmatch(trimmed); match != nil {
			inputs = append(inputs, match[1])
		}
	}
	sort.Strings(inputs)
	return uniqueStrings(inputs)
}

func indentation(line string) int {
	return len(line) - len(strings.TrimLeft(line, " \t"))
}

func uniqueStrings(values []string) []string {
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}
