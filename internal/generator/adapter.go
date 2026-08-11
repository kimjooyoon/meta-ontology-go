package generator

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// SemanticIRProvider is the intentionally small adapter seam for a future
// parser/semantic package.  The generator does not import those packages.
type SemanticIRProvider interface {
	SemanticIR() SemanticIR
}

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

// reflectSemanticGraph is a narrow bridge for early semantic graph packages
// that predate SemanticIRProvider.  It is deliberately structural and only
// reads exported fields named Package, Nodes, Facts, ID, Kind, Name, Subject,
// Predicate, and Object.  Once the semantic package adopts the local provider
// interface, this fallback can be removed without changing the renderer.
func reflectSemanticGraph(input any) (SemanticIR, error) {
	if input == nil {
		return SemanticIR{}, fmt.Errorf("generator: nil semantic input")
	}
	value := reflect.ValueOf(input)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return SemanticIR{}, fmt.Errorf("generator: nil semantic input")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return SemanticIR{}, fmt.Errorf("generator: unsupported semantic input %T", input)
	}
	model := SemanticIR{Package: stringField(value, "Package")}
	nodes := value.FieldByName("Nodes")
	if !nodes.IsValid() || nodes.Kind() != reflect.Map {
		return SemanticIR{}, fmt.Errorf("generator: semantic input %T does not expose a Nodes map; implement SemanticIRProvider", input)
	}
	entityByID := make(map[string]int)
	activityByID := make(map[string]int)
	for _, key := range nodes.MapKeys() {
		node := nodes.MapIndex(key)
		for node.Kind() == reflect.Pointer {
			if node.IsNil() {
				break
			}
			node = node.Elem()
		}
		if node.Kind() != reflect.Struct {
			continue
		}
		id := valueString(node.FieldByName("ID"))
		name := stringField(node, "Name")
		kind := strings.ToLower(valueString(node.FieldByName("Kind")))
		if id == "" || name == "" {
			continue
		}
		switch {
		case strings.Contains(kind, "entity"):
			entityByID[id] = len(model.Entities)
			model.Entities = append(model.Entities, Entity{ID: id, Name: name, GoName: name})
		case strings.Contains(kind, "activity"):
			activityByID[id] = len(model.Activities)
			model.Activities = append(model.Activities, Activity{ID: id, Name: name, GoName: name})
		}
	}
	facts := value.FieldByName("Facts")
	if facts.IsValid() && facts.Kind() == reflect.Map {
		for _, key := range facts.MapKeys() {
			fact := facts.MapIndex(key)
			for fact.Kind() == reflect.Pointer {
				if fact.IsNil() {
					break
				}
				fact = fact.Elem()
			}
			if fact.Kind() != reflect.Struct {
				continue
			}
			subject := valueString(fact.FieldByName("Subject"))
			predicate := strings.ToLower(valueString(fact.FieldByName("Predicate")))
			object := valueString(fact.FieldByName("Object"))
			activityIndex, subjectIsActivity := activityByID[subject]
			objectActivityIndex, objectIsActivity := activityByID[object]
			switch {
			case subjectIsActivity && strings.Contains(predicate, "used"):
				if entityIndex, ok := entityByID[object]; ok {
					entity := model.Entities[entityIndex]
					model.Activities[activityIndex].Inputs = append(model.Activities[activityIndex].Inputs, Port{
						Name: lowerCamel(entity.GoName), EntityID: entity.ID, GoType: entity.GoName,
					})
				}
			case objectIsActivity && strings.Contains(predicate, "generated"):
				if entityIndex, ok := entityByID[subject]; ok {
					entity := model.Entities[entityIndex]
					model.Activities[objectActivityIndex].Outputs = append(model.Activities[objectActivityIndex].Outputs, Port{
						Name: entity.GoName, EntityID: entity.ID, GoType: entity.GoName,
					})
				}
			}
		}
	}
	if model.Package == "" {
		model.Package = "generated"
	}
	for index := range model.Activities {
		sort.SliceStable(model.Activities[index].Inputs, func(left, right int) bool {
			return model.Activities[index].Inputs[left].EntityID < model.Activities[index].Inputs[right].EntityID
		})
		sort.SliceStable(model.Activities[index].Outputs, func(left, right int) bool {
			return model.Activities[index].Outputs[left].EntityID < model.Activities[index].Outputs[right].EntityID
		})
	}
	return model, nil
}

func stringField(value reflect.Value, name string) string {
	return valueString(value.FieldByName(name))
}

func valueString(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}
	if value.Kind() == reflect.String {
		return value.String()
	}
	if value.CanInterface() {
		return fmt.Sprint(value.Interface())
	}
	return ""
}
