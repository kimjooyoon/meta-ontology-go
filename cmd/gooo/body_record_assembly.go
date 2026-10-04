package main

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func generateRecordAssemblyCLI(ctx context.Context, filename string, source []byte, activity, modelPath string) (bodycodegen.Result, error) {
	if err := bodycodegen.ValidateSourceAssembly(ctx, filename, source, activity); err != nil {
		return bodycodegen.Result{}, err
	}
	g, err := bodycodegen.NewTypedPathGenerator(modelPath)
	if err != nil {
		return bodycodegen.Result{}, err
	}
	return g.GenerateSourceAssembly(ctx, filename, source, activity)
}
