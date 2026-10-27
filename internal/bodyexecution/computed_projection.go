package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// ComputedProjection binds a typed source graph to Go without runtime claims.
type ComputedProjection struct {
	Plan         CompositionPlan
	Source       string
	SourceSHA256 string
}

// GenerateComputedProjection emits authored computes bodies using the ordinary
// typed graph and body emitter. Cases, models and native execution are absent.
func GenerateComputedProjection(ctx context.Context, filename string, source []byte, entry string) (ComputedProjection, error) {
	if entry == "" {
		return ComputedProjection{}, fmt.Errorf("computed projection requires an entry activity")
	}
	graph, err := prepareCompositionGraphForEntry(ctx, filename, source, entry)
	if err != nil {
		return ComputedProjection{}, err
	}
	if graph.hasAssembly() {
		return ComputedProjection{}, fmt.Errorf("computed projection requires authored computes bodies without assembling contracts")
	}
	steps := make([]CompositionStep, 0, graph.count)
	for _, node := range graph.nodes[:graph.count] {
		if err := ctx.Err(); err != nil {
			return ComputedProjection{}, err
		}
		generation, err := bodycodegen.GenerateWithPlanner(ctx, filename, source, node.Name, "", "")
		if err != nil {
			return ComputedProjection{}, fmt.Errorf("computed activity %q: %w", node.ID, err)
		}
		steps = append(steps, CompositionStep{Generation: generation})
	}
	projection, err := emitCompositionProjection(steps, graph)
	if err != nil {
		return ComputedProjection{}, err
	}
	return ComputedProjection{Plan: graph.plan, Source: projection, SourceSHA256: digest(source)}, nil
}
