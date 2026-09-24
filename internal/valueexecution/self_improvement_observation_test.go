package valueexecution

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func selfImprovementOrigin(evidence string) provenance.OriginChainObservation {
	return provenance.ObserveOriginChain(provenance.OriginChain{
		DeclarationURI:        "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol:     "activity CommitCandidate",
		IRNode:                "activity:CommitCandidate",
		GeneratedURI:          "cmd/gooo/run_source_typed_chain_part01_test.go",
		GeneratedSymbol:       "TestTypedChain",
		ReverseObservationURI: "internal/valueexecution/replay.go",
		ReverseObservation:    "execution.replay.completed",
		MetricName:            "gooo.self-improvement.metric.v1",
		MetricValue:           "candidate-observed",
		EvidenceDigest:        digestValue(evidence),
	})
}

func selfImprovementEnvironment(source string) EnvironmentObservation {
	return ObserveEnvironment(EnvironmentInputs{
		SourceDigest:        digestValue(source),
		SemanticFingerprint: digestValue("semantic"),
		ToolchainDigest:     digestValue("toolchain"),
		ContractDigest:      digestValue("contract"),
		ModelDigest:         digestValue("model"),
		SkillDigest:         digestValue("skill"),
		GatewayPolicyDigest: digestValue("gateway"),
	})
}

func selfImprovementMetric(value int64, lowerIsBetter bool, evidence string) ImprovementMetric {
	return ImprovementMetric{
		Name:           "runtime.steps",
		Value:          value,
		LowerIsBetter:  lowerIsBetter,
		EvidenceDigest: digestValue(evidence),
	}
}

func TestObserveSelfImprovementReportsImprovementWithStableEnvironment(t *testing.T) {
	environment := selfImprovementEnvironment("source-v1")
	observation := ObserveSelfImprovement(
		selfImprovementOrigin("before"), selfImprovementOrigin("after"),
		environment, environment,
		selfImprovementMetric(10, true, "metric-before"),
		selfImprovementMetric(7, true, "metric-after"),
	)
	if observation.Status != SelfImprovementStatusImproved {
		t.Fatalf("status=%s, want %s: %#v", observation.Status, SelfImprovementStatusImproved, observation)
	}
	if observation.OriginTransition != provenance.OriginChainTransitionChanged {
		t.Fatalf("origin transition=%s, want %s", observation.OriginTransition, provenance.OriginChainTransitionChanged)
	}
	if observation.EnvironmentTransition != EnvironmentTransitionUnchanged {
		t.Fatalf("environment transition=%s, want %s", observation.EnvironmentTransition, EnvironmentTransitionUnchanged)
	}
	if !observation.NonAuthorizing || !validDigest(observation.Digest) {
		t.Fatalf("observation must remain non-authorizing and digest-backed: %#v", observation)
	}
}

func TestObserveSelfImprovementPreservesRegression(t *testing.T) {
	environment := selfImprovementEnvironment("source-v1")
	observation := ObserveSelfImprovement(
		selfImprovementOrigin("before"), selfImprovementOrigin("after"),
		environment, environment,
		selfImprovementMetric(10, false, "metric-before"),
		selfImprovementMetric(7, false, "metric-after"),
	)
	if observation.Status != SelfImprovementStatusRegressed || observation.Reason != "METRIC_REGRESSED" {
		t.Fatalf("observation=%#v, want an explicit regression", observation)
	}
}

func TestObserveSelfImprovementRejectsChangedEnvironmentAndIncompleteOrigin(t *testing.T) {
	ready := selfImprovementEnvironment("source-v1")
	changed := selfImprovementEnvironment("source-v2")
	changedEnvironment := ObserveSelfImprovement(
		selfImprovementOrigin("before"), selfImprovementOrigin("after"),
		ready, changed,
		selfImprovementMetric(10, true, "metric-before"),
		selfImprovementMetric(7, true, "metric-after"),
	)
	if changedEnvironment.Status != SelfImprovementStatusUnknown || changedEnvironment.Reason != "ENVIRONMENT_CHANGED" {
		t.Fatalf("changed environment=%#v, want UNKNOWN", changedEnvironment)
	}

	incompleteOrigin := ObserveSelfImprovement(
		provenance.ObserveOriginChain(provenance.OriginChain{}), selfImprovementOrigin("after"),
		ready, ready,
		selfImprovementMetric(10, true, "metric-before"),
		selfImprovementMetric(7, true, "metric-after"),
	)
	if incompleteOrigin.Status != SelfImprovementStatusUnknown || incompleteOrigin.Reason != "ORIGIN_COMPARISON_UNKNOWN" {
		t.Fatalf("incomplete origin=%#v, want UNKNOWN", incompleteOrigin)
	}
}
