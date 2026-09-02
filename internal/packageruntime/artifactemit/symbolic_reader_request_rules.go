package artifactemit

func symbolicReaderRequestExpectedSourceResolution(audience string) string {
	switch audience {
	case "USER":
		return "USER_VISIBLE"
	case "TOOL_AUTHOR":
		return "TOOL_CONTRACT"
	case "GOVERNOR":
		return "FULL_RECEIPT"
	default:
		return ""
	}
}

func symbolicReaderRequestCountBound(reader SymbolicValueReaderProjectionView) bool {
	return reader.Coordinates.Satisfied == reader.Coordinates.Total &&
		reader.Coordinates.Total == len(reader.IndicatorIDs) &&
		reader.Coordinates.BasisPoints == 10000
}

func symbolicReaderRequestResolutionMatches(
	request SymbolicReaderRequestDeclaration,
	reader SymbolicValueReaderProjectionView,
) bool {
	return request.ExpectedResolution == reader.EffectiveResolution &&
		reader.SourceResolution == symbolicReaderRequestExpectedSourceResolution(request.Audience)
}
