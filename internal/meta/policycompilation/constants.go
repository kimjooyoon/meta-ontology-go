package policycompilation

const (
	SchemaVersion      = "gooo/meta-policy-compilation/v2"
	ArtifactSchema     = "gooo/meta-policy-compilation-artifact/v2"
	ReceiptSchema      = "gooo/meta-policy-compilation-receipt/v2"
	ClaimLedgerSchema  = "gooo/meta-policy-compilation-claims/v2"
	ReductionSchema    = "decision-reduction:v1"
	FixedDenominator   = 8
	ReductionRuleCount = 6
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

const (
	ConditionEvidenceUnavailable = "EVIDENCE_UNAVAILABLE"
	ConditionDigestUnavailable   = "DIGEST_UNAVAILABLE"
	ConditionSourceMismatch      = "SOURCE_DIGEST_MISMATCH"
	ConditionArtifactMismatch    = "ARTIFACT_SOURCE_MISMATCH"
	ConditionIndependentMismatch = "INDEPENDENT_SOURCE_MISMATCH"
	ConditionSemanticEquivalence = "SEMANTIC_EQUIVALENCE"
	EvidenceSyntheticFixture     = "SYNTHETIC_FIXTURE"
	EvidenceCurrent              = "CURRENT_EVIDENCE"
	SubjectUnresolved            = "UNRESOLVED"
	SubjectResolved              = "RESOLVED"
)
