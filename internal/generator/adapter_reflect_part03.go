package generator

import (
	"fmt"
	"reflect"
	"strings"
)

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
