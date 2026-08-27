package valuecatalog

import (
	"io/fs"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func Evaluate(filesystem fs.FS, sourcePath, headSHA string) Report {
	report := newReport(sourcePath, headSHA)
	source, err := fs.ReadFile(filesystem, sourcePath)
	if err != nil {
		return finalize(report)
	}
	report.SourceDigest = digestBytes(source)
	report.SourceLines = countLines(source)
	report.ProcessCoordinate = ProcessCoordinate{Stage: "LOWER", Step: "observe-operation-catalog", Reason: ReasonObservationFailed}
	observed, err := observe(sourcePath, source)
	if err != nil {
		report.Reason = ReasonObservationFailed
		report.Diagnostic = err.Error()
		if failure, ok := valueexecution.FailureOf(err); ok {
			report.ProcessCoordinate = ProcessCoordinate{Stage: failure.Stage, Step: failure.Step, Reason: failure.Code}
		}
		return finalize(report)
	}
	report.Resolution = ResolutionCoreValue
	report.BeforeSourceDigest = digestBytes(observed.baselineSource)
	report.ActivitiesObserved = observed.actualCore.activities
	report.CoreIRFingerprint = observed.actualCore.fingerprint
	report.BaselineCoreProgram = observed.actualCore.programs[BaselineActivity]
	report.ExtensionCoreProgram = observed.actualCore.programs[ExtensionActivity]
	report.OperationSpecs = valueexecution.CanonicalOperationSpecs()
	report.Baseline, report.Extension = observed.baseline, observed.extension
	report.Improvement = Improvement{
		ID: "source-only-catalog-extension", Before: coordinate(0, 1),
		After:          coordinate(boolInt(observed.extensionPresent), 1),
		BeforeEvidence: observed.beforeReason, AfterEvidence: report.SourceDigest,
	}
	report.Summary = Summary{
		BaselineCasesPassed: report.Baseline.Passed, BaselineCasesTotal: len(catalogInputs),
		ExtensionCasesPassed: report.Extension.Passed, ExtensionCasesTotal: len(catalogInputs),
		UnknownCounterexamplePassed: observed.unknownReason == valueexecution.ReasonProgramUnknown,
		RepositoryWrites:            0, CoreFingerprintSensitive: coordinate(boolInt(observed.coreFingerprintSensitive), 1),
	}
	if observed.extensionPresent {
		report.Decision, report.Reason = DecisionExtensionProven, ReasonExtensionExact
	} else {
		report.Decision, report.Reason = DecisionBaselineObserved, ReasonBaselineObserved
	}
	report.ProcessCoordinate = ProcessCoordinate{Stage: "REDUCE", Step: "close-os9", Reason: report.Reason}
	report = closeOperationSpec(report)
	report.Indicators = buildIndicators(report)
	report.Views = buildViews(report.Indicators)
	report.Proofs = buildProofs(report)
	return finalize(report)
}

func newReport(path, head string) Report {
	return Report{
		Schema: ReportSchema, Decision: DecisionFailClosed, Reason: ReasonSourceReadFailed,
		Resolution: ResolutionSyntaxOnly, SourcePath: path, HeadSHA: head,
		ProcessCoordinate: ProcessCoordinate{Stage: "READ", Step: "read-catalog-source", Reason: ReasonSourceReadFailed},
		NonClaims: []string{
			"source-only catalog extension achieved in the baseline", "arbitrary primitive operations",
			"general expressions or multi-file linking", "runtime memory or performance bounds",
			"repository mutation, promotion, or automatic adoption",
		},
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
