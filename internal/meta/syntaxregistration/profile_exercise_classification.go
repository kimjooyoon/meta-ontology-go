package syntaxregistration

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func isTypeRequirement(requirement string) bool {
	return requirement == "entity-string" || requirement == "entity-boolean" ||
		requirement == "entity-integer"
}

func knownBuiltinFieldType(spelling string) bool {
	for _, name := range [...]string{"string", "boolean", "integer"} {
		if spelling == name || spelling == "gooo:"+name || spelling == "urn:gooo:type:"+name {
			return true
		}
	}
	return false
}

func fieldExercisesRequirement(field syntax.FieldDecl, requirement string) bool {
	switch requirement {
	case "entity-string", "entity-boolean", "entity-integer":
		name := strings.TrimPrefix(requirement, "entity-")
		spelling := field.TypeRef.Spelling
		return spelling == name || spelling == "gooo:"+name || spelling == "urn:gooo:type:"+name
	case "entity-optional":
		return field.Presence == syntax.FieldPresenceOptional
	case "entity-many":
		return field.Cardinality == syntax.FieldCardinalityMany
	case "entity-text-bound":
		return field.TextBound.IsSet()
	default:
		return false
	}
}

func expectedExerciseStatus(requirement string, exercise ProfileExercise) string {
	if requirement == "" {
		return "NOT_REQUESTED"
	}
	if !strings.HasPrefix(requirement, "entity-") {
		return "ROUTE_ONLY"
	}
	if !exercise.Known {
		return "UNKNOWN"
	}
	if exercise.MatchCount > 0 {
		return "EXERCISED"
	}
	return "NOT_EXERCISED"
}
