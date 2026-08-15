package generator

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type reflectedCollection struct {
	values  []reflect.Value
	ordered bool
}

func readReflectedCollection(value reflect.Value, name string, input any) (reflectedCollection, error) {
	field := value.FieldByName(name)
	if !field.IsValid() {
		return reflectedCollection{}, fmt.Errorf("generator: semantic input %T does not expose %s; implement SemanticIRProvider", input, name)
	}
	field = indirectValue(field)
	if !field.IsValid() {
		return reflectedCollection{}, fmt.Errorf("generator: semantic input %T has nil %s", input, name)
	}
	switch field.Kind() {
	case reflect.Array, reflect.Slice:
		values := make([]reflect.Value, field.Len())
		for index := range values {
			values[index] = field.Index(index)
		}
		return reflectedCollection{values: values, ordered: true}, nil
	case reflect.Map:
		keys := field.MapKeys()
		sort.SliceStable(keys, func(left, right int) bool {
			return reflectedKey(keys[left]) < reflectedKey(keys[right])
		})
		values := make([]reflect.Value, 0, len(keys))
		for _, key := range keys {
			values = append(values, field.MapIndex(key))
		}
		return reflectedCollection{values: values}, nil
	default:
		return reflectedCollection{}, fmt.Errorf("generator: semantic input %T has unsupported %s collection %s", input, name, field.Kind())
	}
}

func reflectedKey(value reflect.Value) string {
	return value.Kind().String() + ":" + valueString(value)
}

func optionalStringField(value reflect.Value, name string) (string, error) {
	field := value.FieldByName(name)
	if !field.IsValid() {
		return "", nil
	}
	field = indirectValue(field)
	if !field.IsValid() || field.Kind() != reflect.String {
		return "", fmt.Errorf("generator: semantic field %s must be a string", name)
	}
	return strings.TrimSpace(field.String()), nil
}

func requiredStringField(value reflect.Value, name, context string) (string, error) {
	field := value.FieldByName(name)
	if !field.IsValid() {
		return "", fmt.Errorf("generator: %s is missing %s", context, name)
	}
	field = indirectValue(field)
	if !field.IsValid() || field.Kind() != reflect.String {
		return "", fmt.Errorf("generator: %s field %s must be a string", context, name)
	}
	result := strings.TrimSpace(field.String())
	if result == "" {
		return "", fmt.Errorf("generator: %s has empty %s", context, name)
	}
	return result, nil
}

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

func reflectFacts(collection reflectedCollection, model *SemanticIR, entities, activities map[string]int, kinds map[string]string) error {
	seen := make(map[string]struct{}, len(collection.values))
	for index, value := range collection.values {
		relation, err := reflectFact(value, index)
		if err != nil {
			return err
		}
		key := relation.subject + "\x00" + relation.predicate + "\x00" + relation.object
		if _, exists := seen[key]; exists {
			return fmt.Errorf("generator: duplicate reflected relation %q", key)
		}
		seen[key] = struct{}{}
		if err := appendReflectedRelation(relation, model, entities, activities, kinds); err != nil {
			return err
		}
	}
	return nil
}

type reflectedRelation struct {
	subject   string
	predicate string
	object    string
}

func reflectFact(value reflect.Value, index int) (reflectedRelation, error) {
	value = indirectValue(value)
	context := fmt.Sprintf("semantic fact %d", index)
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return reflectedRelation{}, fmt.Errorf("generator: %s must be a struct", context)
	}
	subject, err := requiredStringField(value, "Subject", context)
	if err != nil {
		return reflectedRelation{}, err
	}
	predicate, err := requiredStringField(value, "Predicate", context)
	if err != nil {
		return reflectedRelation{}, err
	}
	object, err := requiredStringField(value, "Object", context)
	if err != nil {
		return reflectedRelation{}, err
	}
	return reflectedRelation{subject: subject, predicate: strings.ToLower(predicate), object: object}, nil
}

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
