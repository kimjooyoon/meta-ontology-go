package symbolicinvocationusecase

import "encoding/json"

func EvaluateSymbolicReaderRequest(expectedSubjectSHA string, data []byte) ReaderObservationReport {
	var input ReaderRequestResultInput
	parseErr := json.Unmarshal(data, &input)
	evaluation := newReaderObservationEvaluation(input, expectedSubjectSHA, parseErr)
	indicators := evaluation.indicators()
	report := buildReaderObservationReport(input, expectedSubjectSHA, data, indicators)
	if report.Coordinates.Satisfied != SymbolicReaderObservationTotal {
		report.Decision = "FAIL_CLOSED"
		report.Resolution = "FAIL_CLOSED"
		report.Reason = evaluation.failureReason()
	}
	report.Digest = readerObservationReportDigest(report)
	return report
}
