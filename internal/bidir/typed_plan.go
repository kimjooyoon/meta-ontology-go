package bidir

import (
	"fmt"
	"slices"
	"sort"
)

// BindingEdge is an explicit typed data-flow edge between activity ports.
// It is kept separate from semantic relations because execution requires
// port identity and type checking, not only activity adjacency.
type BindingEdge struct {
	SourceActivity ID
	SourcePort     string
	TargetActivity ID
	TargetPort     string
}

// TypedPlan is the validated, deterministic execution order for binding edges.
// It contains no inferred edges: callers must provide every edge explicitly.
type TypedPlan struct {
	Activities []ID
	Edges      []BindingEdge
}

// CompileTypedPlan validates explicit binding edges and returns a canonical
// topological plan. Unknown ports, duplicate edges, cycles, and ambiguous
// activity references fail closed rather than being inferred.
func CompileTypedPlan(document Document) (TypedPlan, error) {
	activities := make(map[ID]Declaration)
	for _, declaration := range document.Declarations {
		if declaration.Kind != ActivityKind {
			continue
		}
		id := declaration.ID
		if id == "" {
			canonical, err := declarationIdentity(document.Namespace, declaration)
			if err != nil {
				return TypedPlan{}, fmt.Errorf("typed plan: activity %q: %w", declaration.Name, err)
			}
			id = canonical
		}
		if _, exists := activities[id]; exists {
			return TypedPlan{}, fmt.Errorf("typed plan: duplicate activity %q", id)
		}
		activities[id] = declaration
	}
	if len(document.BindingEdges) == 0 {
		return TypedPlan{}, fmt.Errorf("typed plan: no explicit binding edges")
	}

	plan := TypedPlan{Edges: append([]BindingEdge(nil), document.BindingEdges...)}
	seen := make(map[string]struct{}, len(plan.Edges))
	indegree := make(map[ID]int, len(activities))
	adjacency := make(map[ID][]ID, len(activities))
	for id := range activities {
		indegree[id] = 0
	}
	for _, edge := range plan.Edges {
		source, sourceOK := activities[edge.SourceActivity]
		target, targetOK := activities[edge.TargetActivity]
		if !sourceOK || !targetOK {
			return TypedPlan{}, fmt.Errorf("typed plan: unknown activity in edge %q -> %q", edge.SourceActivity, edge.TargetActivity)
		}
		key := fmt.Sprintf("%s:%s->%s:%s", edge.SourceActivity, edge.SourcePort, edge.TargetActivity, edge.TargetPort)
		if _, exists := seen[key]; exists {
			return TypedPlan{}, fmt.Errorf("typed plan: duplicate edge %q", key)
		}
		seen[key] = struct{}{}
		sourceType, ok := activityPortType(source, edge.SourcePort, false)
		if !ok {
			return TypedPlan{}, fmt.Errorf("typed plan: unknown source port %q", edge.SourcePort)
		}
		targetType, ok := activityPortType(target, edge.TargetPort, true)
		if !ok {
			return TypedPlan{}, fmt.Errorf("typed plan: unknown target port %q", edge.TargetPort)
		}
		if sourceType != targetType {
			return TypedPlan{}, fmt.Errorf("typed plan: port type mismatch %q != %q", sourceType, targetType)
		}
		adjacency[edge.SourceActivity] = append(adjacency[edge.SourceActivity], edge.TargetActivity)
		indegree[edge.TargetActivity]++
	}
	queue := make([]ID, 0, len(activities))
	for id, degree := range indegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	for len(queue) > 0 {
		slices.Sort(queue)
		id := queue[0]
		queue = queue[1:]
		plan.Activities = append(plan.Activities, id)
		for _, next := range adjacency[id] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if len(plan.Activities) != len(activities) {
		return TypedPlan{}, fmt.Errorf("typed plan: cycle detected")
	}
	sort.Slice(plan.Edges, func(i, j int) bool {
		return fmt.Sprintf("%s:%s->%s:%s", plan.Edges[i].SourceActivity, plan.Edges[i].SourcePort, plan.Edges[i].TargetActivity, plan.Edges[i].TargetPort) < fmt.Sprintf("%s:%s->%s:%s", plan.Edges[j].SourceActivity, plan.Edges[j].SourcePort, plan.Edges[j].TargetActivity, plan.Edges[j].TargetPort)
	})
	return plan, nil
}

func activityPortType(declaration Declaration, port string, input bool) (ID, bool) {
	references := declaration.Outputs
	if input {
		references = declaration.Inputs
	}
	if len(references) == 1 {
		if input && port == "input" {
			return references[0].ID, true
		}
		if !input && port == "result" {
			return references[0].ID, true
		}
	}
	for _, reference := range references {
		if reference.Name == port {
			return reference.ID, true
		}
	}
	return "", false
}
