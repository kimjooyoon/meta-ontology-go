package sourceexecution

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func resolveEntry(file *syntax.File, name string) (Entry, string, string) {
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	entities := map[string]Binding{}
	var activity *syntax.ActivityDecl
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			entities[value.Name] = Binding{Name: value.Name, ID: value.ID}
		case *syntax.ActivityDecl:
			if value.Name == name {
				activity = value
			}
		}
	}
	if activity == nil {
		return Entry{}, "SOURCE_ENTRY_UNKNOWN", fmt.Sprintf("activity %q is not declared", name)
	}
	parameters := activity.Inputs
	if parameters == nil {
		parameters = activity.Parameters
	}
	inputs := make([]Binding, 0, len(parameters))
	for _, parameter := range parameters {
		binding, ok := entities[parameter.Name]
		if !ok {
			return Entry{}, "SOURCE_INPUT_ENTITY_UNKNOWN", fmt.Sprintf("input entity %q is not declared", parameter.Name)
		}
		inputs = append(inputs, binding)
	}
	outputName := activity.Output
	if outputName == "" {
		outputName = activity.Result.Name
	}
	output, ok := entities[outputName]
	if !ok {
		return Entry{}, "SOURCE_OUTPUT_ENTITY_UNKNOWN", fmt.Sprintf("output entity %q is not declared", outputName)
	}
	return Entry{Package: file.Package.Name, Namespace: file.Namespace.Name,
		Activity: activity.Name, Inputs: inputs, Output: output}, "", ""
}
