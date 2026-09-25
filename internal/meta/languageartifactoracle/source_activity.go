package languageartifactoracle

import (
	"fmt"
	"strings"
)

func parseActivity(line string) (string, []string, string, error) {
	declaration := strings.TrimPrefix(line, "activity ")
	open := strings.IndexByte(declaration, '(')
	close := strings.Index(declaration, ") -> ")
	if open <= 0 || close <= open {
		return "", nil, "", fmt.Errorf("unsupported activity declaration")
	}
	name := strings.TrimSpace(declaration[:open])
	output := strings.TrimSpace(declaration[close+5:])
	if name == "" || output == "" {
		return "", nil, "", fmt.Errorf("invalid activity declaration")
	}
	rawInputs := strings.TrimSpace(declaration[open+1 : close])
	if rawInputs == "" {
		return name, []string{}, output, nil
	}
	parts := strings.Split(rawInputs, ",")
	inputs := make([]string, len(parts))
	for index, part := range parts {
		inputs[index] = strings.TrimSpace(part)
		if inputs[index] == "" {
			return "", nil, "", fmt.Errorf("invalid activity input")
		}
	}
	return name, inputs, output, nil
}
