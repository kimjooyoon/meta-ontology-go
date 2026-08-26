package selfimprovementcandidate

import "github.com/kimjooyoon/meta-ontology-go/internal/syntax"

var expectedEntities = map[string]string{
	"ReadOnlyImprovementInput":         "gooo://self-improvement/entity/read-only-improvement-input",
	"MissingCapability":                "gooo://self-improvement/entity/missing-capability",
	"NonExecutingImprovementCandidate": "gooo://self-improvement/entity/non-executing-improvement-candidate",
}

var expectedActivities = map[string][2]string{
	"SelectMissingCapability":       {"ReadOnlyImprovementInput", "MissingCapability"},
	"ProposeNonExecutingCandidate": {"MissingCapability", "NonExecutingImprovementCandidate"},
}

func contractDeclarationsKnown(file *syntax.File) bool {
	if file == nil || file.Package == nil || file.Namespace == nil ||
		file.Package.Name != "selfimprovement" || file.Namespace.Name != "selfimprovement" {
		return false
	}
	declarations := file.Decls
	if declarations == nil {
		declarations = file.Declarations
	}
	if len(declarations) != 5 {
		return false
	}
	entities, activities := map[string]string{}, map[string][2]string{}
	for _, declaration := range declarations {
		switch value := declaration.(type) {
		case *syntax.EntityDecl:
			if value.FieldsPresent || len(value.Fields) != 0 {
				return false
			}
			if _, duplicate := entities[value.Name]; duplicate {
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
			if _, duplicate := activities[value.Name]; duplicate {
				return false
			}
			activities[value.Name] = [2]string{inputs[0].Name, value.Output}
		default:
			return false
		}
	}
	return mapsEqual(entities, expectedEntities) && mapsEqual(activities, expectedActivities)
}

func mapsEqual[K comparable, V comparable](left, right map[K]V) bool {
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
