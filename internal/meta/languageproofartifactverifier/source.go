package languageproofartifactverifier

import (
	"fmt"
	"strconv"
	"strings"
)

type binding struct {
	Name string
	ID   string
}

type projection struct {
	Package   string
	Namespace string
	Activity  string
	Inputs    []binding
	Output    binding
}

func projectSource(raw []byte, selected string) (projection, error) {
	var result projection
	entities := map[string]binding{}
	var inputNames []string
	var outputName string
	for number, rawLine := range strings.Split(string(raw), "\n") {
		line := strings.TrimSpace(rawLine)
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "package "):
			value, err := header(line, "package")
			if err != nil {
				return projection{}, err
			}
			result.Package = value
		case strings.HasPrefix(line, "namespace "):
			value, err := header(line, "namespace")
			if err != nil {
				return projection{}, err
			}
			result.Namespace = value
		case strings.HasPrefix(line, "entity "):
			entity, err := entityDeclaration(line)
			if err != nil {
				return projection{}, err
			}
			entities[entity.Name] = binding{Name: entity.Name, ID: entity.ID}
		case strings.HasPrefix(line, "activity "):
			name, inputs, output, err := activityDeclaration(line)
			if err != nil {
				return projection{}, err
			}
			if name == selected {
				result.Activity, inputNames, outputName = name, inputs, output
			}
		default:
			return projection{}, fmt.Errorf("unsupported source statement at line %d", number+1)
		}
	}
	if result.Package == "" || result.Namespace == "" || result.Activity == "" {
		return projection{}, fmt.Errorf("source projection header or activity missing")
	}
	result.Inputs = make([]binding, len(inputNames))
	for index, name := range inputNames {
		value, ok := entities[name]
		if !ok {
			return projection{}, fmt.Errorf("input entity %q missing", name)
		}
		result.Inputs[index] = value
	}
	value, ok := entities[outputName]
	if !ok {
		return projection{}, fmt.Errorf("output entity %q missing", outputName)
	}
	result.Output = value
	return result, nil
}

func header(line, keyword string) (string, error) {
	fields := strings.Fields(line)
	if len(fields) != 2 || fields[0] != keyword || fields[1] == "" {
		return "", fmt.Errorf("unsupported %s declaration", keyword)
	}
	return fields[1], nil
}

func entityDeclaration(line string) (binding, error) {
	fields := strings.Fields(line)
	if len(fields) != 4 || fields[0] != "entity" || fields[2] != "id" {
		return binding{}, fmt.Errorf("unsupported entity declaration")
	}
	id, err := strconv.Unquote(fields[3])
	if err != nil || fields[1] == "" || id == "" {
		return binding{}, fmt.Errorf("invalid entity declaration")
	}
	return binding{Name: fields[1], ID: id}, nil
}

func activityDeclaration(line string) (string, []string, string, error) {
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
