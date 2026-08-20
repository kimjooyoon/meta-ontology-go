package generator

import "strings"

func join(values []string, separator string) string {
	if len(values) == 0 {
		return ""
	}
	var result strings.Builder
	result.WriteString(values[0])
	for _, value := range values[1:] {
		result.WriteString(separator + value)
	}
	return result.String()
}
