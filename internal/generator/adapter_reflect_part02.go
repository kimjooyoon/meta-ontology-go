package generator

import (
	"fmt"
	"reflect"
	"strings"
)

func reflectNodes(collection reflectedCollection, model *SemanticIR, entities, activities map[string]int, kinds map[string]string) error {
	for index, value := range collection.values {
		if err := reflectNode(value, index, model, entities, activities, kinds); err != nil {
			return err
		}
	}
	return nil
}
func reflectNode(value reflect.Value, index int, model *SemanticIR, entities, activities map[string]int, kinds map[string]string) error {
	value = indirectValue(value)
	context := fmt.Sprintf("semantic node %d", index)
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return fmt.Errorf("generator: %s must be a struct", context)
	}
	id, err := requiredStringField(value, "ID", context)
	if err != nil {
		return err
	}
	name, err := requiredStringField(value, "Name", context+" "+id)
	if err != nil {
		return err
	}
	kind, err := requiredStringField(value, "Kind", context+" "+id)
	if err != nil {
		return err
	}
	kind = strings.ToLower(kind)
	if kind != "entity" && kind != "activity" {
		return fmt.Errorf("generator: unsupported semantic node kind %q for %q", kind, id)
	}
	if kind == "entity" && reflectedNodeHasFields(value) {
		return entityFieldsError(entityFieldsDeferredDiagnostic, Field{ID: id}, "reflective entity fields are deferred and require the package-private test seam")
	}
	if previous, exists := kinds[id]; exists {
		return fmt.Errorf("generator: duplicate semantic node ID %q (%s and %s)", id, previous, kind)
	}
	kinds[id] = kind
	if kind == "entity" {
		entities[id] = len(model.Entities)
		model.Entities = append(model.Entities, Entity{ID: id, Name: name, GoName: name})
		return nil
	}
	activities[id] = len(model.Activities)
	model.Activities = append(model.Activities, Activity{ID: id, Name: name, GoName: name})
	return nil
}
func reflectedNodeHasFields(value reflect.Value) bool {
	fields := value.FieldByName("Fields")
	if !fields.IsValid() {
		return false
	}
	fields = indirectValue(fields)
	if !fields.IsValid() {
		return false
	}
	switch fields.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return fields.Len() > 0
	default:
		return true
	}
}
