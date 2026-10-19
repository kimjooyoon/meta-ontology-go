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
		result.Reason = "SOURCE_PARSE_UNAVAILABLE"
		return result
	}
	result.Known = true
	unresolvedType := false
	for _, declaration := range file.Declarations {
		entity, ok := declaration.(*syntax.EntityDecl)
		if !ok {
			continue
		}
		for _, field := range entity.Fields {
			if isTypeRequirement(requirement) && !knownBuiltinFieldType(field.TypeRef.Spelling) {
				unresolvedType = true
			}
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
	if result.MatchCount == 0 && unresolvedType {
		result.Known = false
		result.Reason = "TYPE_REFERENCE_UNRESOLVED"
	}
	return result
}
