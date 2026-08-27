package policycompilation

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCompilePolicyBindsEightMeaningfulObligations(t *testing.T) {
	source := policyFixture(t)
	policy, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Schema != SchemaVersion || policy.Denominator != FixedDenominator || len(policy.Rules) != FixedDenominator || policy.SourceDigest == "" || policy.SemanticDigest == "" {
		t.Fatalf("incomplete compiled policy: %#v", policy)
	}
	for index, rule := range policy.Rules {
		if rule.Step != index+1 || rule.Role == "" || rule.MetaOperation == "" || rule.ProofChoice == "" || rule.Stage == "" || rule.Reason == "" || rule.Claim == "" {
			t.Fatalf("incomplete rule %d: %#v", index, rule)
		}
	}
}

func TestCompileRejectsSemanticMetadataDrift(t *testing.T) {
	source := strings.Replace(policyFixture(t), "reason=SOURCE_BOUND", "reason=SOURCE_LIED", 1)
	if _, err := Compile([]byte(source)); err == nil {
		t.Fatal("semantic metadata drift was accepted")
	}
}

func TestIndependentEvaluatorHasPassFailClosedAndUnknownCases(t *testing.T) {
	policy, err := Compile(policyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	pass := Case{ID: "pass", Expected: DecisionPass, ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest, ObservedIndependentDigest: policy.SourceDigest}
	fail := pass
	fail.ID, fail.Expected, fail.ObservedSourceDigest = "fail", DecisionFailClosed, "sha256:drift"
	unknown := pass
	unknown.ID, unknown.Expected, unknown.ConsumerAvailable = "unknown", DecisionUnknown, false
	for _, test := range []struct {
		input Case
		want  string
	}{{pass, DecisionPass}, {fail, DecisionFailClosed}, {unknown, DecisionUnknown}} {
		if got := IndependentEvaluate(policy, test.input).Decision; got != test.want {
			t.Fatalf("case %q = %s, want %s", test.input.ID, got, test.want)
		}
	}
}

func TestReceiptCarriesFixedDenominatorAndAppendOnlyLedger(t *testing.T) {
	policy, err := Compile(policyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	cases := []Case{
		{ID: "pass", Expected: DecisionPass, ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest, ObservedIndependentDigest: policy.SourceDigest},
		{ID: "fail", Expected: DecisionFailClosed, ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: "sha256:drift", ObservedArtifactSourceDigest: policy.SourceDigest, ObservedIndependentDigest: policy.SourceDigest},
		{ID: "unknown", Expected: DecisionUnknown, ProducerAvailable: true, ConsumerAvailable: false, ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest, ObservedIndependentDigest: policy.SourceDigest},
	}
	judge := GenerateJudge(policy)
	judgeHash := DigestBytes(judge)
	artifact := PolicyArtifact{Schema: ArtifactSchema, Policy: policy, GeneratedJudgeHash: judgeHash}
	generated := make([]DecisionResult, 0, len(cases))
	independent := make([]DecisionResult, 0, len(cases))
	for _, input := range cases {
		result := IndependentEvaluate(policy, input)
		generated = append(generated, result)
		independent = append(independent, result)
	}
	receipt, err := BuildReceipt(policy, artifact, judgeHash, cases, generated, independent)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Claims.EventCount != ExpectedCaseCount*FixedDenominator*2 || receipt.Summary.PassCount != 1 || receipt.Summary.FailClosedCount != 1 || receipt.Summary.UnknownCount != 1 {
		t.Fatalf("unexpected receipt summary: %#v", receipt)
	}
	if err := VerifyReceipt(receipt, policy, artifact, judgeHash, cases); err != nil {
		t.Fatal(err)
	}
	mutated := receipt
	mutated.Claims.Events = append([]ClaimTransition(nil), receipt.Claims.Events...)
	mutated.Claims.Events[0].Reason = "edited"
	if err := VerifyReceipt(mutated, policy, artifact, judgeHash, cases); err == nil {
		t.Fatal("edited claim event was accepted")
	}
}

func policyFixture(t *testing.T) []byte {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "../../../examples/meta-policy-compilation/policy.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
