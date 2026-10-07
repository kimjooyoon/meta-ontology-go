package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

type compositionAssemblyGenerator struct {
	options      CompositionOptions
	retained     *bodycodegen.TypedPathGenerator
	fillProvider *decisionroute.TinyGoProvider
	fillInfo     *CompositionFillModelInfo
}

func (g *compositionAssemblyGenerator) generate(ctx context.Context, filename string, source []byte,
	activity string) (bodycodegen.Result, error) {
	spec, err := bodycodegen.SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return bodycodegen.Result{}, err
	}
	if err := bodycodegen.ValidateSourceAssembly(ctx, filename, source, activity); err != nil {
		return bodycodegen.Result{}, err
	}
	if bodycodegen.IsSourceIRBodyFill(spec) {
		return g.generateFill(ctx, filename, source, activity, spec)
	}
	if bodycodegen.IsSourceIRSearch(spec) {
		return bodycodegen.GenerateWithSourceIRSearch(ctx, filename, source, activity, spec, "", "")
	}
	if g.retained == nil {
		g.retained, err = bodycodegen.NewTypedPathGenerator(g.options.ModelPath)
		if err != nil {
			return bodycodegen.Result{}, err
		}
	}
	if g.options.RecordPolicy != nil && bodycodegen.IsRecordAssembly(spec) {
		return g.retained.GenerateRecordAssemblyWithPolicy(ctx, filename, source, activity, *g.options.RecordPolicy)
	}
	return g.retained.GenerateSourceAssembly(ctx, filename, source, activity)
}

func (graph compositionGraph) validateModelRoute(ctx context.Context, filename string, source []byte, options CompositionOptions) error {
	if options.ModelPath == "" && options.FillModelPath == "" && options.RecordPolicy == nil {
		return nil
	}
	if options.RecordPolicy != nil {
		if err := bodycodegen.ValidateRecordAssemblyPolicy(ctx, *options.RecordPolicy); err != nil {
			return err
		}
	}
	var hasChoice, hasFill, hasRecord bool
	nodes := append([]CompositionActivity(nil), graph.nodes[:graph.count]...)
	for _, helper := range graph.plan.Preparations {
		nodes = append(nodes, CompositionActivity{Name: helper.Name, Assembling: true})
	}
	for _, node := range nodes {
		if !node.Assembling {
			continue
		}
		spec, err := bodycodegen.SourceAssembly(ctx, filename, source, node.Name)
		if err != nil {
			return err
		}
		hasRecord = hasRecord || bodycodegen.IsRecordAssembly(spec)
		if bodycodegen.IsSourceIRBodyFill(spec) {
			hasFill = true
		} else if !bodycodegen.IsSourceIRSearch(spec) {
			hasChoice = true
		}
	}
	if options.ModelPath != "" && !hasChoice {
		return fmt.Errorf("composition --model requires a choice-based assembling activity; source IR search currently uses deterministic ordering")
	}
	if options.FillModelPath != "" && !hasFill {
		return fmt.Errorf("composition --fill-model requires a source_fill assembling activity")
	}
	if options.RecordPolicy != nil && !hasRecord {
		return fmt.Errorf("composition --assembly-policy requires a record-choice assembling activity")
	}
	return nil
}
