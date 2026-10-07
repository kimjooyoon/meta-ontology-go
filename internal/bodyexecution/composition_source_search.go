package bodyexecution

import (
	"context"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

type compositionAssemblyGenerator struct {
	modelPath string
	retained  *bodycodegen.TypedPathGenerator
}

func (g *compositionAssemblyGenerator) generate(ctx context.Context, filename string, source []byte,
	activity string) (bodycodegen.Result, error) {
	spec, err := bodycodegen.SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return bodycodegen.Result{}, err
	}
	if bodycodegen.IsSourceIRSearch(spec) {
		return bodycodegen.GenerateWithSourceIRSearch(ctx, filename, source, activity, spec, "", "")
	}
	if g.retained == nil {
		g.retained, err = bodycodegen.NewTypedPathGenerator(g.modelPath)
		if err != nil {
			return bodycodegen.Result{}, err
		}
	}
	return g.retained.GenerateSourceAssembly(ctx, filename, source, activity)
}

func (graph compositionGraph) validateModelRoute(ctx context.Context, filename string, source []byte, modelPath string) error {
	if modelPath == "" {
		return nil
	}
	for _, node := range graph.nodes[:graph.count] {
		if !node.Assembling {
			continue
		}
		spec, err := bodycodegen.SourceAssembly(ctx, filename, source, node.Name)
		if err != nil {
			return err
		}
		if !bodycodegen.IsSourceIRSearch(spec) {
			return nil
		}
	}
	return fmt.Errorf("composition --model requires a choice-based assembling activity; source IR search currently uses deterministic ordering")
}
