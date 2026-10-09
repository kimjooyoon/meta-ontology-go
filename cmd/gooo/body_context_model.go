package main

import (
	"context"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

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
