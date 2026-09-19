package bidir

import (
	"fmt"
	"slices"
)

// TypedPlan orders activities using only the canonical explicit runtime bindings.
// A plan is compiler output, not evidence that an activity has executed.
type TypedPlan struct {
	Activities []ID
	Edges      []RuntimeBinding
}

// CompileTypedPlan shares identity, port, type and cycle validation with Get.
// It never interprets entity names as ports or invents missing binding edges.
func CompileTypedPlan(document Document) (TypedPlan, error) {
	model, err := Get(document)
	if err != nil {
		return TypedPlan{}, fmt.Errorf("typed plan: %w", err)
	}
	if len(model.RuntimeBindings) == 0 {
		return TypedPlan{}, fmt.Errorf("typed plan: no explicit binding edges")
	}
	edges := sortedModelRuntimeBindings(model.RuntimeBindings)
	activities, err := typedPlanOrder(model, edges)
	if err != nil {
		return TypedPlan{}, err
	}
	return TypedPlan{Activities: activities, Edges: edges}, nil
}

func typedPlanDependencies(model Model, edges []RuntimeBinding) (map[ID]int, map[ID][]ID) {
	indegree := make(map[ID]int)
	outgoing := make(map[ID][]ID)
	for _, node := range model.Nodes {
		if node.Kind == ActivityKind {
			indegree[node.ID] = 0
		}
	}
	for _, edge := range edges {
		producer, consumer := edge.Producer.Activity.ID, edge.Consumer.Activity.ID
		indegree[consumer]++
		outgoing[producer] = append(outgoing[producer], consumer)
	}
	return indegree, outgoing
}

func typedPlanOrder(model Model, edges []RuntimeBinding) ([]ID, error) {
	indegree, outgoing := typedPlanDependencies(model, edges)
	queue := make([]ID, 0, len(indegree))
	for id, degree := range indegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	order := make([]ID, 0, len(indegree))
	for len(queue) > 0 {
		slices.Sort(queue)
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, next := range outgoing[id] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if len(order) != len(indegree) {
		return nil, fmt.Errorf("typed plan: cycle detected")
	}
	return order, nil
}
