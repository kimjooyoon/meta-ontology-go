package symbolicinvocationusecase

func (evaluation readerObservationEvaluation) indicators() []ReaderObservationIndicator {
	return []ReaderObservationIndicator{
		readerObservationIndicator("artifact.schema-known", "GUARDRAIL", "FOUNDATION", "recognize-artifact.schema", evaluation.schemaKnown),
		readerObservationIndicator("artifact.metric-known", "GUARDRAIL", "FOUNDATION", "recognize-artifact.metric", evaluation.metricKnown),
		readerObservationIndicator("artifact.subject-bound", "GUARDRAIL", "REGRESSION", "bind-artifact.subject", evaluation.subjectBound),
		readerObservationIndicator("artifact.decision-explicit-pass", "GUARDRAIL", "REGRESSION", "guard-artifact.explicit-pass", evaluation.decisionPass),
		readerObservationIndicator("request.user-audience", "DRIVER", "FOUNDATION", "select-request.user-audience", evaluation.requestUser),
		readerObservationIndicator("view.audience-matches", "DRIVER", "COHERENCE", "cohere-view.audience", evaluation.audienceMatches),
		readerObservationIndicator("view.resolution-matches", "DRIVER", "COHERENCE", "cohere-view.resolution", evaluation.resolutionMatches),
		readerObservationIndicator("view.selection-count-bound", "OUTCOME", "COHERENCE", "bind-view.selection-count", evaluation.selectionBound),
		readerObservationIndicator("effects.read-only", "OUTCOME", "REGRESSION", "guard-observation.read-only", evaluation.readOnly),
		readerObservationIndicator("observation.compiler-result-known", "OUTCOME", "FOUNDATION", "observe-compiler.reader-result", evaluation.resultKnown),
	}
}

func readerObservationIndicator(id, class, proofChoice, metaOperation string, satisfied bool) ReaderObservationIndicator {
	observed := 0
	if satisfied {
		observed = 1
	}
	return ReaderObservationIndicator{
		ID:            id,
		Class:         class,
		ProofChoice:   proofChoice,
		MetaOperation: metaOperation,
		Observed:      observed,
		Expected:      1,
		Satisfied:     satisfied,
	}
}
