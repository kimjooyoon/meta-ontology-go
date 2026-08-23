package languageassurance

const (
	TransactionSchema = "gooo/language-assurance-transaction/v1"
	ReportSchema      = "gooo/language-assurance-report/v1"
	DenominatorID     = "gooo/language-assurance-denominator/v1"
	Producer          = "internal/meta/languageassurance.Evaluate"
	Consumer          = "language-assurance-gate"

	AssurancePartial      = "PARTIAL"
	CandidateAllowLimited = "ALLOW_LIMITED"
	CandidateBlock        = "BLOCK"
	CandidateFailClosed   = "FAIL_CLOSED"

	ReasonBoundaryClear       = "IMPLEMENTED_ASSURANCE_BOUNDARY_CLEAR"
	ReasonEvidenceUnknown     = "ASSURANCE_EVIDENCE_UNKNOWN"
	ReasonTopDecisionUnknown  = "ASSURANCE_TOP_DECISION_UNKNOWN"
	ReasonGovernanceViolation = "ASSURANCE_GOVERNANCE_VIOLATION"

	MetricSelfMinting       = "gooo.metric.governance.self-minting-paths.v1"
	MetricRoleConflict      = "gooo.metric.governance.role-conflict-paths.v1"
	MetricUnknownLaundering = "gooo.metric.epistemic.unknown-laundering.v1"
)

var denominatorV1 = []ObligationDefinition{
	obligation(MetricSelfMinting, PriorityP0, ClassGuardrail, ProofFoundation, "detect-self-minting-paths"),
	obligation(MetricRoleConflict, PriorityP0, ClassGuardrail, ProofCoherence, "detect-role-conflict-paths"),
	obligation(MetricUnknownLaundering, PriorityP0, ClassGuardrail, ProofRegression, "detect-unknown-laundering"),
	obligation("gooo.metric.evidence.exact-snapshot-binding.v1", PriorityP0, ClassDriver, ProofFoundation, "bind-exact-snapshot"),
	obligation("gooo.metric.evidence.raw-reconstruction.v1", PriorityP0, ClassDriver, ProofRegression, "reconstruct-raw-evidence"),
	obligation("gooo.metric.effects.write-set-exactness.v1", PriorityP0, ClassGuardrail, ProofRegression, "observe-exact-write-set"),
	obligation("gooo.metric.semantic.source-backed-authority.v1", PriorityP1, ClassDriver, ProofFoundation, "bind-source-backed-authority"),
	obligation("gooo.metric.semantic.candidate-leakage.v1", PriorityP1, ClassGuardrail, ProofCoherence, "detect-candidate-leakage"),
	obligation("gooo.metric.semantic.changed-surface-receipt-totality.v1", PriorityP1, ClassDriver, ProofCoherence, "totalize-changed-surface-receipts"),
	obligation("gooo.metric.operation.rollback-integrity.v1", PriorityP1, ClassGuardrail, ProofRegression, "verify-rollback-integrity"),
	obligation("gooo.metric.capability.vertical-slice-closure.v1", PriorityP2, ClassOutcome, ProofCoherence, "close-vertical-slice"),
	obligation("gooo.metric.ecosystem.external-conformance.v1", PriorityP2, ClassOutcome, ProofRegression, "verify-external-conformance"),
}

var operatingOperations = map[string]string{MetricSelfMinting: "detect-self-minting-paths", MetricRoleConflict: "detect-role-conflict-paths", MetricUnknownLaundering: "detect-unknown-laundering"}

var conflictPairs = []RolePair{{Left: RoleContractAuthor, Right: RoleEvaluatorAuthor}, {Left: RoleImplementer, Right: RolePromoter}, {Left: RoleEvaluatorAuthor, Right: RoleAuditor}, {Left: RolePolicyAdopter, Right: RolePromoter}, {Left: RoleAdapterAuthor, Right: RoleAuditor}}

var launderingOutputs = []Decision{DecisionPass, DecisionFixedPoint, DecisionAuthorized, DecisionAllow}

func Denominator() []ObligationDefinition {
	return append([]ObligationDefinition(nil), denominatorV1...)
}

func CanonicalMetaOperations() []MetaOperation {
	return []MetaOperation{
		{ID: "freeze-assurance-denominator", Activity: "FreezeAssuranceDenominator", ProofChoice: ProofCoherence},
		{ID: "observe-transaction-evidence", Activity: "ObserveTransactionEvidence", ProofChoice: ProofFoundation},
		{ID: "detect-self-minting-paths", Activity: "DetectSelfMintingPaths", ProofChoice: ProofFoundation},
		{ID: "detect-role-conflict-paths", Activity: "DetectRoleConflictPaths", ProofChoice: ProofCoherence},
		{ID: "detect-unknown-laundering", Activity: "DetectUnknownLaundering", ProofChoice: ProofRegression},
	}
}

func RoleConflictPairs() []RolePair {
	return append([]RolePair(nil), conflictPairs...)
}

func UnknownLaunderingOutputs() []Decision {
	return append([]Decision(nil), launderingOutputs...)
}

func obligation(metricID string, priority Priority, class IndicatorClass, proof ProofChoice, operation string) ObligationDefinition {
	return ObligationDefinition{MetricID: metricID, Priority: priority, Class: class, ProofChoice: proof, RequiredMetaOperation: operation}
}
