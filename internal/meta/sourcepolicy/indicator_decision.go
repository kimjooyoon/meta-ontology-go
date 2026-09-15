package sourcepolicy

type IndicatorOutcome struct {
	Decision          IndicatorDecision `json:"decision"`
	EvaluationState   EvaluationState   `json:"evaluation_state"`
	FailureReason     FailureReason     `json:"failure_reason"`
	FailureCode       string            `json:"failure_code,omitempty"`
	EnforcementEffect EnforcementEffect `json:"enforcement_effect"`
}

func (indicator Indicator) Outcome() IndicatorOutcome {
	if indicator.Applicability == "NOT_APPLICABLE" {
		return IndicatorOutcome{
			Decision:          IndicatorDecisionNotApplicable,
			EvaluationState:   EvaluationStateEvaluated,
			FailureReason:     FailureReasonCatalogNotApplicable,
			FailureCode:       indicator.failureCode("catalog-not-applicable"),
			EnforcementEffect: EnforcementEffectNone,
		}
	}
	if indicator.Satisfied {
		return IndicatorOutcome{
			Decision:          IndicatorDecisionPass,
			EvaluationState:   EvaluationStateEvaluated,
			FailureReason:     FailureReasonNone,
			EnforcementEffect: indicator.enforcementEffect(true),
		}
	}
	return IndicatorOutcome{
		Decision:          IndicatorDecisionFailClosed,
		EvaluationState:   EvaluationStateEvaluated,
		FailureReason:     FailureReasonPredicateFalse,
		FailureCode:       indicator.failureCode("predicate-false"),
		EnforcementEffect: indicator.enforcementEffect(false),
	}
}

func (indicator Indicator) failureCode(reason string) string {
	return string(indicator.MetricID) + "#" + reason
}

func (indicator Indicator) enforcementEffect(satisfied bool) EnforcementEffect {
	if !indicator.Blocking {
		return EnforcementEffectNone
	}
	if satisfied {
		return EnforcementEffectAllow
	}
	return EnforcementEffectBlock
}

func (outcome IndicatorOutcome) Actionable() bool {
	return outcome.Decision == IndicatorDecisionFailClosed
}
