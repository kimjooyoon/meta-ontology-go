package bodycodegen

import "strings"

// InputParameter follows source declaration order and retains its body binding.
// Unnamed inputs use input/inputN; wire ports remain positional with named inputs.
type InputParameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func parameterDeclaration(parameters []InputParameter) string {
	parts := make([]string, len(parameters))
	for i, parameter := range parameters {
		parts[i] = parameter.Name + " " + parameter.Type
	}
	return strings.Join(parts, ", ")
}

func parameterTypeLabel(parameters []InputParameter) string {
	parts := make([]string, len(parameters))
	for i, parameter := range parameters {
		parts[i] = parameter.Type
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, ",") + ")"
}

func parameterNames(parameters []InputParameter) map[string]bool {
	result := make(map[string]bool, len(parameters))
	for _, parameter := range parameters {
		result[parameter.Name] = true
	}
	return result
}
