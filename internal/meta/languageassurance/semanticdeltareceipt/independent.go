package semanticdeltareceipt

import (
	"errors"
	"sort"
	"strings"
)

var errIndependentSource = errors.New("independent source projection unavailable")

func splitIndependentLines(raw []byte) []string {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		result = append(result, strings.TrimSpace(line))
	}
	return result
}

func independentEntity(line string) (string, string, bool) {
	body := strings.TrimPrefix(line, "entity ")
	parts := strings.SplitN(body, " id ", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	name, quoted := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if name == "" || len(quoted) < 3 || quoted[0] != '"' || quoted[len(quoted)-1] != '"' {
		return "", "", false
	}
	return name, quoted[1 : len(quoted)-1], true
}

func independentActivity(line string) (activityDecl, bool) {
	body := strings.TrimPrefix(line, "activity ")
	open := strings.Index(body, "(")
	close := strings.Index(body, ") ->")
	if open < 1 || close <= open {
		return activityDecl{}, false
	}
	name := strings.TrimSpace(body[:open])
	args := strings.TrimSpace(body[open+1 : close])
	output := strings.TrimSpace(body[close+4:])
	if name == "" || output == "" {
		return activityDecl{}, false
	}
	inputs := []string{}
	if args != "" {
		for _, value := range strings.Split(args, ",") {
			value = strings.TrimSpace(value)
			if value == "" {
				return activityDecl{}, false
			}
			inputs = append(inputs, value)
		}
	}
	return activityDecl{name: name, inputs: inputs, output: output}, true
}

func sortProjected(source *projectedSource) {
	sort.Slice(source.nodes, func(i, j int) bool { return source.nodes[i].ID < source.nodes[j].ID })
	sort.Slice(source.facts, func(i, j int) bool { return factLess(source.facts[i], source.facts[j]) })
	sort.Slice(source.claims, func(i, j int) bool { return source.claims[i].ID < source.claims[j].ID })
}
