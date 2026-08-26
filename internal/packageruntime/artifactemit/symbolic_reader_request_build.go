package artifactemit

func buildSymbolicReaderRequestResult(
	request SymbolicReaderRequestDeclaration,
	projection SymbolicValueReaderProjection,
	reader SymbolicValueReaderProjectionView,
	checks symbolicReaderRequestChecks,
	subjectSHA string,
	projectionPayload []byte,
) (SymbolicReaderRequestResult, error) {
	decision, resolution, reason := checks.decision()
	view := SymbolicReaderRequestView{
		Audience: request.Audience, EffectiveResolution: "INVARIANT_ONLY",
		Coordinates: SymbolicValueContractCoordinates{},
	}
	if decision == "PASS" {
		view.SourceResolution = reader.SourceResolution
		view.EffectiveResolution = reader.EffectiveResolution
		view.IndicatorIDs = append([]string(nil), reader.IndicatorIDs...)
		view.Coordinates = reader.Coordinates
	}
	result := SymbolicReaderRequestResult{
		Schema: symbolicReaderRequestSchema, SubjectSHA: subjectSHA,
		MetricID: symbolicReaderRequestMetric, Decision: decision,
		Resolution: resolution, Reason: reason, Request: request,
		Source: SymbolicReaderRequestSource{
			Schema: projection.Schema, MetricID: projection.MetricID,
			Decision: projection.Decision, Resolution: projection.Resolution,
			Digest: projection.Digest, FileDigest: symbolicReaderBytesDigest(projectionPayload),
		},
		View: view, Coordinates: checks.coordinates(),
		Classes: symbolicReaderRequestClasses(checks),
		Indicators: symbolicReaderRequestIndicators(checks),
		Proofs: symbolicReaderRequestProofs(checks),
		Effects: SymbolicReaderRequestEffects{}, PromotionCreditBPS: 0,
		NotClaimed: []string{
			"reader comprehension", "arbitrary Gooo value-program execution",
			"external user adoption", "domain correctness", "production readiness",
			"access-control enforcement",
		},
	}
	var err error
	result.Digest, err = symbolicReaderRequestResultDigest(result)
	return result, err
}
