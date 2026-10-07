package bodyexecution

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func prepareCompositionGraph(ctx context.Context, filename string, source []byte) (compositionGraph, error) {
	return prepareCompositionGraphForEntry(ctx, filename, source, "")
}

func prepareCompositionGraphForEntry(ctx context.Context, filename string, source []byte, entry string) (compositionGraph, error) {
	var graph compositionGraph
	file, err := compositionSource(ctx, filename, source)
	if err != nil {
		return graph, err
	}
	support := bodycodegen.BodyEntityFieldsSupport(file)
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, support)
	if err != nil {
		return graph, err
	}
	model, err := bidir.GetWithEntityFieldsSupport(document, support)
	if err != nil {
		return graph, err
	}
	typed, err := compileCompositionEntry(document, model, entry)
	if err != nil {
		return graph, err
	}
	if len(typed.Activities) < 1 || len(typed.Activities) > compositionLimit {
		return graph, fmt.Errorf("composition requires 1..16 declared activities")
	}
	graph.count = len(typed.Activities)
	graph.plan = CompositionPlan{Schema: "gooo/body-composition-plan/v1",
		EntryActivity:   entry,
		TypedPlanSHA256: typed.Digest(), SemanticFingerprint: bidir.SemanticFingerprint(model),
		Edges: make([]CompositionEdge, 0, len(typed.Edges))}
	graph.plan.Records, err = bodycodegen.RecordTypesFromModel(model)
	if err != nil {
		return graph, err
	}
	if err := graph.bindNodes(file, model, typed); err != nil {
		return graph, err
	}
	if err := graph.bindEdges(typed); err != nil {
		return graph, err
	}
	if err := graph.bindCallConstruction(ctx, filename, source, model); err != nil {
		return graph, err
	}
	graph.plan.Activities = append([]CompositionActivity(nil), graph.nodes[:graph.count]...)
	return graph, nil
}

func compositionSource(ctx context.Context, filename string, source []byte) (*syntax.File, error) {
	if ctx == nil || len(source) == 0 || len(source) > 128<<10 || !utf8.Valid(source) {
		return nil, fmt.Errorf("composition requires context and 1..128 KiB UTF-8 source")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, diagnostics := bodycodegen.ParseBodyFile(filename, source)
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("composition source: %v", diagnostics)
	}
	for _, binding := range file.Bindings {
		if binding.Feedback {
			return nil, fmt.Errorf("composition requires in-invocation binds; feedback has a separate lifecycle")
		}
	}
	return file, nil
}

func (graph *compositionGraph) bindNodes(file *syntax.File, model bidir.Model, typed bidir.TypedPlan) error {
	for i, id := range typed.Activities {
		for _, node := range model.Nodes {
			if node.Kind == bidir.ActivityKind && node.ID == id {
				graph.nodes[i] = CompositionActivity{Name: node.Name, ID: string(id), InputFrom: -1,
					GoFunction: fmt.Sprintf("GoooComposedActivity%d", i)}
			}
		}
		if err := graph.bindActivity(file, i); err != nil {
			return err
		}
		graph.nodes[i].bindEntityIDs(model)
	}
	return nil
}

func (node *CompositionActivity) bindEntityIDs(model bidir.Model) {
	for _, entity := range model.Nodes {
		if entity.Kind != bidir.EntityKind {
			continue
		}
		if entity.Name == node.InputType {
			node.InputEntityID = string(entity.ID)
		}
		if entity.Name == node.OutputType {
			node.OutputEntityID = string(entity.ID)
		}
		for p := range node.Inputs {
			if entity.Name == node.Inputs[p].Type {
				node.Inputs[p].EntityID = string(entity.ID)
			}
		}
	}
}

// Check independent bodies before loading an optional model. Dependent bodies
// are checked again when their called assembly prerequisites have been realized.
func (graph compositionGraph) preflight(ctx context.Context, filename string, source []byte) error {
	for _, node := range graph.nodes[:graph.count] {
		if graph.deferred[node.Name] {
			continue
		}
		if node.Assembling {
			if err := bodycodegen.ValidateSourceAssembly(ctx, filename, source, node.Name); err != nil {
				return fmt.Errorf("activity %q plan: %w", node.Name, err)
			}
			continue
		}
		if _, err := bodycodegen.GenerateWithPlanner(ctx, filename, source, node.Name, "", ""); err != nil {
			return fmt.Errorf("activity %q: %w", node.Name, err)
		}
	}
	for _, helper := range graph.plan.Preparations {
		if !graph.deferred[helper.Name] {
			if err := bodycodegen.ValidateSourceAssembly(ctx, filename, source, helper.Name); err != nil {
				return fmt.Errorf("called activity %q plan: %w", helper.Name, err)
			}
		}
	}
	return nil
}
