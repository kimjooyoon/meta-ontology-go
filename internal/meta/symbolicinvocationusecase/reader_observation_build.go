package symbolicinvocationusecase

func buildReaderObservationReport(input ReaderRequestResultInput, expectedSubjectSHA string, data []byte, indicators []ReaderObservationIndicator) ReaderObservationReport {
	return ReaderObservationReport{
		Schema:     SymbolicReaderObservationSchema,
		SubjectSHA: expectedSubjectSHA,
		MetricID:   SymbolicReaderObservationMetric,
		Decision:   "PASS",
		Resolution: SymbolicReaderObservationResolution,
		Reason:     "CANONICAL_READER_REQUEST_OBSERVED",
		Source: ReaderObservationSource{
			Schema:         input.Schema,
			MetricID:       input.MetricID,
			SubjectSHA:     input.SubjectSHA,
			Decision:       input.Decision,
			Resolution:     input.Resolution,
			ArtifactDigest: input.Digest,
			FileDigest:     readerObservationBytesDigest(data),
		},
		View: ReaderObservationView{
			Audience:             input.View.Audience,
			EffectiveResolution:  input.View.EffectiveResolution,
			SelectedIndicatorIDs: append([]string(nil), input.View.IndicatorIDs...),
		},
		Coordinates:        readerObservationCoordinates(indicators),
		Classes:            readerObservationClasses(indicators),
		Indicators:         indicators,
		Proofs:             readerObservationProofs(indicators),
		Effects:            ReaderObservationEffects{},
		PromotionCreditBPS: 0,
		NotClaimed: []string{
			"human comprehension",
			"authorization enforcement",
			"arbitrary Gooo value-program execution",
			"exact-head cross-job artifact transport",
			"performance determinism",
			"production readiness",
		},
	}
}
