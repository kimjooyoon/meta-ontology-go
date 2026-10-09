package main

import (
	"context"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func exportTypedBodyContext(ctx context.Context, source []byte, document pathplan.Document,
	options bodyContextArgs) (bodycodegen.TypedPathContextExport, error) {
	if options.model == "" {
		return bodycodegen.ExportTypedPathContextWithFeature(ctx, options.filename, source,
			options.activity, document, options.featureVersion)
	}
	feature := ""
	if options.featureExplicit {
		feature = options.featureVersion
	}
	return bodycodegen.ExportTypedPathModelContext(ctx, options.filename, source, options.activity,
		document, options.model, feature)
}

func exportRecordBodyContext(ctx context.Context, source []byte, options bodyContextArgs) (bodycodegen.RecordAssemblyContextExport, error) {
	if options.model != "" {
		modelOptions := bodycodegen.RecordModelContextOptions{IncludePlan: options.includePlan, ValueFlow: options.valueFlow}
		if options.featureExplicit {
			modelOptions.FeatureVersion = options.featureVersion
		}
		return bodycodegen.ExportRecordAssemblyModelContext(ctx, options.filename, source, options.activity, options.model, modelOptions)
	}
	export := bodycodegen.ExportRecordAssemblyContextWithFeature
	if options.valueFlow {
		export = bodycodegen.ExportRecordAssemblyContextWithFlow
	}
	return export(ctx, options.filename, source, options.activity, options.includePlan, options.featureVersion)
}
