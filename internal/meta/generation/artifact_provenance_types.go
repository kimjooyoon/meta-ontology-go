package generation

const ArtifactProvenanceSchemaVersion = "gooo/meta-artifact-provenance/v1"

type ArtifactProvenanceDecision string

const (
	ArtifactProvenanceDecisionBound    ArtifactProvenanceDecision = "BOUND"
	ArtifactProvenanceDecisionUnknown  ArtifactProvenanceDecision = "UNKNOWN"
	ArtifactProvenanceDecisionRejected ArtifactProvenanceDecision = "REJECTED"
)

type ArtifactProvenanceReason string

const (
	ArtifactProvenanceReasonBound    ArtifactProvenanceReason = "ARTIFACT_PROVENANCE_BOUND"
	ArtifactProvenanceReasonUnproven ArtifactProvenanceReason = "ARTIFACT_PROVENANCE_UNPROVEN"
	ArtifactProvenanceReasonMismatch ArtifactProvenanceReason = "ARTIFACT_PROVENANCE_MISMATCH"
)

type ArtifactProvenanceIndicator struct {
	ID             string           `json:"id"`
	Route          TrilemmaRoute    `json:"route"`
	Verdict        IndicatorVerdict `json:"verdict"`
	EvidenceDigest string           `json:"evidence_digest"`
}

type ArtifactProvenanceSummary struct {
	Pass    int `json:"pass"`
	Fail    int `json:"fail"`
	Unknown int `json:"unknown"`
}

type ArtifactProvenance struct {
	SchemaVersion                 string                        `json:"schema_version"`
	BaseSHA                       string                        `json:"base_sha"`
	HeadSHA                       string                        `json:"head_sha"`
	PlanDigest                    string                        `json:"plan_digest"`
	ExecutionManifestDigest       string                        `json:"execution_manifest_digest"`
	ReceiptReportDigest           string                        `json:"receipt_report_digest"`
	IndicatorDecisionLedgerDigest string                        `json:"indicator_decision_ledger_digest"`
	IndicatorDecisionLedgerCount  int                           `json:"indicator_decision_ledger_count"`
	InputDigest                   string                        `json:"input_digest"`
	Decision                      ArtifactProvenanceDecision    `json:"decision"`
	Reason                        ArtifactProvenanceReason      `json:"reason"`
	Indicators                    []ArtifactProvenanceIndicator `json:"indicators"`
	Summary                       ArtifactProvenanceSummary     `json:"summary"`
	PromotionAuthorized           bool                          `json:"promotion_authorized"`
	EnvelopeDigest                string                        `json:"envelope_digest"`
	ReplayDigest                  string                        `json:"replay_digest"`
}
