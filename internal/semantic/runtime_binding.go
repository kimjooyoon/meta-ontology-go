package semantic

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	RuntimeBindingSchema = "gooo.runtime-binding/v1"
	RuntimeOutputPort    = "result"
	RuntimeInputPort     = "input"
)

var (
	ErrRuntimeBindingInvalid       = errors.New("invalid runtime binding")
	ErrRuntimeBindingUnknownNode   = errors.New("runtime binding references unknown node")
	ErrRuntimeBindingPort          = errors.New("runtime binding port is unsupported")
	ErrRuntimeBindingDuplicate     = errors.New("duplicate runtime binding")
	ErrRuntimeBindingInputConflict = errors.New("runtime binding input already bound")
	ErrRuntimeBindingTypeMismatch  = errors.New("runtime binding type mismatch")
	ErrRuntimeBindingCycle          = errors.New("runtime binding cycle")
)

// RuntimeBinding is an explicit, port-sensitive execution edge. It is a
// typed semantic extension, not a PROV fact: entity-level PROV relations do
// not identify a producer result instance or a consumer port.
type RuntimeBinding struct {
	Schema           string `json:"schema"`
	ProducerActivity ID     `json:"producer_activity"`
	ProducerPort     string `json:"producer_port"`
	ConsumerActivity ID     `json:"consumer_activity"`
	ConsumerPort     string `json:"consumer_port"`
	Entity           ID     `json:"entity"`
	Span             Span   `json:"span,omitempty"`
}

type RuntimeBindingKey struct {
	ProducerActivity ID
	ProducerPort     string
	ConsumerActivity ID
	ConsumerPort     string
}

func (b RuntimeBinding) Key() RuntimeBindingKey {
	return RuntimeBindingKey{
		ProducerActivity: b.ProducerActivity,
		ProducerPort:     b.ProducerPort,
		ConsumerActivity: b.ConsumerActivity,
		ConsumerPort:     b.ConsumerPort,
	}
}

func (b RuntimeBinding) Normalized() (RuntimeBinding, error) {
	if strings.TrimSpace(b.Schema) == "" {
		b.Schema = RuntimeBindingSchema
	}
	if b.Schema != RuntimeBindingSchema {
		return RuntimeBinding{}, fmt.Errorf("%w: unsupported schema %q", ErrRuntimeBindingInvalid, b.Schema)
	}
	producer, err := ParseIdentity(b.ProducerActivity.String())
	if err != nil {
		return RuntimeBinding{}, fmt.Errorf("%w: producer activity: %v", ErrRuntimeBindingInvalid, err)
	}
	consumer, err := ParseIdentity(b.ConsumerActivity.String())
	if err != nil {
		return RuntimeBinding{}, fmt.Errorf("%w: consumer activity: %v", ErrRuntimeBindingInvalid, err)
	}
	entity := b.Entity
	if entity != "" {
		entity, err = ParseIdentity(entity.String())
		if err != nil {
			return RuntimeBinding{}, fmt.Errorf("%w: entity: %v", ErrRuntimeBindingInvalid, err)
		}
	}
	span := b.Span.Normalized()
	if err := span.Validate(); err != nil {
		return RuntimeBinding{}, fmt.Errorf("%w: span: %v", ErrRuntimeBindingInvalid, err)
	}
	b.ProducerActivity = producer
	b.ConsumerActivity = consumer
	b.ProducerPort = strings.TrimSpace(b.ProducerPort)
	b.ConsumerPort = strings.TrimSpace(b.ConsumerPort)
	b.Entity = entity
	b.Span = span
	return b, nil
}

func (b RuntimeBinding) Canonical() string {
	if normalized, err := b.Normalized(); err == nil {
		b = normalized
	}
	var builder strings.Builder
	builder.WriteString("runtime-binding\t")
	writeCanonicalField(&builder, b.Schema)
	writeCanonicalField(&builder, b.ProducerActivity.String())
	writeCanonicalField(&builder, b.ProducerPort)
	writeCanonicalField(&builder, b.ConsumerActivity.String())
	writeCanonicalField(&builder, b.ConsumerPort)
	writeCanonicalField(&builder, b.Entity.String())
	writeCanonicalSpan(&builder, b.Span)
	return builder.String()
}

func (b RuntimeBinding) SemanticCanonical() string {
	if normalized, err := b.Normalized(); err == nil {
		b = normalized
	}
	var builder strings.Builder
	builder.WriteString("runtime-binding\t")
	writeCanonicalField(&builder, b.ProducerActivity.String())
	writeCanonicalField(&builder, b.ProducerPort)
	writeCanonicalField(&builder, b.ConsumerActivity.String())
	writeCanonicalField(&builder, b.ConsumerPort)
	writeCanonicalField(&builder, b.Entity.String())
	return builder.String()
}

