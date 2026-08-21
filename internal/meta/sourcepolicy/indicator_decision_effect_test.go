package sourcepolicy

import "testing"

func TestIndicatorOutcomeClosesEnforcementEffect(t *testing.T) {
	cases := []struct {
		name      string
		indicator Indicator
		decision  IndicatorDecision
		effect    EnforcementEffect
	}{
		{
			name: "blocking pass",
			indicator: Indicator{
				MetricID: MetricFileLines, Blocking: true, Satisfied: true,
			},
			decision: IndicatorDecisionPass, effect: EnforcementEffectAllow,
		},
		{
			name: "blocking failure",
			indicator: Indicator{
				MetricID: MetricFileLines, Blocking: true, Satisfied: false,
			},
			decision: IndicatorDecisionFailClosed, effect: EnforcementEffectBlock,
		},
		{
			name: "observational failure",
			indicator: Indicator{
				MetricID: MetricFileLines, Blocking: false, Satisfied: false,
			},
			decision: IndicatorDecisionFailClosed, effect: EnforcementEffectNone,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			outcome := test.indicator.Outcome()
			if outcome.Decision != test.decision || outcome.EnforcementEffect != test.effect {
				t.Fatalf("outcome = %+v", outcome)
			}
		})
	}
}
