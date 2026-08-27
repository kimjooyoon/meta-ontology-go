package minimalcausalexplanation

import "testing"

func TestEvaluateRequiresSourceBackedObservations(t *testing.T) {
	source := []byte(`package minimalcausalexplanation
namespace minimalcausalexplanation
entity SourceParsedEvidence id "gooo://meta/mce/evidence/source-parsed"
entity SemanticIRBoundEvidence id "gooo://meta/mce/evidence/semantic-ir-bound"
entity CompilerReceiptProvenEvidence id "gooo://meta/mce/evidence/compiler-receipt-proven"
entity AuditNoiseEvidence id "gooo://meta/mce/evidence/audit-noise"
entity DecisionOutput id "gooo://meta/mce/output/decision"
entity PriorClaimState id "gooo://meta/mce/claim/prior-state"
entity PreservedClaims id "gooo://meta/mce/output/preserved-claims"
activity BindSource(SourceParsedEvidence) -> SemanticIRBoundEvidence computes "mce.operation:bind-source|source-reader|causal-evaluator|FOUNDATION:v1"
activity BindCompilerReceipt(SemanticIRBoundEvidence) -> CompilerReceiptProvenEvidence computes "mce.operation:bind-compiler-receipt|receipt-reader|causal-evaluator|FOUNDATION:v1"
activity EvaluatePredicate(CompilerReceiptProvenEvidence) -> DecisionOutput computes "mce.operation:judge-predicate|causal-evaluator|path-checker|COHERENCE;mce.predicate:PASS_IF:source-parsed+semantic-ir-bound+compiler-receipt-proven:v1"
activity ObserveAudit(CompilerReceiptProvenEvidence) -> AuditNoiseEvidence computes "mce.operation:observe-audit|audit-reader|audit-archive|FOUNDATION:v1"
activity OpenClaimLedger(DecisionOutput) -> PriorClaimState computes "mce.operation:open-claims|claim-ledger|path-checker|REGRESSION;mce.claim-state:OPEN:v1"
activity PreserveClaims(PriorClaimState) -> PreservedClaims computes "mce.operation:preserve-claims|claim-ledger|ci-judge|REGRESSION;mce.program:gooo://meta/mce/evaluator|gooo://meta/mce/judge:v1;mce.indicators:12:v1;mce.decision-output:VALUE_WITNESS_PROVEN:v1;mce.claims:source-bound+graph-predicate-reconstructed+subset-minimal+cardinality-minimum+counterfactual-difference+read-only-preserved:v1"
`)
	receipt, err := Evaluate("fixture.gooo", source, []byte(`{"schema":"gooo.language.value-witness/v2","decision":"VALUE_WITNESS_PROVEN","reason":"VALUE_WITNESS_EXACT","resolution":"CORE_IR_ACTIVITY_VALUE_PROGRAM","head_sha":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","source_path":"examples/language-value-witness/main.gooo","source_digest":"sha256:source","semantic_fingerprint":"semantic","core_ir_fingerprint":"core"}`), []byte(`{"schema":"gooo/meta-minimal-causal-explanation-repository/v1","status":"","workspace_writes":false,"promotion_authorized":false}`), []byte(`{"schema":"gooo/meta-minimal-causal-explanation-repository/v1","status":"","workspace_writes":false,"promotion_authorized":false}`), []byte(`{"schema":"gooo/meta-minimal-causal-explanation-independence/v1","producer_package_import_count":0,"producer_package_import_total":1}`), "example/repository", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != StatusPass || receipt.Summary.SubsetMinimalNumerator != 1 || receipt.Summary.SubsetMinimalDenominator != 2 || receipt.Summary.CardinalityMinimumNumerator != 1 || receipt.Summary.CardinalityMinimumDenominator != 2 {
		t.Fatalf("unexpected path result: %+v", receipt.Summary)
	}
	if receipt.Summary.CounterfactualExecutions != 7 || receipt.Summary.ChangedCounterfactuals != 6 || len(receipt.ClaimTransitions) != 12 || receipt.Preservation.PreservedTotal != 6 {
		t.Fatalf("unexpected receipt result: %+v", receipt.Summary)
	}
}