func normalizeRuntimeBindings(bindings []RuntimeBinding, graph Graph) ([]RuntimeBinding, error) {
	if len(bindings) == 0 {
		return nil, nil
	}
	result := make([]RuntimeBinding, len(bindings))
	seen := make(map[RuntimeBindingKey]struct{}, len(bindings))
	incoming := make(map[string]struct{}, len(bindings))
	for index, raw := range bindings {
		binding, err := raw.Normalized()
		if err != nil {
			return nil, fmt.Errorf("runtime binding %d: %w", index, err)
		}
		producer, producerOK := graph.Node(binding.ProducerActivity)
		consumer, consumerOK := graph.Node(binding.ConsumerActivity)
		if !producerOK || producer.Kind != Activity {
			return nil, fmt.Errorf("%w: producer %s", ErrRuntimeBindingUnknownNode, binding.ProducerActivity)
		}
		if !consumerOK || consumer.Kind != Activity {
			return nil, fmt.Errorf("%w: consumer %s", ErrRuntimeBindingUnknownNode, binding.ConsumerActivity)
		}
		if binding.ProducerPort != RuntimeOutputPort || binding.ConsumerPort != RuntimeInputPort {
			return nil, fmt.Errorf("%w: %s -> %s", ErrRuntimeBindingPort, binding.ProducerPort, binding.ConsumerPort)
		}
		key := binding.Key()
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("%w: %v", ErrRuntimeBindingDuplicate, key)
		}
		seen[key] = struct{}{}
		incomingKey := binding.ConsumerActivity.String() + "\x00" + binding.ConsumerPort
		if _, exists := incoming[incomingKey]; exists {
			return nil, fmt.Errorf("%w: %s.%s", ErrRuntimeBindingInputConflict, binding.ConsumerActivity, binding.ConsumerPort)
		}
		incoming[incomingKey] = struct{}{}
		producerOutputs := runtimePortEntities(graph, binding.ProducerActivity, WasGeneratedBy)
		consumerInputs := runtimePortEntities(graph, binding.ConsumerActivity, Used)
		if len(producerOutputs) != 1 || len(consumerInputs) != 1 {
			return nil, fmt.Errorf("%w: runtime bindings require one input and one output per activity", ErrRuntimeBindingPort)
		}
		if producerOutputs[0] != consumerInputs[0] {
			return nil, fmt.Errorf("%w: producer %s and consumer %s", ErrRuntimeBindingTypeMismatch, producerOutputs[0], consumerInputs[0])
		}
		if binding.Entity != "" && binding.Entity != producerOutputs[0] {
			return nil, fmt.Errorf("%w: declared entity %s, producer entity %s", ErrRuntimeBindingTypeMismatch, binding.Entity, producerOutputs[0])
		}
		binding.Entity = producerOutputs[0]
		result[index] = binding
	}
	if err := validateRuntimeBindingAcyclic(result); err != nil {
		return nil, err
	}
	return result, nil
}

func runtimePortEntities(graph Graph, activity ID, predicate Relation) []ID {
	result := make([]ID, 0, 1)
	for _, fact := range graph.DeterministicFacts() {
		switch {
		case predicate == Used && fact.Predicate == Used && fact.Subject == activity:
			result = append(result, fact.Object)
		case predicate == WasGeneratedBy && fact.Predicate == WasGeneratedBy && fact.Object == activity:
			result = append(result, fact.Subject)
		}
	}
	return result
}

func validateRuntimeBindingAcyclic(bindings []RuntimeBinding) error {
	if len(bindings) == 0 {
		return nil
	}
	indegree := make(map[ID]int, len(bindings)*2)
	outgoing := make(map[ID][]ID, len(bindings))
	for _, binding := range bindings {
		producer, consumer := binding.ProducerActivity, binding.ConsumerActivity
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
		return ErrRuntimeBindingCycle
	}
	return nil
}

func sortedRuntimeBindings(bindings []RuntimeBinding) []RuntimeBinding {
	result := append([]RuntimeBinding(nil), bindings...)
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i].Key(), result[j].Key()
		if left.ProducerActivity != right.ProducerActivity {
			return left.ProducerActivity < right.ProducerActivity
		}
		if left.ProducerPort != right.ProducerPort {
			return left.ProducerPort < right.ProducerPort
		}
		if left.ConsumerActivity != right.ConsumerActivity {
			return left.ConsumerActivity < right.ConsumerActivity
		}
		return left.ConsumerPort < right.ConsumerPort
	})
	return result
}
