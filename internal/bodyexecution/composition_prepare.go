package bodyexecution

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"unicode/utf8"
)

func prepareCompositionGraph(ctx context.Context, filename string, source []byte) (compositionGraph, error) {
	var graph compositionGraph
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 || !utf8.Valid(source) {
		return graph, fmt.Errorf("composition requires context and 1..128 KiB UTF-8 source")
	}
	if err := ctx.Err(); err != nil {
		return graph, err
	}
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return graph, fmt.Errorf("composition source: %v", diagnostics)
	}
	for _, binding := range file.Bindings {
		if binding.Feedback {
			return graph, fmt.Errorf("composition requires in-invocation binds; feedback has a separate lifecycle")
		}
	}
	document, err := bidir.DocumentFromSyntaxContext(ctx, file)
	if err != nil {
		return graph, err
	}
	model, err := bidir.Get(document)
	if err != nil {
		return graph, err
	}
	typed, err := bidir.CompileTypedPlan(document)
	if err != nil {
		return graph, err
	}
	if len(typed.Activities) < 2 || len(typed.Activities) > compositionLimit {
		return graph, fmt.Errorf("composition requires 2..16 activities and explicit binds")
	}
	graph.count = len(typed.Activities)
	graph.plan = CompositionPlan{Schema: "gooo/body-composition-plan/v1",
		TypedPlanSHA256: typed.Digest(), SemanticFingerprint: bidir.SemanticFingerprint(model),
		Edges: make([]CompositionEdge, 0, len(typed.Edges))}
	for i, id := range typed.Activities {
		for _, node := range model.Nodes {
			if node.Kind == bidir.ActivityKind && node.ID == id {
				graph.nodes[i] = CompositionActivity{Name: node.Name, ID: string(id), InputFrom: -1,
					GoFunction: fmt.Sprintf("GoooComposedActivity%d", i)}
			}
		}
		if err := graph.bindActivity(file, i); err != nil {
			return graph, err
		}
		for _, entity := range model.Nodes {
			if entity.Kind != bidir.EntityKind {
				continue
			}
			if entity.Name == graph.nodes[i].InputType {
				graph.nodes[i].InputEntityID = string(entity.ID)
			}
			if entity.Name == graph.nodes[i].OutputType {
				graph.nodes[i].OutputEntityID = string(entity.ID)
			}
			for p := range graph.nodes[i].Inputs {
				if entity.Name == graph.nodes[i].Inputs[p].Type {
					graph.nodes[i].Inputs[p].EntityID = string(entity.ID)
				}
			}
		}
	}
	if err := graph.bindEdges(typed); err != nil {
		return graph, err
	}
	graph.plan.Activities = append([]CompositionActivity(nil), graph.nodes[:graph.count]...)
	return graph, nil
}

// Check all bodies and embedded plans before loading an optional model.
func (graph compositionGraph) preflight(ctx context.Context, filename string, source []byte) error {
	for _, node := range graph.nodes[:graph.count] {
		if _, err := bodycodegen.GenerateWithPlanner(ctx, filename, source, node.Name, "", ""); err != nil {
			return fmt.Errorf("activity %q: %w", node.Name, err)
		}
		if node.Assembling {
			if _, err := bodycodegen.DecodeSourcePathDocument(ctx, filename, source, node.Name, nil); err != nil {
				return fmt.Errorf("activity %q plan: %w", node.Name, err)
			}
		}
	}
	return nil
}
