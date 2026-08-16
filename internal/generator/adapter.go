package generator

import (
	"errors"
	"fmt"
	"reflect"
)

// SemanticIRProvider is the intentionally small adapter seam for a future
// parser/semantic package. The generator does not import those packages.
type SemanticIRProvider interface {
	SemanticIR() SemanticIR
}

// ErrDeferredRelationOrder identifies a reflective graph whose relation
// collection has no authoritative order. Typed adapters must provide the
// ordered SemanticIR when port order is semantically meaningful.
var ErrDeferredRelationOrder = errors.New("generator: relation order is DEFERRED/non-authoritative")

func adaptInput(input any) (SemanticIR, error) {
	switch value := input.(type) {
	case SemanticIR:
		return value, nil
	case *SemanticIR:
		if value == nil {
			return SemanticIR{}, fmt.Errorf("generator: nil SemanticIR")
		}
		return *value, nil
	case SemanticIRProvider:
		return value.SemanticIR(), nil
	default:
		return reflectSemanticGraph(input)
	}
}

// reflectSemanticGraph is a strict compatibility bridge for structural
// semantic graphs. It rejects malformed input rather than dropping facts.
func reflectSemanticGraph(input any) (SemanticIR, error) {
	value, err := reflectedStruct(input)
	if err != nil {
		return SemanticIR{}, err
	}
	model := SemanticIR{Package: "generated"}
	if packageName, err := optionalStringField(value, "Package"); err != nil {
		return SemanticIR{}, err
	} else if packageName != "" {
		model.Package = packageName
	}
	nodes, err := readReflectedCollection(value, "Nodes", input)
	if err != nil {
		return SemanticIR{}, err
	}
	entities := make(map[string]int)
	activities := make(map[string]int)
	kinds := make(map[string]string)
	if err := reflectNodes(nodes, &model, entities, activities, kinds); err != nil {
		return SemanticIR{}, err
	}
	facts, err := readReflectedCollection(value, "Facts", input)
	if err != nil {
		return SemanticIR{}, err
	}
	if err := reflectFacts(facts, &model, entities, activities, kinds); err != nil {
		return SemanticIR{}, err
	}
	if !facts.ordered && len(facts.values) > 0 {
		return SemanticIR{}, fmt.Errorf("%w: reflective Facts map has no source order; implement SemanticIRProvider", ErrDeferredRelationOrder)
	}
	return model, nil
}

func reflectedStruct(input any) (reflect.Value, error) {
	if input == nil {
		return reflect.Value{}, fmt.Errorf("generator: nil semantic input")
	}
	value := indirectValue(reflect.ValueOf(input))
	if !value.IsValid() {
		return reflect.Value{}, fmt.Errorf("generator: nil semantic input")
	}
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("generator: unsupported semantic input %T", input)
	}
	return value, nil
}

func indirectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func valueString(value reflect.Value) string {
	value = indirectValue(value)
	if !value.IsValid() {
		return ""
	}
	if value.Kind() == reflect.String {
		return value.String()
	}
	if value.CanInterface() {
		return fmt.Sprint(value.Interface())
	}
	return ""
}
