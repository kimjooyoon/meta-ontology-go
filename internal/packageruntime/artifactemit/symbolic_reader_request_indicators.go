package artifactemit

func symbolicReaderRequestIndicator(
	id, class, proofChoice, operation string,
	satisfied bool,
) SymbolicReaderRequestIndicator {
	observed := 0
	if satisfied {
		observed = 1
	}
	return SymbolicReaderRequestIndicator{
		ID: id, Class: class, ProofChoice: proofChoice, MetaOperation: operation,
		Observed: observed, Expected: 1, Satisfied: satisfied,
	}
}

func symbolicReaderRequestIndicators(checks symbolicReaderRequestChecks) []SymbolicReaderRequestIndicator {
	indicator := symbolicReaderRequestIndicator
	return []SymbolicReaderRequestIndicator{
		indicator("request.syntax-valid", "GUARDRAIL", "FOUNDATION", "parse-request.syntax", checks.SyntaxValid),
		indicator("request.single-operation", "GUARDRAIL", "FOUNDATION", "select-request.operation", checks.SingleProjectionOperation),
		indicator("source.contract-valid", "GUARDRAIL", "FOUNDATION", "verify-source.contract", checks.SourceContractValid),
		indicator("source.read-only", "GUARDRAIL", "REGRESSION", "guard-source.read-only", checks.SourceReadOnly),
		indicator("source.indicators-unique", "GUARDRAIL", "REGRESSION", "guard-source.unique-indicators", checks.SourceIndicatorsUnique),
		indicator("request.audience-known", "DRIVER", "FOUNDATION", "bind-request.audience", checks.AudienceKnown),
		indicator("request.resolution-known", "DRIVER", "FOUNDATION", "bind-request.resolution", checks.ResolutionKnown),
		indicator("source.reader-present", "DRIVER", "COHERENCE", "select-source.reader", checks.ReaderPresent),
		indicator("source.reader-count-bound", "DRIVER", "COHERENCE", "bind-source.reader-count", checks.ReaderCountBound),
		indicator("request.resolution-matches", "OUTCOME", "COHERENCE", "cohere-request.resolution", checks.ResolutionMatches),
		indicator("view.indicators-selected", "OUTCOME", "COHERENCE", "project-view.indicators", checks.IndicatorsSelected),
		indicator("output.source-bound", "OUTCOME", "REGRESSION", "bind-output.source", checks.OutputSourceBound),
	}
}
