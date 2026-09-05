package bidir

import (
	"fmt"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func lowerRuntimeBindings(model *Model, bindings []RuntimeBinding, names map[string]ID, ids map[ID]struct{}) error {
	for index, binding := range bindings {
		producer, err := resolveReference(binding.Producer.Activity, model.Namespace, names, ids)
		if err != nil {
			return fmt.Errorf("runtime binding %d producer: %w", index, err)
		}
		consumer, err := resolveReference(binding.Consumer.Activity, model.Namespace, names, ids)
		if err != nil {
			return fmt.Errorf("runtime binding %d consumer: %w", index, err)
		}
		producerNode, producerOK := model.node(producer)
		consumerNode, consumerOK := model.node(consumer)
		if !producerOK || producerNode.Kind != ActivityKind {
			return fmt.Errorf("runtime binding %d producer %q is not an activity", index, producer)
		}
		if !consumerOK || consumerNode.Kind != ActivityKind {
			return fmt.Errorf("runtime binding %d consumer %q is not an activity", index, consumer)
		}
		model.RuntimeBindings = append(model.RuntimeBindings, RuntimeBinding{
			Producer: BindingEndpoint{
				Activity: Reference{ID: producer, Name: binding.Producer.Activity.Name, Namespace: model.Namespace, Span: binding.Producer.Activity.Span},
				Port:     binding.Producer.Port,
			},
			Consumer: BindingEndpoint{
				Activity: Reference{ID: consumer, Name: binding.Consumer.Activity.Name, Namespace: model.Namespace, Span: binding.Consumer.Activity.Span},
				Port:     binding.Consumer.Port,
			},
			Span: binding.Span,
		})
		outputs := modelRuntimePorts(*model, producer, PredicateWasGeneratedBy, false)
		if len(outputs) == 1 {
			model.RuntimeBindings[len(model.RuntimeBindings)-1].Entity = outputs[0]
		}
	}
	return nil
}

func validateModelRuntimeBindings(model Model) error {
	seen := make(map[string]struct{}, len(model.RuntimeBindings))
	incoming := make(map[string]struct{}, len(model.RuntimeBindings))
	for index, binding := range model.RuntimeBindings {
		if err := binding.Span.Validate(); err != nil {
			return fmt.Errorf("runtime binding %d: %w", index, err)
		}
		if err := validateID(binding.Producer.Activity.ID); err != nil {
			return fmt.Errorf("runtime binding %d producer: %w", index, err)
		}
		if err := validateID(binding.Consumer.Activity.ID); err != nil {
			return fmt.Errorf("runtime binding %d consumer: %w", index, err)
		}
		producer, producerOK := findModelNode(model.Nodes, binding.Producer.Activity.ID)
		consumer, consumerOK := findModelNode(model.Nodes, binding.Consumer.Activity.ID)
		if !producerOK || producer.Kind != ActivityKind {
			return fmt.Errorf("runtime binding %d: %w: producer %s", index, semantic.ErrRuntimeBindingUnknownNode, binding.Producer.Activity.ID)
		}
		if !consumerOK || consumer.Kind != ActivityKind {
			return fmt.Errorf("runtime binding %d: %w: consumer %s", index, semantic.ErrRuntimeBindingUnknownNode, binding.Consumer.Activity.ID)
		}
		if binding.Producer.Port.Name != semantic.RuntimeOutputPort || binding.Consumer.Port.Name != semantic.RuntimeInputPort {
			return fmt.Errorf("runtime binding %d: %w", index, semantic.ErrRuntimeBindingPort)
		}
		key := runtimeBindingKey(binding)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("runtime binding %d: %w", index, semantic.ErrRuntimeBindingDuplicate)
		}
		seen[key] = struct{}{}
		incomingKey := string(binding.Consumer.Activity.ID) + "\x00" + binding.Consumer.Port.Name
		if _, exists := incoming[incomingKey]; exists {
			return fmt.Errorf("runtime binding %d: %w", index, semantic.ErrRuntimeBindingInputConflict)
		}
		incoming[incomingKey] = struct{}{}
		producerOutputs := modelRuntimePorts(model, binding.Producer.Activity.ID, PredicateWasGeneratedBy, false)
		consumerInputs := modelRuntimePorts(model, binding.Consumer.Activity.ID, PredicateUsed, true)
		if len(producerOutputs) != 1 || len(consumerInputs) != 1 {
			return fmt.Errorf("runtime binding %d: %w: one input/output is required", index, semantic.ErrRuntimeBindingPort)
		}
		if producerOutputs[0] != consumerInputs[0] {
			return fmt.Errorf("runtime binding %d: %w", index, semantic.ErrRuntimeBindingTypeMismatch)
		}
		if binding.Entity != "" && binding.Entity != producerOutputs[0] {
			return fmt.Errorf("runtime binding %d: %w", index, semantic.ErrRuntimeBindingTypeMismatch)
		}
		model.RuntimeBindings[index].Entity = producerOutputs[0]
	}
	if err := validateModelRuntimeBindingAcyclic(model.RuntimeBindings); err != nil {
		return err
	}
	return nil
}

func findModelNode(nodes []Node, id ID) (Node, bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
	}
	return Node{}, false
}

func modelRuntimePorts(model Model, activity ID, predicate Predicate, input bool) []ID {
	result := make([]ID, 0, 1)
	for _, relation := range model.Relations {
		if relation.Kind != predicate {
			continue
		}
		if input && relation.Source == activity {
			result = append(result, relation.Target)
		}
		if !input && relation.Target == activity {
			result = append(result, relation.Source)
		}
	}
	return result
}

func validateModelRuntimeBindingAcyclic(bindings []RuntimeBinding) error {
	indegree := make(map[ID]int, len(bindings)*2)
	outgoing := make(map[ID][]ID, len(bindings))
	for _, binding := range bindings {
		producer, consumer := binding.Producer.Activity.ID, binding.Consumer.Activity.ID
		indegree[producer] = indegree[producer]
		indegree[consumer]++
		outgoing[producer] = append(outgoing[producer], consumer)
	}
	queue := make([]ID, 0, len(indegree))
	for id, degree := range indegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	sort.Slice(queue, func(i, j int) bool { return queue[i] < queue[j] })
	visited := 0
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		visited++
		for _, next := range outgoing[id] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
		sort.Slice(queue, func(i, j int) bool { return queue[i] < queue[j] })
	}
	if visited != len(indegree) {
		return semantic.ErrRuntimeBindingCycle
	}
	return nil
}

func runtimeBindingKey(binding RuntimeBinding) string {
	return string(binding.Producer.Activity.ID) + "\x00" + binding.Producer.Port.Name + "\x00" +
		string(binding.Consumer.Activity.ID) + "\x00" + binding.Consumer.Port.Name
}

func sortedModelRuntimeBindings(bindings []RuntimeBinding) []RuntimeBinding {
	result := append([]RuntimeBinding(nil), bindings...)
	sort.SliceStable(result, func(i, j int) bool { return runtimeBindingKey(result[i]) < runtimeBindingKey(result[j]) })
	return result
}

func sameRuntimeBindings(left, right []RuntimeBinding) bool {
	left, right = sortedModelRuntimeBindings(left), sortedModelRuntimeBindings(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if runtimeBindingKey(left[index]) != runtimeBindingKey(right[index]) || left[index].Entity != right[index].Entity {
			return false
		}
	}
	return true
}
