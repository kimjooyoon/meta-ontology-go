package semanticdelta

import (
	"fmt"
	"sort"
	"strings"
)

func normalizeValues(label string, values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make([]string, len(values))
	for i, value := range values {
		normalized, err := normalizeToken(label, value)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		result[i] = normalized
	}
	sort.Strings(result)
	return uniqueStrings(result), nil
}
func normalizeToken(label, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is empty", label)
	}
	if strings.IndexFunc(value, func(r rune) bool { return r == '\n' || r == '\r' || r == '\t' || r == ' ' }) >= 0 {
		return "", fmt.Errorf("%s %q contains whitespace", label, value)
	}
	return value, nil
}
func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}
func uniqueNodes(nodes []Node) []Node {
	result := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		if len(result) == 0 || result[len(result)-1] != node {
			result = append(result, node)
		}
	}
	return result
}
func uniqueFacts(facts []Fact) []Fact {
	result := make([]Fact, 0, len(facts))
	for _, fact := range facts {
		if len(result) == 0 || result[len(result)-1] != fact {
			result = append(result, fact)
		}
	}
	return result
}
func factKey(fact Fact) string {
	return fact.Subject + "\x00" + fact.Predicate + "\x00" + fact.Object
}
