package languageconcept

func Catalog() []Concept {
	return []Concept{
		concept("metric-meta-program", "scalar metrics do not explain action", "every metric names its producer, consumer, and operation", "bind-indicator-meta-program", "OPERATING",
			[]string{"internal/meta/metabinding", "internal/meta/metricprogram"}, []string{"gooo.metric.meta.unbound-indicators.v1"},
			UseCase{"metric-without-operation", "an indicator lacks a meta operation", "FAIL_CLOSED"}),
		concept("executable-actionability", "advice can be impossible to execute", "blocking indicators resolve to canonical executors", "resolve-indicator-executor", "OPERATING",
			[]string{"internal/meta/actionability", "bootstrap/meta-binding-witness"}, []string{"gooo.metric.meta.actionable-blocking.coverage-bps.v1"},
			UseCase{"missing-executor", "a blocking operation has no executor", "FAIL_CLOSED"}),
		concept("effect-bounded-observation", "analysis tools can mutate the subject", "observers carry a zero-write guardrail", "preserve-read-only-semantic-state", "CONFORMED",
			[]string{"internal/meta/artifactfeedback", "internal/meta/feedbackpredecessor", "internal/meta/feedbackstate"}, []string{"gooo.metric.meta.predecessor-observer-writes.guardrail.v1"},
			UseCase{"observer-write", "a receipt reports a repository write", "FAIL_CLOSED"}),
		concept("monotone-semantic-resolution", "unknown meaning can masquerade as success", "uncertainty descends from exact operation to invariant only", "lower-semantic-resolution", "CONFORMED",
			[]string{"internal/meta/semanticresolution", "internal/meta/feedbackstate"}, []string{"gooo.metric.meta.semantic-resolution-descents.guardrail.v1"},
			UseCase{"unknown-top-decision", "an exact decision is unknown", "LOWER_RESOLUTION"}),
		concept("causal-feedback-chain", "a green artifact is not necessarily the cause", "the unique exact predecessor payload becomes executable semantic state", "select-predecessor-semantic-state", "OPERATING",
			[]string{"internal/meta/feedbackpredecessor", "internal/meta/feedbackstate", "cmd/feedback-semantic-state-witness"}, []string{"gooo.metric.meta.predecessor-feedback-readiness.coverage-bps.v1", "gooo.metric.meta.semantic-snapshot.readiness-bps.v1"},
			UseCase{"ambiguous-rerun", "two canonical predecessor artifacts exist", "FAIL_CLOSED"}),
		concept("ci-selected-refactoring", "manual cleanup can outrun its evidence", "CI metrics select bounded AST rewrites before acceptance", "compact-obvious-lines", "OPERATING",
			[]string{"bootstrap/logical-split-planner", "bootstrap/line-density-rewriter", "bootstrap/function-extractor"}, []string{"source.line-cap-debt"},
			UseCase{"line-cap-debt", "a Go source has 77 lines under a 75 line cap", "REWRITE_TO_71_AND_REPLAY"}),
		concept("concept-governed-refactoring", "a metric-selected operation can lack semantic authorization", "every candidate operation is digest-bound to a language concept", "bind-refactoring-concept", "OPERATING",
			[]string{"internal/meta/metricstrategy"}, []string{"gooo.metric.meta.concept-operation-binding-bps.v1"},
			UseCase{"unregistered-strategy-operation", "a strategy names an operation without a concept binding", "LOWER_RESOLUTION"}),
		concept("quantified-improvement", "qualitative progress can conceal an unchanged denominator", "exact predecessor and current receipts prove integer gains without inference", "compare-readiness-receipts", "OPERATING",
			[]string{"internal/meta/languagereadiness/improvement", "internal/meta/languagereadiness/artifact", "cmd/language-readiness-witness/transition"},
			[]string{"completed-obligations", "readiness-basis-points", "newly-satisfied", "regressions", "unresolved-evidence"},
			UseCase{"one-obligation-gain", "comparable 7/24 and 8/24 receipts have one gain and zero regressions", "IMPROVED_PLUS_1_PLUS_417_BPS"}),
		concept("verified-transformation-transaction", "a passing proposal can lack a merged predecessor witness", "the merged predecessor receipt is selected and replayed before readiness promotion", "verify-transformation-transaction", "OPERATING",
			[]string{"internal/meta/languagereadiness/artifact/predecessorselection", "internal/meta/languagereadiness/artifact/predecessorbinding", "cmd/language-readiness-witness/predecessor-selection"},
			[]string{"gooo.metric.language.predecessor-dynamic-binding-bps.v1", "gooo.metric.language.predecessor-dynamic-coordinates.v1", "gooo.metric.language.predecessor-static-coordinates.guardrail.v1", "gooo.metric.language.predecessor-unknown-coordinates.guardrail.v1", "gooo.metric.language.predecessor-observer-writes.guardrail.v1"},
			UseCase{"merged-dynamic-predecessor", "the merged predecessor changes eight static coordinates into eight dynamic inputs with zero unknowns and writes", "IMPROVED_STATIC_8_TO_0_DYNAMIC_0_TO_8_BPS_0_TO_10000"}),
	}
}

func concept(id, problem, effect, operation, stage string, code, metrics []string, useCase UseCase) Concept {
	return Concept{ID: id, Problem: problem, PositiveEffect: effect, MetaOperation: operation,
		Rarity: "UNCOMMON_COMBINATION", Stage: stage, CodeBindings: code,
		MetricBindings: metrics, UseCases: []UseCase{useCase}}
}
