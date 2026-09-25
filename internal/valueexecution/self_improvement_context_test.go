package valueexecution

import "testing"

func selfImprovementContext(source, semantic string) ImprovementContextObservation {
	return ObserveImprovementEvaluationContext(EnvironmentObservation{
		Inputs: EnvironmentInputs{
			SourceDigest:        digestValue(source),
			SemanticFingerprint: digestValue(semantic),
			ToolchainDigest:     digestValue("toolchain"),
			ContractDigest:      digestValue("contract"),
			ModelDigest:         digestValue("model"),
			SkillDigest:         digestValue("skill"),
			GatewayPolicyDigest: digestValue("gateway"),
		},
	})
}

func TestObserveSelfImprovementWithContextAllowsDifferentCandidates(t *testing.T) {
	before := selfImprovementContext("source-v1", "semantic-v1")
	after := selfImprovementContext("source-v2", "semantic-v2")
	observation := ObserveSelfImprovementWithContext(
		selfImprovementOrigin("candidate-before"), selfImprovementOrigin("candidate-after"),
		before, after,
		selfImprovementMetric(10, true, "metric-before"),
		selfImprovementMetric(7, true, "metric-after"),
	)
	if observation.Status != SelfImprovementStatusImproved || observation.Reason != "METRIC_IMPROVED" {
		t.Fatalf("observation=%#v, want improvement under fixed context", observation)
	}
	if observation.EnvironmentTransition != EnvironmentTransitionUnchanged || !validDigest(observation.EnvironmentDigest) {
		t.Fatalf("observation=%#v, want unchanged digest-backed context", observation)
	}
}

func TestObserveSelfImprovementWithContextRejectsChangedToolchain(t *testing.T) {
	before := selfImprovementContext("source-v1", "semantic-v1")
	after := before
	after.Inputs.ToolchainDigest = digestValue("toolchain-v2")
	after.Digest = improvementContextDigest(after)
	observation := ObserveSelfImprovementWithContext(
		selfImprovementOrigin("candidate-before"), selfImprovementOrigin("candidate-after"),
		before, after,
		selfImprovementMetric(10, true, "metric-before"),
		selfImprovementMetric(7, true, "metric-after"),
	)
	if observation.Status != SelfImprovementStatusUnknown || observation.Reason != "EVALUATION_CONTEXT_CHANGED" {
		t.Fatalf("observation=%#v, want UNKNOWN for changed toolchain", observation)
	}
}

func TestObserveImprovementEvaluationContextPreservesMissingBoundary(t *testing.T) {
	observation := ObserveImprovementEvaluationContext(EnvironmentObservation{
		Inputs: EnvironmentInputs{ToolchainDigest: digestValue("toolchain")},
	})
	if observation.Status != ImprovementContextStatusUnknown || len(observation.Missing) != 4 || observation.Digest != "" {
		t.Fatalf("observation=%#v, want incomplete context", observation)
	}
	if got := CompareImprovementContexts(observation, observation); got != EnvironmentTransitionUnknown {
		t.Fatalf("incomplete context transition=%s, want %s", got, EnvironmentTransitionUnknown)
	}
}
