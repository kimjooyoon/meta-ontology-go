package sourcepolicy

type IndicatorDecision string

const (
	IndicatorDecisionPass          IndicatorDecision = "PASS"
	IndicatorDecisionFailClosed    IndicatorDecision = "FAIL_CLOSED"
	IndicatorDecisionNotApplicable IndicatorDecision = "NOT_APPLICABLE"
)

type EvaluationState string

const (
	EvaluationStateEvaluated EvaluationState = "EVALUATED"
)

type FailureReason string

const (
	FailureReasonNone                 FailureReason = "NONE"
	FailureReasonPredicateFalse       FailureReason = "PREDICATE_FALSE"
	FailureReasonCatalogNotApplicable FailureReason = "CATALOG_NOT_APPLICABLE"
)

type EnforcementEffect string

const (
	EnforcementEffectNone  EnforcementEffect = "NO_EFFECT"
	EnforcementEffectAllow EnforcementEffect = "ALLOW"
	EnforcementEffectBlock EnforcementEffect = "BLOCK"
)
