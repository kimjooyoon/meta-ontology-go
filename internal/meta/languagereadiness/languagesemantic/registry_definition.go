package languagesemantic

const (
	expectedSources       = 30
	expectedLaws          = 3
	expectedRejections    = 2
	expectedSyntaxCases   = 33
	expectedSyntaxValid   = 30
	expectedSyntaxInvalid = 3
	expectedSyntaxFiles   = 36
	expectedSyntaxLines   = 449
)

func expectedSyntaxPackageUnits() []syntaxPackageUnit {
	return []syntaxPackageUnit{
		{
			ID:                   "billing-package",
			Path:                 "examples/billing-package",
			Members:              []string{"examples/billing-package/activity.gooo", "examples/billing-package/entities.gooo"},
			Entry:                "PayOrder",
			ReportSchema:         "gooo/language-package-execution-report/v1",
			MetaReducer:          "languagepackageexecution.Evaluate",
			SourceFilesIndicator: "PACKAGE_SOURCE_FILES",
			ExecutionIndicator:   "PACKAGE_EXECUTIONS",
		},
		{
			ID:                   "symbolic-invocation-schema",
			Path:                 "examples/symbolic-invocation-schema",
			Members:              []string{"examples/symbolic-invocation-schema/activity.gooo", "examples/symbolic-invocation-schema/entities.gooo", "examples/symbolic-invocation-schema/reader-request.gooo"},
			Entry:                "Checkout",
			ReportSchema:         "gooo/language-package-execution-report/v1",
			MetaReducer:          "languagepackageexecution.Evaluate",
			SourceFilesIndicator: "PACKAGE_SOURCE_FILES",
			ExecutionIndicator:   "PACKAGE_EXECUTIONS",
		},
	}
}
