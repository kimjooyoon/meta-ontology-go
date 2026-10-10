package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func runSourceBodyContext(ctx context.Context, source, raw []byte, assembly *assemblyspec.Spec,
	options bodyContextArgs, stdout, stderr io.Writer) int {
	if bodycodegen.IsRecordAssembly(assembly) {
		return runRecordBodyContext(ctx, source, raw, options, stdout)
	}
	if options.valueFlow {
		return bodyContextFailure(stdout, fmt.Errorf("value flow requires record source assembly"))
	}
	if bodycodegen.IsSourceIRSearch(assembly) {
		return runSearchBodyContext(ctx, source, options, stdout, stderr)
	}
	return runTypedBodyContext(ctx, source, raw, options, stdout, stderr)
}

func runRecordBodyContext(ctx context.Context, source, raw []byte, options bodyContextArgs, stdout io.Writer) int {
	if len(raw) != 0 {
		return bodyContextFailure(stdout, fmt.Errorf("record source assembly owns its plan"))
	}
	result, err := exportRecordBodyContext(ctx, source, options)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	if err = json.NewEncoder(stdout).Encode(result); err != nil {
		return bodyContextFailure(stdout, err)
	}
	return exitOK
}

func runSearchBodyContext(ctx context.Context, source []byte, options bodyContextArgs, stdout, stderr io.Writer) int {
	if options.plan != "" || options.featureExplicit || options.model != "" {
		return bodyContextFailure(stdout, fmt.Errorf("source search owns its plan and has no model feature encoding"))
	}
	result, err := bodycodegen.ExportSourceIRSearchContext(ctx, options.filename, source, options.activity, options.includePlan)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	return writeBodyContextOutput(stdout, stderr, result)
}

func runTypedBodyContext(ctx context.Context, source, raw []byte, options bodyContextArgs, stdout, stderr io.Writer) int {
	document, err := bodycodegen.DecodeSourcePathDocument(ctx, options.filename, source, options.activity, raw)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	result, err := exportTypedBodyContext(ctx, source, document, options)
	if err != nil {
		return bodyContextFailure(stdout, err)
	}
	output := bodyContextOutput{TypedPathContextExport: result}
	if options.includePlan {
		output.ExpandedPlan = &document.Plan
	}
	return writeBodyContextOutput(stdout, stderr, output)
}

func writeBodyContextOutput(stdout, stderr io.Writer, output any) int {
	if err := json.NewEncoder(stdout).Encode(output); err != nil {
		fmt.Fprintln(stderr, "context export output failed")
		return exitFailure
	}
	return exitOK
}
