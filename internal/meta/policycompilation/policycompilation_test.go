package policycompilation

import (
	"os"
	"testing"
)

func canonicalPolicyForTest(t *testing.T) CompiledPolicy {
	t.Helper()
	source, err := os.ReadFile("../../../examples/meta-policy-compilation/policy.gooo")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Denominator != FixedDenominator || len(policy.Rules) != FixedDenominator || len(policy.Reduction.Rules) != ReductionRuleCount {
		t.Fatalf("source did not produce fixed policy: %#v", policy)
	}
	return policy
}

func TestGoooSourceOwnsEightPredicatesAndReduction(t *testing.T) {
	policy := canonicalPolicyForTest(t)
	if err := validateClaimPredicates(policy.Rules); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, rule := range policy.Rules {
		if seen[rule.Claim] {
			t.Fatalf("duplicate predicate %q", rule.Claim)
		}
		seen[rule.Claim] = true
	}
	if len(seen) != ClaimPredicateCount {
		t.Fatalf("predicate denominator got %d want %d", len(seen), ClaimPredicateCount)
	}
}

func TestValidContradictionPrecedesUnknownAndMalformedPreservesUnknown(t *testing.T) {
	policy := canonicalPolicyForTest(t)
	judgeHash := DigestBytes(GenerateJudge(policy))
	valid := Case{ID: "valid", ProducerAvailable: false, ConsumerAvailable: false, ObservedSourceDigest: DigestBytes([]byte("contradiction")), ObservedArtifactSourceDigest: policy.SourceDigest, ObservedGeneratedJudgeDigest: judgeHash, ObservedIndependentDigest: policy.SemanticDigest}
	contradiction := EvaluateSourcePolicy(policy, valid)
	if contradiction.Decision != DecisionFailClosed || contradiction.Reason != ConditionSourceMismatch {
		t.Fatalf("valid contradiction was not prioritized: %#v", contradiction)
	}
	malformed := valid
	malformed.ID = "malformed"
	malformed.ProducerAvailable, malformed.ConsumerAvailable = true, true
	malformed.ObservedSourceDigest = "sha256:not-a-digest"
	unknown := EvaluateSourcePolicy(policy, malformed)
	if unknown.Decision != DecisionUnknown || unknown.UnknownClass != UnknownClassMalformedInput || unknown.NextOperation == "" || len(unknown.BlockedBy) == 0 || unknown.Stage == "" || unknown.Step == 0 || unknown.Reason == "" {
		t.Fatalf("UNKNOWN did not preserve six fields: %#v", unknown)
	}
}

func TestUnrecognizedUpperDecisionFailsClosed(t *testing.T) {
	policy := canonicalPolicyForTest(t)
	input := Case{ID: "upper", ProducerAvailable: true, ConsumerAvailable: true, UpperDecision: "FIXED_POINT", ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest, ObservedGeneratedJudgeDigest: DigestBytes(GenerateJudge(policy)), ObservedIndependentDigest: policy.SemanticDigest}
	result := EvaluateSourcePolicy(policy, input)
	if result.Decision != DecisionFailClosed || result.Reason != "FEEDBACK_COVERAGE_DECISION_UNKNOWN" {
		t.Fatalf("unrecognized upper decision was accepted: %#v", result)
	}
}
