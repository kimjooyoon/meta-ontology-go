package languagesyntax

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagepackageexecution"

func expectedSyntaxPackageUnits() []PackageDefinition {
	packageUnit := PackageDefinition{ID: "billing-package", Path: "examples/billing-package", Members: []string{"examples/billing-package/activity.gooo", "examples/billing-package/entities.gooo"}, Entry: "PayOrder", ReportSchema: languagepackageexecution.ReportSchema, MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES", ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	symbolicUnit := PackageDefinition{ID: "symbolic-invocation-schema", Path: "examples/symbolic-invocation-schema", Members: []string{"examples/symbolic-invocation-schema/activity.gooo", "examples/symbolic-invocation-schema/entities.gooo", "examples/symbolic-invocation-schema/reader-request.gooo"}, Entry: "Checkout", ReportSchema: languagepackageexecution.ReportSchema, MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES", ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	selfImprovementObservationUnit := PackageDefinition{ID: "self-improvement-observation", Path: "examples/self-improvement-observation", Members: []string{"examples/self-improvement-observation/observation.gooo"}, Entry: "DeclareOperationIntent", ReportSchema: languagepackageexecution.ReportSchema, MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES", ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	partialReuseUnit := PackageDefinition{ID: "self-improvement-partial-reuse", Path: "examples/self-improvement-partial-reuse", Members: []string{"examples/self-improvement-partial-reuse/main.gooo"}, Entry: "CreateReceipt", ReportSchema: languagepackageexecution.ReportSchema, MetaReducer: "languagepackageexecution.Evaluate", SourceFilesIndicator: "PACKAGE_SOURCE_FILES", ExecutionIndicator: "PACKAGE_EXECUTIONS"}
	return []PackageDefinition{packageUnit, symbolicUnit, selfImprovementObservationUnit, partialReuseUnit}
}
