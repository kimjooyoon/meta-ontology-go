package artifactemit

type symbolicReaderRequestChecks struct {
	SyntaxValid               bool
	SingleProjectionOperation bool
	SourceContractValid       bool
	SourceReadOnly            bool
	SourceIndicatorsUnique    bool
	AudienceKnown             bool
	ResolutionKnown           bool
	ReaderPresent             bool
	ReaderCountBound          bool
	ResolutionMatches         bool
	IndicatorsSelected        bool
	OutputSourceBound         bool
}

func symbolicReaderRequestCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func (checks symbolicReaderRequestChecks) values() []bool {
	return []bool{
		checks.SyntaxValid, checks.SingleProjectionOperation,
		checks.SourceContractValid, checks.SourceReadOnly,
		checks.SourceIndicatorsUnique, checks.AudienceKnown,
		checks.ResolutionKnown, checks.ReaderPresent,
		checks.ReaderCountBound, checks.ResolutionMatches,
		checks.IndicatorsSelected, checks.OutputSourceBound,
	}
}

func (checks symbolicReaderRequestChecks) coordinates() SymbolicReaderRequestCoordinates {
	satisfied := symbolicReaderRequestCount(checks.values()...)
	return SymbolicReaderRequestCoordinates{
		Satisfied: satisfied, Total: 12, BasisPoints: satisfied * 10000 / 12,
	}
}
