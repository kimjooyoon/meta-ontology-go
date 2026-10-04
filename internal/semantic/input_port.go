package semantic

import (
	"strconv"
	"strings"
)

// InputPortIndex preserves the legacy single input and gives multiple inputs
// distinct source-position identities, including repeated entity types.
func InputPortIndex(port string, arity int) (int, bool) {
	if arity == 1 {
		return 0, port == RuntimeInputPort
	}
	if arity < 2 || !strings.HasPrefix(port, RuntimeInputPort) {
		return 0, false
	}
	index, err := strconv.Atoi(strings.TrimPrefix(port, RuntimeInputPort))
	return index, err == nil && index >= 0 && index < arity && port == "input"+strconv.Itoa(index)
}

func runtimeInputEntity(graph Graph, activity Node, port string) (ID, bool) {
	inputs := activity.InputSequence
	if len(inputs) == 0 {
		inputs = runtimePortEntities(graph, activity.ID, Used)
		if len(inputs) != 1 {
			return "", false
		}
	}
	index, ok := InputPortIndex(port, len(inputs))
	if !ok {
		return "", false
	}
	actualInputs := runtimePortEntities(graph, activity.ID, Used)
	for _, entity := range inputs {
		node, present := graph.Node(entity)
		if !present || node.Kind != Entity {
			return "", false
		}
		found := false
		for _, actual := range actualInputs {
			found = found || actual == entity
		}
		if !found {
			return "", false
		}
	}
	return inputs[index], true
}
