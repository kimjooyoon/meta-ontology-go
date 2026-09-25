package selfimprovementtransport

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func contractKnown(file *syntax.File) bool {
	if file == nil || file.Package == nil || file.Namespace == nil ||
		file.Package.Name != "selfimprovementtransport" || file.Namespace.Name != "selfimprovementtransport" {
		return false
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	if len(declarations) != len(expectedEntities)+len(expectedActivities) {
		return false
	}
	entities, activities := map[string]string{}, map[string]expectedActivity{}
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			if value.FieldsPresent || len(value.Fields) != 0 {
				return false
			}
			entities[value.Name] = value.ID
		case *syntax.ActivityDecl:
			inputs := value.Inputs
			if inputs == nil {
				inputs = value.Parameters
			}
			if len(inputs) != 1 {
				return false
			}
			if value.ValueProgramPresent != (value.ValueProgram != "") {
				return false
			}
			activities[value.Name] = expectedActivity{
				Input: inputs[0].Name, Output: value.Output, Program: value.ValueProgram,
			}
		default:
			return false
		}
	}
	if !equalMap(entities, expectedEntities) || len(activities) != len(expectedActivities) {
		return false
	}
	for name, expected := range expectedActivities {
		actual, ok := activities[name]
		if !ok || actual.Input != expected.Input || actual.Output != expected.Output {
			return false
		}
		if name == "ResolveConsumerSubject" {
			if actual.Program != expected.Program && !strings.HasPrefix(actual.Program, expected.Program+";") {
				return false
			}
			continue
		}
		if actual.Program != expected.Program {
			return false
		}
	}
	return true
}

func equalMap[K comparable, V comparable](left, right map[K]V) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
