package syntaxregistration

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// observeProfileExercise reads the same pinned bytes used for parser replay.
// It observes authored field syntax only; nominal type resolution and
// registration acceptance remain separate operations.
func observeProfileExercise(path string, raw []byte, requirement string) ProfileExercise {
	result := ProfileExercise{Status: "UNKNOWN"}
	if !strings.HasPrefix(requirement, "entity-") {
		return result
	}
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(path, string(raw), syntax.EntityFieldsV6Support())
	if file == nil || diagnostics.HasErrors() {
		return result
	}
	result.Known = true
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			continue
		}
		for _, field := range entity.Fields {
			if !fieldExercisesRequirement(field, requirement) {
				continue
			}
			result.MatchCount++
			if result.FirstWitness == nil {
				result.FirstWitness = &ProfileExerciseWitness{EntityID: entity.ID, FieldID: field.ID,
					StartByte: field.Span.Start.Offset, EndByte: field.Span.End.Offset}
			}
		}
	}
	return result
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
