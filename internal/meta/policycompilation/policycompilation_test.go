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

func TestCompileAcceptsSemanticMetadataIntervention(t *testing.T) {
	source := strings.Replace(policyFixture(t), "reason=SOURCE_BOUND", "reason=SOURCE_LIED", 1)
	changed, err := Compile([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	original, err := Compile(policyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if changed.SemanticDigest == original.SemanticDigest || changed.Rules[0].Reason != "SOURCE_LIED" {
		t.Fatalf("semantic intervention did not change source-derived policy: %#v", changed)
	}
}

func TestIndependentEvaluatorHasPassFailClosedAndUnknownCases(t *testing.T) {
	policy, err := Compile(policyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	judgeHash := DigestBytes(GenerateJudge(policy))
	pass := Case{ID: "pass", ValidatorExpectation: DecisionPass, EvidenceClass: EvidenceSyntheticFixture, Provenance: "test synthetic fixture", ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest, ObservedGeneratedJudgeDigest: judgeHash, ObservedIndependentDigest: policy.SemanticDigest}
	fail := pass
	fail.ID, fail.ValidatorExpectation, fail.ObservedSourceDigest = "fail", DecisionFailClosed, "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	unknown := pass
	unknown.ID, unknown.ValidatorExpectation, unknown.ConsumerAvailable = "unknown", DecisionUnknown, false
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
	judge := GenerateJudge(policy)
	judgeHash := DigestBytes(judge)
	artifact := PolicyArtifact{Schema: ArtifactSchema, Policy: policy, GeneratedJudgeHash: judgeHash}
	base := Case{ValidatorExpectation: DecisionPass, EvidenceClass: EvidenceSyntheticFixture, Provenance: "test synthetic fixture", ProducerAvailable: true, ConsumerAvailable: true, ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest, ObservedGeneratedJudgeDigest: judgeHash, ObservedIndependentDigest: policy.SemanticDigest}
	cases := []Case{
		func() Case { value := base; value.ID = "pass"; return value }(),
		func() Case {
			value := base
			value.ID, value.ValidatorExpectation, value.ObservedSourceDigest = "fail", DecisionFailClosed, "sha256:0000000000000000000000000000000000000000000000000000000000000000"
			return value
		}(),
		func() Case {
			value := base
			value.ID, value.ValidatorExpectation, value.ConsumerAvailable = "unknown", DecisionUnknown, false
			return value
		}(),
		func() Case {
			value := base
			value.ID, value.ValidatorExpectation, value.ObservedSourceDigest = "malformed", DecisionUnknown, "sha256:not-a-valid-content-digest"
			return value
		}(),
	}
	generated := make([]DecisionResult, 0, len(cases))
	independent := make([]DecisionResult, 0, len(cases))
	for _, input := range cases {
		result := IndependentEvaluate(policy, input)
		generated = append(generated, result)
		independent = append(independent, result)
	}
	receipt, err := BuildReceipt(policy, artifact, judgeHash, cases, generated, independent, WriteSetObservation{RepositoryBeforeDigest: "sha256:unchanged", RepositoryAfterDigest: "sha256:unchanged", GeneratedRootClass: "RUNNER_TEMP_ONLY", GeneratedFiles: []string{"artifact.json", "generated-results.json", "independent-results.json", "judge.go", "policy.json", "receipt.json"}})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Claims.EventCount != ExpectedCaseCount*ClaimPredicateCount*2 || receipt.Summary.PassCount != 1 || receipt.Summary.FailClosedCount != 1 || receipt.Summary.UnknownCount != 2 {
		t.Fatalf("unexpected receipt summary: %#v", receipt)
	}
	if receipt.Summary.SourceAllEquivalent != ExpectedCaseCount || receipt.Summary.ValidatorExpectationsConfirmed != ExpectedCaseCount || len(receipt.Cases) != ExpectedCaseCount || !receipt.Cases[0].AllDecisionsEquivalent {
		t.Fatalf("source lineage is incomplete: %#v", receipt.Summary)
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
