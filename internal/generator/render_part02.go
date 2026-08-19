package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
)

func renderInputs(inputs []Port) string {
	parts := make([]string, len(inputs))
	for index, input := range inputs {
		parts[index] = input.GoName + " " + input.GoType
	}
	return strings.Join(parts, ", ")
}
func renderOutputs(outputs []Port) string {
	switch len(outputs) {
	case 0:
		return ""
	case 1:
		return " " + outputs[0].GoType
	default:
		parts := make([]string, len(outputs))
		for index, output := range outputs {
			parts[index] = output.GoType
		}
		return " (" + strings.Join(parts, ", ") + ")"
	}
}
func indentSnippet(snippet, prefix string) string {
	lines := strings.Split(snippet, "\n")
	var output strings.Builder
	for index, line := range lines {
		if line != "" {
			output.WriteString(prefix)
		}
		output.WriteString(line)
		if index < len(lines)-1 {
			output.WriteByte('\n')
		}
	}
	return output.String()
}
func formatBlock(packageName, block string) ([]byte, error) {
	source := []byte("package " + packageName + "\n\n" + block)
	originalMarkers, err := parseMarkers(source)
	if err != nil {
		return nil, fmt.Errorf("generator: parse block markers: %w", err)
	}
	formatted, err := format.Source(source)
	if err != nil {
		return nil, err
	}
	formatted, err = restoreSlotBodies(formatted, originalMarkers.Slots)
	if err != nil {
		return nil, err
	}
	marker := []byte(generatedStartPrefix)
	start := bytes.Index(formatted, marker)
	if start < 0 {
		return nil, fmt.Errorf("formatted block lost generated marker")
	}
	result := bytes.TrimLeft(formatted[start:], "\n")
	result = bytes.TrimRight(result, "\n")
	result = append(result, '\n')
	return result, nil
}
