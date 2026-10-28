package syntaxregistration

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

type ProfileExerciseWitness struct {
	EntityID  string `json:"entity_id"`
	FieldID   string `json:"field_id"`
	StartByte int    `json:"start_byte"`
	EndByte   int    `json:"end_byte"`
}

// ProfileExercise observes authored syntax. Its status is a checked Gooo
// assessment, independent of parser compatibility and registration authority.
type ProfileExercise struct {
	Status       string                  `json:"status"`
	Reason       string                  `json:"reason,omitempty"`
	Known        bool                    `json:"source_observation_known"`
	MatchCount   int                     `json:"match_count"`
	FirstWitness *ProfileExerciseWitness `json:"first_witness,omitempty"`
}

// observeProfileExercise reads the same pinned bytes used for parser replay.
// It observes authored field syntax only; nominal type resolution and
// registration acceptance remain separate operations.
func observeProfileExercise(path string, raw []byte, requirement string) ProfileExercise {
	result := ProfileExercise{Status: "UNKNOWN"}
	if !strings.HasPrefix(requirement, "entity-") {
		return result
	}
	file, diagnostics := syntax.ParseFileWithEntityFieldsSupport(path, string(raw), latestExerciseProfile())
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

func latestExerciseProfile() syntax.EntityFieldsSupport {
	profiles := syntax.SupportedEntityFieldsProfiles()
	if len(profiles) == 0 {
		return syntax.EntityFieldsSupport{}
	}
	support, _ := syntax.EntityFieldsSupportForProfileID(profiles[len(profiles)-1].ID)
	return support
}
