package artifactemit

func (checks symbolicReaderRequestChecks) decision() (string, string, string) {
	if reason := checks.failureReason(); reason != "" {
		return "FAIL_CLOSED", "INVARIANT_ONLY", reason
	}
	return "PASS", "GOOO_REQUEST_BOUND_ONLY", "CANONICAL_GOOO_READER_REQUEST_BOUND"
}

func (checks symbolicReaderRequestChecks) failureReason() string {
	ordered := []struct {
		passed bool
		reason string
	}{
		{checks.SyntaxValid, "GOOO_READER_REQUEST_SYNTAX_INVALID"},
		{checks.SingleProjectionOperation, "GOOO_READER_REQUEST_OPERATION_INVALID"},
		{checks.AudienceKnown, "GOOO_READER_REQUEST_AUDIENCE_UNKNOWN"},
		{checks.ResolutionKnown, "GOOO_READER_REQUEST_RESOLUTION_UNKNOWN"},
		{checks.SourceContractValid, "GOOO_READER_REQUEST_SOURCE_INVALID"},
		{checks.SourceReadOnly, "GOOO_READER_REQUEST_SOURCE_NOT_READ_ONLY"},
		{checks.ReaderPresent, "GOOO_READER_REQUEST_READER_MISSING"},
		{checks.ReaderCountBound, "GOOO_READER_REQUEST_COUNT_MISMATCH"},
		{checks.ResolutionMatches, "GOOO_READER_REQUEST_RESOLUTION_MISMATCH"},
		{checks.SourceIndicatorsUnique, "GOOO_READER_REQUEST_INDICATORS_NOT_UNIQUE"},
		{checks.IndicatorsSelected, "GOOO_READER_REQUEST_INDICATORS_UNAVAILABLE"},
		{checks.OutputSourceBound, "GOOO_READER_REQUEST_SOURCE_BINDING_INVALID"},
	}
	for _, check := range ordered {
		if !check.passed {
			return check.reason
		}
	}
	return ""
}
