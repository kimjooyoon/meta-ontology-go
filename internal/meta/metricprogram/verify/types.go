package verify

const (
	strategySchema             = "gooo/metric-meta-strategy/v1"
	strategyVerificationSchema = "gooo/metric-meta-strategy-verification/v1"
	programSchema              = "gooo/metric-meta-program/v1"
	reportSchema               = "gooo/metric-meta-program-verification/v1"
	programSourceFilename      = "program.gooo"
)

type rootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	ReadmeRequirement     string `json:"readme_requirement"`
}

type strategyBinding struct {
	IndicatorID    string `json:"indicator_id"`
	Family         string `json:"family"`
	Trilemma       string `json:"trilemma"`
	MetaOperation  string `json:"meta_operation"`
	Expected       string `json:"expected"`
	Actual         string `json:"actual"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type strategyCandidate struct {
	ProofChoice      string   `json:"proof_choice"`
	IndicatorIDs     []string `json:"indicator_ids"`
	MetaOperations   []string `json:"meta_operations"`
	IndicatorCount   int      `json:"indicator_count"`
	UnsatisfiedCount int      `json:"unsatisfied_count"`
	Admissible       bool     `json:"admissible"`
	EvidenceDigest   string   `json:"evidence_digest"`
}

type strategySelection struct {
	ProofChoice          string   `json:"proof_choice"`
	Decision             string   `json:"decision"`
	MetaOperation        string   `json:"meta_operation"`
	Reason               string   `json:"reason"`
	CandidateDigest      string   `json:"candidate_digest"`
	SourceMetaOperations []string `json:"source_meta_operations"`
}
