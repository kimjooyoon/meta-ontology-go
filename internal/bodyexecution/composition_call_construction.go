package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func (graph *compositionGraph) bindCallConstruction(ctx context.Context, filename string, source []byte, model bidir.Model) error {
	roots := make([]string, graph.count)
	for i, node := range graph.nodes[:graph.count] {
		roots[i] = node.Name
	}
	plan, err := bodycodegen.PlanCallConstruction(ctx, filename, source, roots)
	if err != nil {
		return err
	}
	graph.deferred = make(map[string]bool, len(plan.Deferred))
	for _, name := range plan.Deferred {
		graph.deferred[name] = true
	}
	for _, name := range plan.Helpers {
		for _, node := range model.Nodes {
			if node.Kind == bidir.ActivityKind && node.Name == name {
				graph.plan.Preparations = append(graph.plan.Preparations, CompositionHelper{Name: name, ID: string(node.ID)})
			}
		}
		for i := range graph.count {
			if graph.nodes[i].Name == name {
				graph.nodes[i].Prepared = true
			}
		}
	}
	return nil
}

func generateCalledBodies(ctx context.Context, filename string, source []byte, graph compositionGraph,
	generator *compositionAssemblyGenerator, result *Composition) ([]byte, error) {
	current := source
	for _, helper := range graph.plan.Preparations {
		result.ActiveActivity = helper.Name
		generation, err := generator.generate(ctx, filename, current, helper.Name)
		if err != nil {
			return nil, fmt.Errorf("called activity %q generation: %w", helper.Name, err)
		}
		result.Preparations = append(result.Preparations, CompositionStep{InputSourceSHA256: digest(current), Generation: generation})
		result.FillModel = generator.fillInfo
		if generator.retained != nil {
			info := generator.retained.Info()
			result.Model = &info
		}
		realized, err := bodycodegen.RealizeCalledAssembly(ctx, filename, current, generation)
		if err != nil {
			return nil, fmt.Errorf("called activity %q checkpoint: %w", helper.Name, err)
		}
		current = []byte(realized.Source)
	}
	return current, nil
}

func replayCalledBodies(ctx context.Context, filename string, source []byte, graph compositionGraph, prior Composition) ([]byte, error) {
	if len(prior.Preparations) != len(graph.plan.Preparations) {
		return nil, fmt.Errorf("called activity construction count differs")
	}
	current := source
	for i, helper := range graph.plan.Preparations {
		step := prior.Preparations[i]
		if step.InputSourceSHA256 != digest(current) || step.Generation.Report.Activity != helper.Name ||
			step.Generation.Report.ActivityID != helper.ID {
			return nil, fmt.Errorf("called activity construction %d source, identity or order differs", i)
		}
		realized, err := bodycodegen.RealizeCalledAssembly(ctx, filename, current, step.Generation)
		if err != nil {
			return nil, fmt.Errorf("called activity %q replay: %w", helper.Name, err)
		}
		current = []byte(realized.Source)
	}
	return current, nil
}

// ConstructionSteps returns saved construction in source-checkpoint order.
// Preparations build callable bodies; Steps project independently run activities.
func (c Composition) ConstructionSteps() []CompositionStep {
	steps := make([]CompositionStep, 0, len(c.Preparations)+len(c.Steps))
	steps = append(steps, c.Preparations...)
	return append(steps, c.Steps...)
}
