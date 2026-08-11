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
	value, err := reflectedStruct(input)
	if err != nil {
		return SemanticIR{}, err
	}
	model := SemanticIR{Package: stringField(value, "Package")}
	entityByID, activityByID, err := reflectNodes(value, &model, input)
	if err != nil {
		return SemanticIR{}, err
	}
	reflectFacts(value, &model, entityByID, activityByID)
	finishReflectedModel(&model)
	return model, nil
}

func reflectedStruct(input any) (reflect.Value, error) {
	if input == nil {
		return reflect.Value{}, fmt.Errorf("generator: nil semantic input")
	}
	value := reflect.ValueOf(input)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, fmt.Errorf("generator: nil semantic input")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("generator: unsupported semantic input %T", input)
	}
	return value, nil
}

func reflectNodes(value reflect.Value, model *SemanticIR, input any) (map[string]int, map[string]int, error) {
	nodes := value.FieldByName("Nodes")
	if !nodes.IsValid() || nodes.Kind() != reflect.Map {
		return nil, nil, fmt.Errorf("generator: semantic input %T does not expose a Nodes map; implement SemanticIRProvider", input)
	}
	entityByID := make(map[string]int)
	activityByID := make(map[string]int)
	for _, key := range nodes.MapKeys() {
		reflectNode(nodes.MapIndex(key), model, entityByID, activityByID)
	}
	return entityByID, activityByID, nil
}

func reflectNode(node reflect.Value, model *SemanticIR, entities, activities map[string]int) {
	node = indirectValue(node)
	if node.Kind() != reflect.Struct {
		return
	}
	id := valueString(node.FieldByName("ID"))
	name := stringField(node, "Name")
	kind := strings.ToLower(valueString(node.FieldByName("Kind")))
	if id == "" || name == "" {
		return
	}
	switch {
	case strings.Contains(kind, "entity"):
		entities[id] = len(model.Entities)
		model.Entities = append(model.Entities, Entity{ID: id, Name: name, GoName: name})
	case strings.Contains(kind, "activity"):
		activities[id] = len(model.Activities)
		model.Activities = append(model.Activities, Activity{ID: id, Name: name, GoName: name})
	}
}

func reflectFacts(value reflect.Value, model *SemanticIR, entities, activities map[string]int) {
	facts := value.FieldByName("Facts")
	if !facts.IsValid() || facts.Kind() != reflect.Map {
		return
	}
	for _, key := range facts.MapKeys() {
		reflectFact(facts.MapIndex(key), model, entities, activities)
	}
}

func reflectFact(fact reflect.Value, model *SemanticIR, entities, activities map[string]int) {
	fact = indirectValue(fact)
	if fact.Kind() != reflect.Struct {
		return
	}
	subject := valueString(fact.FieldByName("Subject"))
	predicate := strings.ToLower(valueString(fact.FieldByName("Predicate")))
	object := valueString(fact.FieldByName("Object"))
	activityIndex, subjectIsActivity := activities[subject]
	objectActivityIndex, objectIsActivity := activities[object]
	if subjectIsActivity && strings.Contains(predicate, "used") {
		appendInput(model, activityIndex, entities, object)
	}
	if objectIsActivity && strings.Contains(predicate, "generated") {
		appendOutput(model, objectActivityIndex, entities, subject)
	}
}

func appendInput(model *SemanticIR, activityIndex int, entities map[string]int, entityID string) {
	entityIndex, ok := entities[entityID]
	if !ok {
		return
	}
	entity := model.Entities[entityIndex]
	model.Activities[activityIndex].Inputs = append(model.Activities[activityIndex].Inputs, Port{
		Name: lowerCamel(entity.GoName), EntityID: entity.ID, GoType: entity.GoName,
	})
}

func appendOutput(model *SemanticIR, activityIndex int, entities map[string]int, entityID string) {
	entityIndex, ok := entities[entityID]
	if !ok {
		return
	}
	entity := model.Entities[entityIndex]
	model.Activities[activityIndex].Outputs = append(model.Activities[activityIndex].Outputs, Port{
		Name: entity.GoName, EntityID: entity.ID, GoType: entity.GoName,
	})
}

func finishReflectedModel(model *SemanticIR) {
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
}

func indirectValue(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
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
