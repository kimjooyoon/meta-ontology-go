package policycompilation

const (
	SchemaVersion     = "gooo/meta-policy-compilation/v3"
	ArtifactSchema    = "gooo/meta-policy-compilation-artifact/v3"
	ReceiptSchema     = "gooo/meta-policy-compilation-receipt/v3"
	ClaimLedgerSchema = "gooo/meta-policy-compilation-claims/v3"
	ReductionSchema   = "decision-reduction:v2"

	FixedDenominator    = 8
	ReductionRuleCount  = 8
	ExpectedCaseCount   = 3
	ClaimPredicateCount = 8

	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	DecisionUnknown    = "UNKNOWN"

	ClaimUnrecorded = "UNRECORDED"
	ClaimOpen       = "OPEN"
	ClaimDischarged = "DISCHARGED"
	ClaimRefuted    = "REFUTED"

	VerificationPass = "PASS"
	VerificationFail = "FAIL_CLOSED"

	EvidenceSyntheticFixture = "SYNTHETIC_FIXTURE"
	EvidenceCurrent          = "CURRENT_EVIDENCE"
	SubjectUnresolved        = "UNRESOLVED"
	SubjectResolved          = "RESOLVED"
)

const (
	ConditionEvidenceUnavailable     = "EVIDENCE_UNAVAILABLE"
	ConditionDigestUnavailable       = "DIGEST_UNAVAILABLE"
	ConditionMalformedDigest         = "MALFORMED_DIGEST"
	ConditionSourceMismatch          = "SOURCE_DIGEST_MISMATCH"
	ConditionArtifactMismatch        = "ARTIFACT_SOURCE_MISMATCH"
	ConditionIndependentMismatch     = "INDEPENDENT_SOURCE_MISMATCH"
	ConditionUnrecognizedTopDecision = "UNRECOGNIZED_TOP_LEVEL_DECISION"
	ConditionSemanticEquivalence     = "SEMANTIC_EQUIVALENCE"
	UnknownClassEvidenceUnavailable  = "EVIDENCE_UNAVAILABLE"
	UnknownClassDigestUnavailable    = "DIGEST_UNAVAILABLE"
	UnknownClassMalformedInput       = "MALFORMED_INPUT"
	NextCollectPolicyEvidence        = "collect-policy-evidence"
	NextRepairDigestEvidence         = "repair-digest-evidence"
	UnknownBlockedProducer           = "producer-evidence"
	UnknownBlockedConsumer           = "consumer-evidence"
	UnknownBlockedDigest             = "well-formed-sha256-digests"
)

const (
	ClaimPredicateSourceBound        = "source-bound"
	ClaimPredicateArtifactBound      = "artifact-digest-bound"
	ClaimPredicateGeneratedExecution = "generated-execution"
	ClaimPredicateIndependentReplay  = "independent-replay"
	ClaimPredicateProofSelection     = "proof-selection"
	ClaimPredicateLedgerChain        = "ledger-chain"
	ClaimPredicateDecisionReduction  = "decision-reduction"
	ClaimPredicateLineageSeal        = "lineage-seal"
)
