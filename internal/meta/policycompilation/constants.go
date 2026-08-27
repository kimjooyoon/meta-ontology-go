package policycompilation

const (
	SchemaVersion      = "gooo/meta-policy-compilation/v1"
	ArtifactSchema     = "gooo/meta-policy-compilation-artifact/v1"
	ReceiptSchema      = "gooo/meta-policy-compilation-receipt/v1"
	ClaimLedgerSchema  = "gooo/meta-policy-compilation-claims/v1"
	FixedDenominator   = 8
	ExpectedCaseCount  = 3
	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	DecisionUnknown    = "UNKNOWN"
	ClaimUnrecorded    = "UNRECORDED"
	ClaimOpen          = "OPEN"
	ClaimDischarged    = "DISCHARGED"
	ClaimRefuted       = "REFUTED"
	VerificationPass   = "PASS"
	VerificationFail   = "FAIL_CLOSED"
	ProofChoice        = "SEMANTIC_EQUIVALENCE_WITH_INDEPENDENT_EXECUTION"
	MetaOperation      = "COMPILE_GOOO_POLICY_TO_DECISION_KERNEL"
)

// RuleSpec is the source-independent meaning that the policy compiler expects
// from this experiment. The .gooo value program supplies the same fields; the
// equality check is what makes compilation a semantic binding rather than a
// text-preserving DSL copy.
type RuleSpec struct {
	Name          string
	Role          string
	MetaOperation string
	ProofChoice   string
	Stage         string
	Step          int
	Reason        string
	Claim         string
}

var ruleSpecs = []RuleSpec{
	{Name: "BindProducer", Role: "PRODUCER", MetaOperation: "bind-policy-source", ProofChoice: "source-ir-digest", Stage: "PRODUCE", Step: 1, Reason: "SOURCE_BOUND", Claim: "producer-bound"},
	{Name: "BindConsumer", Role: "CONSUMER", MetaOperation: "bind-compiled-artifact", ProofChoice: "artifact-source-equivalence", Stage: "CONSUME", Step: 2, Reason: "ARTIFACT_BOUND", Claim: "consumer-bound"},
	{Name: "CompileGeneratedJudge", Role: "META_OPERATION", MetaOperation: "compile-generated-judge", ProofChoice: "policy-semantic-preservation", Stage: "COMPILE", Step: 3, Reason: "JUDGE_COMPILED", Claim: "judge-compiled"},
	{Name: "RunIndependentVerifier", Role: "CONSUMER", MetaOperation: "run-independent-verifier", ProofChoice: "independent-reexecution", Stage: "VERIFY", Step: 4, Reason: "INDEPENDENT_VERIFIED", Claim: "independent-verified"},
	{Name: "SelectSemanticProof", Role: "PROOF", MetaOperation: "select-semantic-equivalence", ProofChoice: ProofChoice, Stage: "PROVE", Step: 5, Reason: "PROOF_SELECTED", Claim: "proof-selected"},
	{Name: "RecordClaimTransition", Role: "META_OPERATION", MetaOperation: "record-claim-transition", ProofChoice: "append-only-ledger", Stage: "LEDGER", Step: 6, Reason: "CLAIM_TRANSITION_RECORDED", Claim: "claim-persisted"},
	{Name: "ReduceDecision", Role: "META_OPERATION", MetaOperation: "reduce-decision", ProofChoice: "fail-closed-reconciliation", Stage: "REDUCE", Step: 7, Reason: "DECISION_REDUCED", Claim: "decision-reduced"},
	{Name: "BindLineage", Role: "CONSUMER", MetaOperation: "bind-lineage-receipt", ProofChoice: "lineage-receipt", Stage: "RECEIPT", Step: 8, Reason: "LINEAGE_BOUND", Claim: "lineage-bound"},
}

func expectedRule(name string) (RuleSpec, bool) {
	for _, spec := range ruleSpecs {
		if spec.Name == name {
			return spec, true
		}
	}
	return RuleSpec{}, false
}
