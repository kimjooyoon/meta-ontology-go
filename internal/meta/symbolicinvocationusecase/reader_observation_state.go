package symbolicinvocationusecase

import "strings"

type readerObservationEvaluation struct {
	parseErr          error
	schemaKnown       bool
	metricKnown       bool
	subjectBound      bool
	decisionPass      bool
	requestUser       bool
	audienceMatches   bool
	resolutionMatches bool
	selectionBound    bool
	readOnly          bool
	resultKnown       bool
}

func newReaderObservationEvaluation(input ReaderRequestResultInput, expectedSubjectSHA string, parseErr error) readerObservationEvaluation {
	valid := parseErr == nil
	requestUser := valid && input.Request.Audience == "USER"
	return readerObservationEvaluation{
		parseErr:          parseErr,
		schemaKnown:       valid && input.Schema == SymbolicReaderRequestResultSchema,
		metricKnown:       valid && input.MetricID == SymbolicReaderRequestResultMetric,
		subjectBound:      valid && expectedSubjectSHA != "" && input.SubjectSHA == expectedSubjectSHA,
		decisionPass:      valid && input.Decision == "PASS",
		requestUser:       requestUser,
		audienceMatches:   requestUser && input.View.Audience == input.Request.Audience,
		resolutionMatches: valid && input.Request.ExpectedResolution != "" && input.View.EffectiveResolution == input.Request.ExpectedResolution,
		selectionBound:    valid && readerObservationSelectionBound(input.View),
		readOnly:          valid && input.Effects.RepositoryWrites == 0 && !input.Effects.MutationAuthority && input.PromotionCreditBPS == 0,
		resultKnown:       valid && input.Resolution == "GOOO_REQUEST_BOUND_ONLY" && input.Reason == "CANONICAL_GOOO_READER_REQUEST_BOUND" && input.View.SourceResolution == "USER_VISIBLE" && readerObservationValidDigest(input.Digest),
	}
}

func readerObservationSelectionBound(view ReaderRequestViewInput) bool {
	if len(view.IndicatorIDs) == 0 || view.Coordinates.Satisfied != len(view.IndicatorIDs) || view.Coordinates.Total != len(view.IndicatorIDs) || view.Coordinates.BasisPoints != 10000 {
		return false
	}
	seen := make(map[string]struct{}, len(view.IndicatorIDs))
	for _, id := range view.IndicatorIDs {
		if strings.TrimSpace(id) == "" {
			return false
		}
		if _, exists := seen[id]; exists {
			return false
		}
		seen[id] = struct{}{}
	}
	return true
}
