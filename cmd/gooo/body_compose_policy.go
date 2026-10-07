package main

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func readRecordAssemblyPolicy(flags map[string]string) (*bodycodegen.RecordAssemblyPolicy, error) {
	if flags["--assembly-policy"] == "" {
		return nil, nil
	}
	source, err := readBodyExecutionFile(flags["--assembly-policy"], 128<<10)
	if err != nil {
		return nil, err
	}
	return &bodycodegen.RecordAssemblyPolicy{Source: string(source), Activity: flags["--policy-activity"]}, nil
}

func resumeBodyComposition(ctx context.Context, flags map[string]string, source []byte,
	suite bodyexecution.CompositionCases, policy bodycodegen.RecordAssemblyPolicy) (bodyexecution.Composition, error) {
	raw, err := readBodyExecutionFile(flags["--resume-composition"], 32<<20)
	if err != nil {
		return bodyexecution.Composition{}, err
	}
	prior, err := bodyexecution.DecodeComposition(raw)
	if err != nil {
		return bodyexecution.Composition{}, err
	}
	return bodyexecution.ResumeComposition(ctx, flags["--source"], source, prior, suite, policy)
}

func buildOrReadBodyComposition(ctx context.Context, flags map[string]string, source []byte,
	suite bodyexecution.CompositionCases) (bodyexecution.Composition, error) {
	if flags["--composition"] != "" {
		raw, err := readBodyExecutionFile(flags["--composition"], 32<<20)
		if err != nil {
			return bodyexecution.Composition{}, err
		}
		return bodyexecution.DecodeComposition(raw)
	}
	options := bodyexecution.CompositionOptions{ModelPath: flags["--model"], FillModelPath: flags["--fill-model"]}
	var err error
	options.RecordPolicy, err = readRecordAssemblyPolicy(flags)
	if err != nil {
		return bodyexecution.Composition{}, err
	}
	if flags["--resume-composition"] != "" {
		return resumeBodyComposition(ctx, flags, source, suite, *options.RecordPolicy)
	}
	return bodyexecution.GenerateCompositionWithOptions(ctx, flags["--source"], source, suite, options)
}
