package valueexecution

import "io/fs"

func Evaluate(filesystem fs.FS, sourcePath, activity, headSHA string) Report {
	report := newReport(sourcePath, headSHA)
	source, err := fs.ReadFile(filesystem, sourcePath)
	if err != nil {
		return finalize(report)
	}
	report.SourceBytes = len(source)
	report.SourceLines = countLines(source)
	report.SourceDigest = digestBytes(source)
	program, err := Compile(sourcePath, source, activity)
	if err != nil {
		report.Reason = Reason(err)
		return finalize(report)
	}
	return evaluateProgram(report, source, program)
}

func newReport(sourcePath, headSHA string) Report {
	return Report{
		Schema: ReportSchema, Decision: DecisionFailClosed, Reason: ReasonSourceReadFailed,
		Resolution: ResolutionSyntaxOnly, HeadSHA: headSHA, SourcePath: sourcePath,
		NonClaims: []string{
			"general expression language", "arbitrary value types", "core semantic IR value-program preservation",
			"runtime memory or performance bounds", "repository mutation, promotion, or automatic adoption",
		},
	}
}
