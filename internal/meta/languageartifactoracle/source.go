package languageartifactoracle

import (
	"fmt"
	"strings"
)

func projectSource(raw []byte, selectedEntry string) (projection, error) {
	var result projection
	entities := map[string]artifactBinding{}
	var inputNames []string
	var outputName string
	for number, rawLine := range strings.Split(string(raw), "\n") {
		line := strings.TrimSpace(rawLine)
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "package "):
			value, err := parseHeader(line, "package")
			if err != nil {
				return projection{}, err
			}
			result.Package = value
		case strings.HasPrefix(line, "namespace "):
			value, err := parseHeader(line, "namespace")
			if err != nil {
				return projection{}, err
			}
			result.Namespace = value
		case strings.HasPrefix(line, "entity "):
			entity, err := parseEntity(line)
			if err != nil {
				return projection{}, err
			}
			entities[entity.Name] = entity
		case strings.HasPrefix(line, "activity "):
			name, inputs, output, err := parseActivity(line)
			if err != nil {
				return projection{}, err
			}
			if name == selectedEntry {
				result.Activity, inputNames, outputName = name, inputs, output
			}
		default:
			return projection{}, fmt.Errorf("unsupported source statement at line %d", number+1)
		}
	}
	if result.Package == "" || result.Namespace == "" || result.Activity == "" {
		return projection{}, fmt.Errorf("source projection header or entry missing")
	}
	result.Inputs = make([]artifactBinding, len(inputNames))
	for index, name := range inputNames {
		binding, ok := entities[name]
		if !ok {
			return projection{}, fmt.Errorf("input entity %q missing", name)
		}
		result.Inputs[index] = binding
	}
	output, ok := entities[outputName]
	if !ok {
		return projection{}, fmt.Errorf("output entity %q missing", outputName)
	}
	result.Output = output
	return result, nil
}
