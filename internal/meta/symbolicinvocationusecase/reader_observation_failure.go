package symbolicinvocationusecase

func (evaluation readerObservationEvaluation) failureReason() string {
	switch {
	case evaluation.parseErr != nil:
		return "READER_REQUEST_INPUT_INVALID"
	case !evaluation.schemaKnown:
		return "READER_REQUEST_SCHEMA_UNKNOWN"
	case !evaluation.metricKnown:
		return "READER_REQUEST_METRIC_UNKNOWN"
	case !evaluation.subjectBound:
		return "READER_REQUEST_SUBJECT_MISMATCH"
	case !evaluation.decisionPass:
		return "READER_REQUEST_DECISION_NOT_EXPLICIT_PASS"
	case !evaluation.requestUser:
		return "READER_REQUEST_AUDIENCE_UNKNOWN"
	case !evaluation.audienceMatches:
		return "READER_REQUEST_AUDIENCE_MISMATCH"
	case !evaluation.resolutionMatches:
		return "READER_REQUEST_RESOLUTION_MISMATCH"
	case !evaluation.selectionBound:
		return "READER_REQUEST_SELECTION_COUNT_MISMATCH"
	case !evaluation.readOnly:
		return "READER_REQUEST_UNSAFE_EFFECTS"
	case !evaluation.resultKnown:
		return "READER_REQUEST_RESULT_SEMANTICS_UNKNOWN"
	default:
		return "READER_REQUEST_INDICATOR_UNKNOWN"
	}
}
