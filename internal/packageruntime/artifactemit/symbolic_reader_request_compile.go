package artifactemit

func CompileSymbolicReaderRequest(
	requestPayload, projectionPayload []byte,
	subjectSHA string,
) (SymbolicReaderRequestResult, error) {
	request, syntaxValid, operationValid := parseSymbolicReaderRequest(requestPayload)
	projection, decoded := decodeSymbolicReaderRequestSource(projectionPayload)
	reader, readerPresent := symbolicReaderRequestReader(projection.Readers, request.Audience)
	checks := symbolicReaderRequestChecks{
		SyntaxValid:               syntaxValid,
		SingleProjectionOperation: operationValid,
		SourceContractValid:       decoded && symbolicReaderRequestSourceValid(projection, subjectSHA),
		SourceReadOnly: decoded && projection.Effects.RepositoryWrites == 0 &&
			!projection.Effects.MutationAuthority && projection.PromotionCreditBPS == 0,
		AudienceKnown:  symbolicReaderRequestAudienceKnown(request.Audience),
		ResolutionKnown: symbolicReaderRequestResolutionKnown(request.ExpectedResolution),
		ReaderPresent:   readerPresent,
	}
	checks.ReaderCountBound = readerPresent && symbolicReaderRequestCountBound(reader)
	checks.ResolutionMatches = readerPresent && symbolicReaderRequestResolutionMatches(request, reader)
	checks.SourceIndicatorsUnique = readerPresent && symbolicReaderRequestIDsUnique(reader.IndicatorIDs)
	checks.IndicatorsSelected = checks.ReaderCountBound && len(reader.IndicatorIDs) > 0
	checks.OutputSourceBound = checks.SourceContractValid &&
		symbolicReaderValidDigest(request.SourceDigest) &&
		symbolicReaderValidDigest(symbolicReaderBytesDigest(projectionPayload))
	return buildSymbolicReaderRequestResult(
		request, projection, reader, checks, subjectSHA, projectionPayload,
	)
}
