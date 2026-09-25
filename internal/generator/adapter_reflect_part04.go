package generator

import (
	"fmt"
)

func appendReflectedRelation(relation reflectedRelation, model *SemanticIR, entities, activities map[string]int, kinds map[string]string) error {
	subjectKind, subjectExists := kinds[relation.subject]
	objectKind, objectExists := kinds[relation.object]
	if !subjectExists || !objectExists {
		return fmt.Errorf("generator: reflected relation %q has missing endpoint (%q, %q)", relation.predicate, relation.subject, relation.object)
	}
	switch relation.predicate {
	case "used":
		if subjectKind != "activity" || objectKind != "entity" {
			return fmt.Errorf("generator: unsupported endpoint kinds for relation %q", relation.predicate)
		}
		appendReflectedInput(model, activities[relation.subject], entities[relation.object])
	case "wasgeneratedby":
		if subjectKind != "entity" || objectKind != "activity" {
			return fmt.Errorf("generator: unsupported endpoint kinds for relation %q", relation.predicate)
		}
		appendReflectedOutput(model, activities[relation.object], entities[relation.subject])
	default:
		return fmt.Errorf("generator: unsupported reflected relation %q", relation.predicate)
	}
	return nil
}
func appendReflectedInput(model *SemanticIR, activityIndex, entityIndex int) {
	entity := model.Entities[entityIndex]
	model.Activities[activityIndex].Inputs = append(model.Activities[activityIndex].Inputs, Port{
		Name: lowerCamel(entity.GoName), EntityID: entity.ID, GoType: entity.GoName,
	})
}
func appendReflectedOutput(model *SemanticIR, activityIndex, entityIndex int) {
	entity := model.Entities[entityIndex]
	model.Activities[activityIndex].Outputs = append(model.Activities[activityIndex].Outputs, Port{
		Name: entity.GoName, EntityID: entity.ID, GoType: entity.GoName,
	})
}
